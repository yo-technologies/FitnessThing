package tools

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/opentracing/opentracing-go"
)

func (t *Tools) newListExercisesTool() agentTool {
	schema := shared.FunctionParameters{
		"type":     "object",
		"required": []string{"muscle_group_ids", "limit", "exclude_exercise_ids"},
		"properties": map[string]any{
			"muscle_group_ids": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string", "description": "UUID of the muscle group"},
			},
			"exclude_exercise_ids": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string", "description": "UUID of exercises to exclude"},
			},
			"limit": map[string]any{
				"type":        "integer",
				"minimum":     1,
				"maximum":     50,
				"description": "Maximum number of exercises to return. Defaults to 10.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "list_exercises",
		definition: openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "list_exercises",
			Description: openai.String("Fetch exercises optionally filtered by muscle group ids."),
			Parameters:  schema,
			Strict:      openai.Bool(true),
		}),
		handler: t.listExercisesHandler,
	}
}

func (t *Tools) listExercisesHandler(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.list_exercises")
	defer span.Finish()

	var args listExercisesArgs
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("invalid arguments for list_exercises: %w", err)
		}
	}

	limit := 10
	if args.Limit != nil {
		limit = *args.Limit
	}

	toIDs := func(values []string) ([]domain.ID, error) {
		ids := make([]domain.ID, 0, len(values))
		for _, v := range values {
			if v == "" {
				continue
			}
			id, err := domain.ParseID(v)
			if err != nil {
				logger.Errorf("failed to parse id %s: %w", v, err)
				continue
			}
			ids = append(ids, id)
		}
		return ids, nil
	}

	muscleGroupIDs, err := toIDs(args.MuscleGroupIDs)
	if err != nil {
		return "", err
	}

	excludedExerciseIDs, err := toIDs(args.ExcludeExerciseIDs)
	if err != nil {
		return "", err
	}

	exercises, err := t.service.GetExercises(ctx, muscleGroupIDs, excludedExerciseIDs)
	if err != nil {
		return "", fmt.Errorf("failed to get exercises: %w", err)
	}

	if len(exercises) > limit {
		exercises = exercises[:limit]
	}

	payload := convertExercisesToListResponse(exercises)

	rawResp, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal exercises: %w", err)
	}

	return string(rawResp), nil
}
