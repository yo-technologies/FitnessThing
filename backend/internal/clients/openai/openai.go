package openai_client

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/opentracing/opentracing-go"
)

type ChatClient interface {
	CreateChatCompletion(ctx context.Context, params openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error)
	CreateChatCompletionStream(ctx context.Context, params openai.ChatCompletionNewParams, opts ...option.RequestOption) (ChatCompletionStream, error)
}

type Client struct {
	client openai.Client
}

func New(client openai.Client) *Client {
	return &Client{client: client}
}

func (c *Client) CreateChatCompletion(ctx context.Context, params openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "client.openai.CreateChatCompletion")
	defer span.Finish()

	completion, err := c.client.Chat.Completions.New(ctx, params, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	return completion, nil
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, params openai.ChatCompletionNewParams, opts ...option.RequestOption) (ChatCompletionStream, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "client.openai.CreateChatCompletionStream")
	defer span.Finish()

	stream := c.client.Chat.Completions.NewStreaming(ctx, params, opts...)
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("failed to create chat completion stream: %w", err)
	}

	return &chatCompletionStream{stream: stream}, nil
}

type ChatCompletionStream interface {
	Next() bool
	Chunk() openai.ChatCompletionChunk
	Err() error
	Close() error
}

type chatCompletionStream struct {
	stream *ssestream.Stream[openai.ChatCompletionChunk]
}

func (s *chatCompletionStream) Next() bool {
	return s.stream.Next()
}

func (s *chatCompletionStream) Chunk() openai.ChatCompletionChunk {
	return s.stream.Current()
}

func (s *chatCompletionStream) Err() error {
	return s.stream.Err()
}

func (s *chatCompletionStream) Close() error {
	return s.stream.Close()
}
