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

func (t *Tools) newRemoveExerciseLogsTool() agentTool {
	schema := shared.FunctionParameters{
		"type":     "object",
		"required": []string{"exercise_log_ids"},
		"properties": map[string]any{
			"exercise_log_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string", "description": "UUID of exercise log to remove"},
				"minItems":    1,
				"description": "List of exercise log UUIDs to remove from the workout.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "remove_exercise_logs_from_workout",
		definition: openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "remove_exercise_logs_from_workout",
			Description: openai.String("Remove one or more exercise logs from the current workout."),
			Parameters:  schema,
			Strict:      openai.Bool(true),
		}),
		handler: t.removeExerciseLogsHandler,
	}
}

func (t *Tools) removeExerciseLogsHandler(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.remove_exercise_logs_from_workout")
	defer span.Finish()

	if chatCtx.UserID == (domain.ID{}) {
		return "", fmt.Errorf("user context is required for remove_exercise_logs_from_workout")
	}

	if !chatCtx.WorkoutID.IsValid {
		return "", fmt.Errorf("workout context is required for remove_exercise_logs_from_workout")
	}

	var args removeExerciseLogsArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", fmt.Errorf("invalid arguments for remove_exercise_logs_from_workout: %w", err)
	}
	if len(args.ExerciseLogIDs) == 0 {
		return "", fmt.Errorf("exercise_log_ids must contain at least one id")
	}

	uniqueLogIDs := make([]domain.ID, 0, len(args.ExerciseLogIDs))
	seen := make(map[domain.ID]struct{}, len(args.ExerciseLogIDs))
	for idx, idStr := range args.ExerciseLogIDs {
		if idStr == "" {
			return "", fmt.Errorf("exercise_log_ids[%d] must not be empty", idx)
		}

		logID, err := domain.ParseID(idStr)
		if err != nil {
			return "", fmt.Errorf("invalid exercise_log_id at index %d: %w", idx, err)
		}

		if _, exists := seen[logID]; exists {
			continue
		}
		seen[logID] = struct{}{}
		uniqueLogIDs = append(uniqueLogIDs, logID)
	}

	if len(uniqueLogIDs) == 0 {
		return "", fmt.Errorf("no valid exercise_log_ids provided")
	}

	// Check that all logs belong to the current workout
	for _, logID := range uniqueLogIDs {
		exerciseLog, err := t.service.GetExerciseLogByID(ctx, logID)
		if err != nil {
			return "", fmt.Errorf("failed to load exercise log %s: %w", logID.String(), err)
		}

		if exerciseLog.WorkoutID != chatCtx.WorkoutID.V {
			return "", fmt.Errorf("exercise log %s does not belong to the current workout", logID.String())
		}
	}

	for _, logID := range uniqueLogIDs {
		if err := t.service.DeleteExerciseLog(ctx, chatCtx.UserID, chatCtx.WorkoutID.V, logID); err != nil {
			return "", fmt.Errorf("failed to delete exercise log %s: %w", logID.String(), err)
		}
	}

	return string(`{"success": true}`), nil
}
