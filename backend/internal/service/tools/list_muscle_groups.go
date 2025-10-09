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

func (t *Tools) newListMuscleGroupsTool() agentTool {
	return agentTool{
		name: "list_muscle_groups",
		definition: openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "list_muscle_groups",
			Description: openai.String("Return the available muscle groups with their identifiers."),
			Parameters: shared.FunctionParameters{
				"type":                 "object",
				"required":             []string{},
				"properties":           map[string]any{},
				"additionalProperties": false,
			},
			Strict: openai.Bool(true),
		}),
		handler: t.listMuscleGroupsHandler,
	}
}

func (t *Tools) listMuscleGroupsHandler(ctx context.Context, chatCtx domain.AgentChatContext, _ json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.list_muscle_groups")
	defer span.Finish()

	groups, err := t.service.GetMuscleGroups(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load muscle groups: %w", err)
	}

	payload := listMuscleGroupsResponse{MuscleGroups: groups}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal muscle groups: %w", err)
	}

	return string(raw), nil
}
