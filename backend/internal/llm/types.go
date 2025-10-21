package llm

import "context"

// CompletionProvider abstracts completion/generation API
// Implementations may call any LLM provider (OpenAI, etc.)
type CompletionProvider interface {
	// CreateCompletion generates a completion given system and user prompts.
	CreateCompletion(ctx context.Context, params ChatParams) (string, Usage, error)
	// CreateCompletionStream initiates a streaming chat completion.
	CreateCompletionStream(ctx context.Context, params ChatParams) (ChatStream, error)
}

// Roles for chat messages
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// ToolDefinition describes a callable tool for the assistant
type ToolDefinition struct {
	Name           string
	Description    string
	Parameters     map[string]any // JSON Schema for parameters as a map
}

// MessageParam is a message input to the model
type MessageParam struct {
	Role    string
	Content string
	// Assistant-only: optional planned tool calls
	ToolCalls []ToolCall
	// Tool-only: references tool call id
	ToolCallID string
}

// ChatParams groups arguments for chat completion
type ChatParams struct {
	Messages     []MessageParam
	Tools        []ToolDefinition
	IncludeUsage bool
}

// Usage token accounting
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ToolCall is a fully specified tool invocation
type ToolCall struct {
	ID        string
	Name      string
	Arguments string // raw JSON
}

// ToolCallDelta accumulates incremental tool call info in stream
type ToolCallDelta struct {
	Index     int
	ID        string
	Name      string
	Arguments string
}

// ChatDelta is a single streamed update
type ChatDelta struct {
	Content   string
	ToolCalls []ToolCallDelta
	Usage     Usage
}

// ChatMessage final assistant message
type ChatMessage struct {
	Role      string
	Content   string
	ToolCalls []ToolCall
}

// ChatStream abstracts streaming responses
type ChatStream interface {
	Next() bool
	Chunk() ChatDelta
	Err() error
	Close() error
}
