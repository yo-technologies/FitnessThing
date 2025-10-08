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

func (t *Tools) newGetWorkoutHistoryTool() agentTool {
	schema := shared.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"limit": map[string]any{
				"type":        "integer",
				"minimum":     1,
				"maximum":     20,
				"description": "Maximum number of workouts to return. Defaults to 5.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "get_workout_history",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("get_workout_history"),
				Description: openai.String("Return recent workouts for the current user including exercises."),
				Parameters:  openai.F(schema),
				Strict:      openai.Bool(true),
			}),
		},
		handler: t.getWorkoutHistoryHandler,
	}
}

func (t *Tools) getWorkoutHistoryHandler(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.get_workout_history")
	defer span.Finish()

	if chatCtx.UserID == (domain.ID{}) {
		return "", fmt.Errorf("user context is required for get_workout_history")
	}

	args := getWorkoutHistoryArgs{Limit: 5}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("invalid arguments for get_workout_history: %w", err)
		}
		if args.Limit <= 0 {
			args.Limit = 5
		}
	}

	workoutsDTO, err := t.service.GetWorkouts(ctx, chatCtx.UserID, args.Limit, 0)
	if err != nil {
		return "", fmt.Errorf("failed to load workouts: %w", err)
	}

	payload, err := t.convertWorkoutsToHistoryResponse(ctx, workoutsDTO)
	if err != nil {
		return "", err
	}

	rawResp, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workout history: %w", err)
	}

	return string(rawResp), nil
}
