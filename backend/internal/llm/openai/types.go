package openai_llm

import (
	"fitness-trainer/internal/llm"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/shared"
)

type chatStream struct {
	inner *ssestream.Stream[openai.ChatCompletionChunk]
}

func (s *chatStream) Next() bool   { return s.inner.Next() }
func (s *chatStream) Err() error   { return s.inner.Err() }
func (s *chatStream) Close() error { return s.inner.Close() }

func (s *chatStream) Chunk() llm.ChatDelta {
	chunk := s.inner.Current()

	delta := llm.ChatDelta{}

	if len(chunk.Choices) > 0 {
		delta.Content = chunk.Choices[0].Delta.Content
	}

	if chunk.Usage.TotalTokens != 0 || chunk.Usage.PromptTokens != 0 || chunk.Usage.CompletionTokens != 0 {
		delta.Usage = llm.Usage{PromptTokens: int(chunk.Usage.PromptTokens), CompletionTokens: int(chunk.Usage.CompletionTokens), TotalTokens: int(chunk.Usage.TotalTokens)}
	}

	if len(chunk.Choices) > 0 {
		for _, tc := range chunk.Choices[0].Delta.ToolCalls {
			delta.ToolCalls = append(delta.ToolCalls, llm.ToolCallDelta{Index: int(tc.Index), ID: tc.ID, Name: tc.Function.Name, Arguments: tc.Function.Arguments})
		}
	}

	return delta
}

// toOpenAIMessages maps generic LLM messages to OpenAI chat message params.
func (c *CompletionProvider) toOpenAIMessages(msgs []llm.MessageParam) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case llm.RoleSystem:
			out = append(out, openai.SystemMessage(m.Content))
		case llm.RoleUser:
			out = append(out, openai.UserMessage(m.Content))
		case llm.RoleAssistant:
			assistant := openai.ChatCompletionAssistantMessageParam{}
			if m.Content != "" {
				assistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{OfString: openai.String(m.Content)}
			}
			if len(m.ToolCalls) > 0 {
				calls := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(m.ToolCalls))
				for _, tc := range m.ToolCalls {
					calls = append(calls, openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID: tc.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      tc.Name,
								Arguments: tc.Arguments,
							},
						},
					})
				}
				assistant.ToolCalls = calls
			}
			out = append(out, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant})
		case llm.RoleTool:
			out = append(out, openai.ToolMessage(m.Content, m.ToolCallID))
		}
	}
	return out
}

// toOpenAITools maps provider-agnostic tool definitions to OpenAI format.
func (c *CompletionProvider) toOpenAITools(tools []llm.ToolDefinition) []openai.ChatCompletionToolUnionParam {
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))
	for _, t := range tools {
		var schema shared.FunctionParameters
		if t.Parameters != nil {
			schema = t.Parameters
		}
		out = append(out, openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        t.Name,
			Description: openai.String(t.Description),
			Parameters:  schema,
			Strict:      openai.Bool(true),
		}))
	}
	return out
}

// newOpenAIParams constructs OpenAI request from generic params and mapped components.
func (c *CompletionProvider) newOpenAIParams(p llm.ChatParams) openai.ChatCompletionNewParams {
	o := openai.ChatCompletionNewParams{
		Model:           shared.ChatModel(c.cfg.GetOpenAIModel()),
		Messages:        c.toOpenAIMessages(p.Messages),
		ReasoningEffort: shared.ReasoningEffort(c.cfg.GetReasoningEffort()),
		Temperature:     openai.Float(1.0),
	}
	if len(p.Tools) > 0 {
		o.Tools = c.toOpenAITools(p.Tools)
		o.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{OfAuto: openai.String("auto")}
	}
	if p.IncludeUsage {
		o.StreamOptions = openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)}
	}
	return o
}
