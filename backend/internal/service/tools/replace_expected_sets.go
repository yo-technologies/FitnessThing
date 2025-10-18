package tools

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fmt"
	"time"
	"github.com/opentracing/opentracing-go"
)

func (t *Tools) newReplaceExpectedSetsTool() agentTool {
	schema := map[string]any{
		"type":     "object",
		"required": []string{"exercise_log_id", "sets"},
		"properties": map[string]any{
			"exercise_log_id": map[string]any{
				"type":        "string",
				"description": "UUID of the exercise log to update.",
			},
			"sets": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":     "object",
					"required": []string{"set_type", "reps", "weight", "time_seconds"},
					"properties": map[string]any{
						"set_type": map[string]any{
							"type": "string",
							"enum": []string{"reps", "weight", "time"},
						},
						"reps": map[string]any{
							"type":        "integer",
							"description": "Target repetitions for the set.",
						},
						"weight": map[string]any{
							"type":        "number",
							"description": "Target weight in the workout's unit.",
						},
						"time_seconds": map[string]any{
							"type":        "integer",
							"description": "Target duration in seconds for time-based sets.",
						},
					},
					"additionalProperties": false,
				},
				"minItems": 1,
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name:    "replace_expected_sets",
		desc:    "Replace the planned sets for an exercise log in the current workout.",
		params:  schema,
		handler: t.replaceExpectedSetsHandler,
	}
}

func (t *Tools) replaceExpectedSetsHandler(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.replace_expected_sets")
	defer span.Finish()

	if !chatCtx.WorkoutID.IsValid {
		return "", fmt.Errorf("workout context is required for replace_expected_sets")
	}

	var args replaceExpectedSetsArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", fmt.Errorf("invalid arguments for replace_expected_sets: %w", err)
	}

	if args.ExerciseLogID == "" {
		return "", fmt.Errorf("exercise_log_id is required")
	}

	exerciseLogID, err := domain.ParseID(args.ExerciseLogID)
	if err != nil {
		return "", fmt.Errorf("invalid exercise_log_id: %w", err)
	}

	inputs := make([]dto.ExpectedSetInput, 0, len(args.Sets))
	for idx, set := range args.Sets {
		setType, err := domain.NewSetType(set.SetType)
		if err != nil {
			return "", fmt.Errorf("invalid set_type at index %d: %w", idx, err)
		}

		reps := 0
		if set.Reps != nil {
			reps = *set.Reps
		}

		weight := float32(0)
		if set.Weight != nil {
			weight = float32(*set.Weight)
		}

		var duration time.Duration
		if set.TimeSeconds != nil {
			duration = time.Duration(*set.TimeSeconds) * time.Second
		}

		inputs = append(inputs, dto.ExpectedSetInput{
			SetType: setType,
			Reps:    reps,
			Weight:  weight,
			Time:    duration,
		})
	}

	if err := t.service.ReplaceExpectedSets(ctx, chatCtx.UserID, chatCtx.WorkoutID.V, exerciseLogID, inputs); err != nil {
		return "", fmt.Errorf("failed to replace expected sets: %w", err)
	}

	return `{"status":"success"}`, nil
}
