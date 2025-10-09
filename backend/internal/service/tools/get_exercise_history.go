package tools

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/opentracing/opentracing-go"
)

func (t *Tools) newGetExerciseHistoryTool() agentTool {
	schema := shared.FunctionParameters{
		"type":     "object",
		"required": []string{"exercise_id", "logs_limit"},
		"properties": map[string]any{
			"exercise_id": map[string]any{
				"type":        "string",
				"description": "UUID of the exercise to get history for.",
			},
			"logs_limit": map[string]any{
				"type":        "integer",
				"minimum":     1,
				"maximum":     20,
				"description": "Maximum number of recent logs to return. Defaults to 10.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "get_exercise_history",
		definition: openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "get_exercise_history",
			Description: openai.String("Return recent performance history for a specific exercise."),
			Parameters:  schema,
			Strict:      openai.Bool(true),
		}),
		handler: t.getExerciseHistoryHandler,
	}
}

func (t *Tools) getExerciseHistoryHandler(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.get_exercise_history")
	defer span.Finish()

	if chatCtx.UserID == (domain.ID{}) {
		return "", fmt.Errorf("user context is required for get_exercise_history")
	}

	args := getExerciseHistoryArgs{LogsLimit: new(int)}
	*args.LogsLimit = 10
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("invalid arguments for get_exercise_history: %w", err)
		}
		if args.LogsLimit == nil || *args.LogsLimit <= 0 {
			*args.LogsLimit = 10
		}
	}

	exerciseID, err := domain.ParseID(args.ExerciseID)
	if err != nil {
		return "", fmt.Errorf("invalid exercise_id: %w", err)
	}

	logs, err := t.service.GetExerciseHistory(ctx, chatCtx.UserID, exerciseID, 0, *args.LogsLimit)
	if err != nil {
		return "", fmt.Errorf("failed to get exercise history: %w", err)
	}

	rawResp, err := json.Marshal(exerciseLogHistoryPayloadFromDomain(logs))
	if err != nil {
		return "", fmt.Errorf("failed to marshal exercise history: %w", err)
	}

	return string(rawResp), nil
}
