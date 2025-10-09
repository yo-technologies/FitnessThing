package tools

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/utils"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/opentracing/opentracing-go"
)

func (t *Tools) newUpdateSetLogTool() agentTool {
	schema := shared.FunctionParameters{
		"type":     "object",
		"required": []string{"exercise_log_id", "set_log_id", "reps", "weight"},
		"properties": map[string]any{
			"exercise_log_id": map[string]any{
				"type":        "string",
				"description": "UUID of the exercise log that holds the set log.",
			},
			"set_log_id": map[string]any{
				"type":        "string",
				"description": "UUID of the set log to update.",
			},
			"reps": map[string]any{
				"type":        "integer",
				"description": "New repetitions count for the set.",
			},
			"weight": map[string]any{
				"type":        "number",
				"description": "New weight value for the set in the workout's unit.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "update_set_log",
		definition: openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "update_set_log",
			Description: openai.String("Update actual performance data (reps or weight) for a specific logged set."),
			Parameters:  schema,
			Strict:      openai.Bool(true),
		}),
		handler: t.updateSetLogHandler,
	}
}

func (t *Tools) updateSetLogHandler(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.update_set_log")
	defer span.Finish()

	if !chatCtx.WorkoutID.IsValid {
		return "", fmt.Errorf("workout context is required for update_set_log")
	}

	var args updateSetLogArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", fmt.Errorf("invalid arguments for update_set_log: %w", err)
	}

	exerciseLogID, err := domain.ParseID(args.ExerciseLogID)
	if err != nil {
		return "", fmt.Errorf("invalid exercise_log_id: %w", err)
	}

	setLogID, err := domain.ParseID(args.SetLogID)
	if err != nil {
		return "", fmt.Errorf("invalid set_log_id: %w", err)
	}

	dtoInput := dto.UpdateSetLogDTO{}
	if args.Reps != nil {
		dtoInput.Reps = utils.NewNullable(*args.Reps, true)
	}
	if args.Weight != nil {
		dtoInput.Weight = utils.NewNullable(float32(*args.Weight), true)
	}

	if !dtoInput.Reps.IsValid && !dtoInput.Weight.IsValid {
		return "", fmt.Errorf("at least one of reps or weight must be provided")
	}

	if _, err := t.service.UpdateSetLog(ctx, chatCtx.UserID, chatCtx.WorkoutID.V, exerciseLogID, setLogID, dtoInput); err != nil {
		return "", fmt.Errorf("failed to update set log: %w", err)
	}

	return `{"status":"success"}`, nil
}
