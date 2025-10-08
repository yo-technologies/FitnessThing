package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/utils"

	openai_client "fitness-trainer/internal/clients/openai"

	"github.com/openai/openai-go"
	"github.com/opentracing/opentracing-go"
)

type agentToolHandler func(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error)

type agentTool struct {
	name       string
	definition openai.ChatCompletionToolParam
	handler    agentToolHandler
}

var _ openai_client.ChatClient = (*openai_client.Client)(nil)

type service interface {
	// Workout methods
	GetWorkout(ctx context.Context, userID, workoutID domain.ID) (dto.WorkoutDetailsDTO, error)
	GetWorkouts(ctx context.Context, userID domain.ID, limit, offset int) ([]dto.WorkoutDTO, error)

	// Exercise methods
	GetExercises(ctx context.Context, muscleGroups, excludedExercises []domain.ID) ([]domain.Exercise, error)
	GetExerciseByID(ctx context.Context, id domain.ID) (domain.Exercise, error)
	GetExerciseHistory(ctx context.Context, userID, exerciseID domain.ID, offset, limit int) ([]dto.ExerciseLogDTO, error)

	// Exercise log methods
	GetExerciseLogByID(ctx context.Context, id domain.ID) (domain.ExerciseLog, error)
	LogExercise(ctx context.Context, userID, workoutID, exerciseID domain.ID) (domain.ExerciseLog, error)
	DeleteExerciseLog(ctx context.Context, userID, workoutID, exerciseLogID domain.ID) error
	ReplaceExpectedSets(ctx context.Context, userID, workoutID, exerciseLogID domain.ID, sets []dto.ExpectedSetInput) error

	// Set log methods
	UpdateSetLog(ctx context.Context, userID, workoutID, exerciseLogID, setLogID domain.ID, setlogDTO dto.UpdateSetLogDTO) (domain.ExerciseSetLog, error)

	// Muscle groups
	GetMuscleGroups(ctx context.Context) ([]dto.MuscleGroupDTO, error)
}

type Tools struct {
	chatTools     map[string]agentTool
	chatToolsOnce sync.Once

	service service
}

func New(
	service service,
) *Tools {
	return &Tools{
		service:   service,
		chatTools: make(map[string]agentTool),
	}
}

func (t *Tools) ChatAgentToolDefinitions() []openai.ChatCompletionToolParam {
	return t.chatToolDefinitions()
}

func (t *Tools) ExecuteChatAgentTool(ctx context.Context, ctxData domain.AgentChatContext, name string, arguments string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.ExecuteChatAgentTool")
	defer span.Finish()

	return t.executeTool(ctx, ctxData, name, arguments)
}

func (t *Tools) NewChatCompletionParams(messages []openai.ChatCompletionMessageParamUnion, tools []openai.ChatCompletionToolParam, model string, stream bool) openai.ChatCompletionNewParams {
	return t.newChatCompletionParams(messages, tools, model, stream)
}

func (t *Tools) ensureChatTools() {
	t.chatToolsOnce.Do(func() {
		t.chatTools = map[string]agentTool{}

		for _, tool := range []agentTool{
			t.newListMuscleGroupsTool(),
			t.newListExercisesTool(),
			t.newGetWorkoutHistoryTool(),
			t.newGetExerciseHistoryTool(),
			t.newGetWorkoutPlanTool(),
			t.newAddExercisesToWorkoutTool(),
			t.newRemoveExerciseLogsTool(),
			t.newReplaceExpectedSetsTool(),
			t.newUpdateSetLogTool(),
		} {
			if tool.name == "" {
				panic("agent tool definition missing name")
			}
			t.chatTools[tool.name] = tool
		}
	})
}

func (t *Tools) chatToolDefinitions() []openai.ChatCompletionToolParam {
	t.ensureChatTools()

	defs := make([]openai.ChatCompletionToolParam, 0, len(t.chatTools))
	for _, tool := range t.chatTools {
		defs = append(defs, tool.definition)
	}

	return defs
}

func (t *Tools) executeTool(ctx context.Context, chatCtx domain.AgentChatContext, name string, arguments string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.executeTool")
	defer span.Finish()

	t.ensureChatTools()

	tool, ok := t.chatTools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	// Try to split multiple JSON objects if they were concatenated
	argsList := utils.SplitMultipleJSONObjects(arguments)

	if len(argsList) == 1 {
		// Single call
		result, err := tool.handler(ctx, chatCtx, json.RawMessage(argsList[0]))
		if err != nil {
			return "", err
		}

		return result, nil
	}

	// Multiple calls - execute them sequentially and combine results
	type callResult struct {
		Index  int    `json:"index"`
		Result string `json:"result"`
		Error  string `json:"error,omitempty"`
	}

	combinedResults := struct {
		CallCount int          `json:"call_count"`
		Results   []callResult `json:"results"`
	}{
		CallCount: len(argsList),
		Results:   make([]callResult, 0, len(argsList)),
	}

	for i, args := range argsList {
		result, err := tool.handler(ctx, chatCtx, json.RawMessage(args))
		callRes := callResult{
			Index:  i,
			Result: result,
		}
		if err != nil {
			callRes.Error = err.Error()
		}
		combinedResults.Results = append(combinedResults.Results, callRes)
	}

	resultJSON, err := json.Marshal(combinedResults)
	if err != nil {
		return "", fmt.Errorf("failed to marshal combined results: %w", err)
	}

	return string(resultJSON), nil
}

func (t *Tools) newChatCompletionParams(messages []openai.ChatCompletionMessageParamUnion, tools []openai.ChatCompletionToolParam, model string, stream bool) openai.ChatCompletionNewParams {
	params := openai.ChatCompletionNewParams{
		Model:    openai.String(model),
		Messages: openai.F(messages),
	}

	if len(tools) > 0 {
		params.Tools = openai.F(tools)
		params.ToolChoice = openai.F[openai.ChatCompletionToolChoiceOptionUnionParam](openai.ChatCompletionToolChoiceOptionAuto(openai.ChatCompletionToolChoiceOptionAutoAuto))
	}

	if stream {
		params.StreamOptions = openai.F(openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)})
	}

	return params
}
