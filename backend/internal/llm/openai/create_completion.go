package openai_llm

import (
	"context"
	"fitness-trainer/internal/llm"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (c *CompletionProvider) CreateCompletion(ctx context.Context, p llm.ChatParams) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "llm.openai.CreateCompletion")
	defer span.Finish()

	params := c.newOpenAIParams(p)

	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion for prompt: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	return completion.Choices[0].Message.Content, nil
}
