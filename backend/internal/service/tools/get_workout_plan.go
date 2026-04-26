package tools

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

var getWorkoutPlanSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []string{},
	"properties":           map[string]any{},
}

func (t *Tools) newGetWorkoutPlanTool() agentTool {
	return agentTool{
		name:    "get_workout_plan",
		desc:    "Return the current state of the workout including exercises, expected sets, and logged performance. Call this before setting weights for any exercise to check which muscle groups have already been worked in this session — use that to apply intra-workout fatigue adjustments.",
		params:  getWorkoutPlanSchema,
		handler: t.getWorkoutPlanToolHandler,
	}
}

func (t *Tools) getWorkoutPlanToolHandler(ctx context.Context, chatCtx domain.AgentChatContext, _ json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.get_workout_plan")
	defer span.Finish()

	if !chatCtx.WorkoutID.IsValid {
		return "", fmt.Errorf("workout context is required for get_workout_plan")
	}

	workout, err := t.service.GetWorkout(ctx, chatCtx.UserID, chatCtx.WorkoutID.V)
	if err != nil {
		return "", fmt.Errorf("failed to load workout: %w", err)
	}

	raw, err := json.Marshal(workoutPlanFromDomain(workout))
	if err != nil {
		return "", fmt.Errorf("failed to marshal workout plan: %w", err)
	}

	return string(raw), nil
}
