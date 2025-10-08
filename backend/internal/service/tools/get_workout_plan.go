package tools

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
	"github.com/opentracing/opentracing-go"
)

var getWorkoutPlanSchema = shared.FunctionParameters{
	"type":                 "object",
	"additionalProperties": false,
}

func (t *Tools) newGetWorkoutPlanTool() agentTool {
	return agentTool{
		name: "get_workout_plan",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("get_workout_plan"),
				Description: openai.String("Return the current state of the workout including exercises, expected sets, and logged performance."),
				Parameters:  openai.F(getWorkoutPlanSchema),
				Strict:      openai.Bool(true),
			}),
		},
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
