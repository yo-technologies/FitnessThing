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

func (t *Tools) newAddExercisesToWorkoutTool() agentTool {
	schema := shared.FunctionParameters{
		"type":     "object",
		"required": []string{"exercise_ids"},
		"properties": map[string]any{
			"exercise_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string", "description": "UUID of exercise to add"},
				"minItems":    1,
				"description": "List of exercise UUIDs to add to the workout.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "add_exercises_to_workout",
		definition: openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "add_exercises_to_workout",
			Description: openai.String("Add one or more exercises to the current workout. Creates exercise logs for each exercise."),
			Parameters:  schema,
			Strict:      openai.Bool(true),
		}),
		handler: t.addExercisesToWorkoutHandler,
	}
}

func (t *Tools) addExercisesToWorkoutHandler(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.add_exercises_to_workout")
	defer span.Finish()

	if !chatCtx.WorkoutID.IsValid {
		return "", fmt.Errorf("workout context is required for add_exercises_to_workout")
	}

	var args addExercisesToWorkoutArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", fmt.Errorf("invalid arguments for add_exercises_to_workout: %w", err)
	}
	if len(args.ExerciseIDs) == 0 {
		return "", fmt.Errorf("exercise_ids is required and must not be empty")
	}

	for _, exerciseIDStr := range args.ExerciseIDs {
		exerciseID, err := domain.ParseID(exerciseIDStr)
		if err != nil {
			return "", fmt.Errorf("invalid exercise_id %q: %w", exerciseIDStr, err)
		}

		exercise, err := t.service.GetExerciseByID(ctx, exerciseID)
		if err != nil {
			return "", fmt.Errorf("failed to load exercise %s: %w", exerciseID, err)
		}

		_, err = t.service.LogExercise(ctx, chatCtx.UserID, chatCtx.WorkoutID.V, exerciseID)
		if err != nil {
			return "", fmt.Errorf("failed to log exercise %s: %w", exercise.Name, err)
		}
	}

	return string(`{"success": true}`), nil
}
