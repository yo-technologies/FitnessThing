package tools

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (t *Tools) newListMuscleGroupsTool() agentTool {
	return agentTool{
		name: "list_muscle_groups",
		desc: "Return the available muscle groups with their identifiers.",
		params: map[string]any{
			"type":                 "object",
			"required":             []string{},
			"properties":           map[string]any{},
			"additionalProperties": false,
		},
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
