package openai_llm

import (
	"context"
	"fitness-trainer/internal/llm"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (c *CompletionProvider) CreateCompletion(ctx context.Context, p llm.ChatParams) (string, llm.Usage, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "llm.openai.CreateCompletion")
	defer span.Finish()

	params := c.newOpenAIParams(p)

	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", llm.Usage{}, fmt.Errorf("failed to create chat completion for prompt: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", llm.Usage{}, fmt.Errorf("empty response from OpenAI")
	}

	usage := llm.Usage{}
	if completion.Usage.TotalTokens != 0 || completion.Usage.PromptTokens != 0 || completion.Usage.CompletionTokens != 0 {
		usage = llm.Usage{PromptTokens: int(completion.Usage.PromptTokens), CompletionTokens: int(completion.Usage.CompletionTokens), TotalTokens: int(completion.Usage.TotalTokens)}
	}

	return completion.Choices[0].Message.Content, usage, nil
}
