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

func (t *Tools) newListMuscleGroupsTool() agentTool {
	return agentTool{
		name: "list_muscle_groups",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("list_muscle_groups"),
				Description: openai.String("Return the available muscle groups with their identifiers."),
				Parameters: openai.F(shared.FunctionParameters{
					"type": "object",
				}),
				Strict: openai.Bool(true),
			}),
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
