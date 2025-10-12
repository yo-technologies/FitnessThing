package dto

import (
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/utils"
)

type SendChatMessageRequest struct {
	ChatID    utils.Nullable[domain.ID]
	WorkoutID utils.Nullable[domain.ID]
	Content   string
}

type ChatUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type ChatCompletionDTO struct {
	Chat     domain.Chat
	Messages []domain.ChatMessage
	Usage    utils.Nullable[ChatUsage]
}

type ChatStreamCallbacks struct {
	OnContentDelta  func(string) error
	OnUsage         func(ChatUsage) error
	OnStatus        func(string) error
	OnToolEvent     func(ToolEvent) error
	OnFinalResponse func(ChatCompletionDTO) error
}

type GetChatDTO struct {
	Chat     domain.Chat
	Messages []domain.ChatMessage
}

type GetChatRequest struct {
	ChatID    utils.Nullable[domain.ID]
	WorkoutID utils.Nullable[domain.ID]
}

type ToolEventState int

const (
	ToolStateUnspecified ToolEventState = iota
	ToolInvoking
	ToolCompleted
	ToolError
)

type ToolEvent struct {
	ToolName   string
	ToolCallID string
	ArgsJSON   string
	State      ToolEventState
	Error      string
}
