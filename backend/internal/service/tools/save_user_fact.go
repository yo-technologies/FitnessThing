package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"fitness-trainer/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (t *Tools) newSaveUserFactTool() agentTool {
	schema := map[string]any{
		"type":                 "object",
		"required":             []string{"fact"},
		"additionalProperties": false,
		"properties": map[string]any{
			"fact": map[string]any{
				"type":        "string",
				"description": fmt.Sprintf("User fact (max %d characters). Use for stable preferences, injuries, routines.", domain.MaxUserFactLength),
				"maxLength":   domain.MaxUserFactLength,
			},
		},
	}

	return agentTool{
		name:    "save_user_fact",
		desc:    "Store a concise fact about the user (preferences, injuries, lifestyle notes) for future context.",
		params:  schema,
		handler: t.saveUserFactHandler,
	}
}

func (t *Tools) saveUserFactHandler(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.save_user_fact")
	defer span.Finish()

	var args saveUserFactArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", fmt.Errorf("invalid arguments for save_user_fact: %w", err)
	}

	factText := strings.TrimSpace(args.Fact)
	if factText == "" {
		return "", fmt.Errorf("fact is required")
	}

	fact, err := t.service.SaveUserFact(ctx, chatCtx.UserID, factText)
	if err != nil {
		return "", fmt.Errorf("failed to save user fact: %w", err)
	}

	payload := userFactPayloadFromDomain(fact)
	resp, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal save_user_fact response: %w", err)
	}

	return string(resp), nil
}
