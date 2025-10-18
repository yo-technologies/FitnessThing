package openai_llm

import (
	"context"
	"fmt"

	"fitness-trainer/internal/llm"
)

// CreateChatCompletionStream implements llm.ChatClient via OpenAI SDK.
func (c *CompletionProvider) CreateCompletionStream(ctx context.Context, p llm.ChatParams) (llm.ChatStream, error) {
	req := c.newOpenAIParams(p)

	stream := c.client.Chat.Completions.NewStreaming(ctx, req)
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("failed to create chat completion stream: %w", err)
	}

	return &chatStream{inner: stream}, nil
}
