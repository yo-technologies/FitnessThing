package chat

import (
	"context"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/llm"
)

type toolsService interface {
	ChatAgentToolDefinitions() []llm.ToolDefinition
	ExecuteChatAgentTool(ctx context.Context, ctxData domain.AgentChatContext, name string, arguments string) (string, error)
}

type chatRepository interface {
	CreateChat(ctx context.Context, chat domain.Chat) (domain.Chat, error)
	GetChatByID(ctx context.Context, id domain.ID) (domain.Chat, error)
	GetChatByWorkoutID(ctx context.Context, workoutID domain.ID) (domain.Chat, error)

	CreateChatMessage(ctx context.Context, message domain.ChatMessage) (domain.ChatMessage, error)
	ListChatMessages(ctx context.Context, chatID domain.ID, limit, offset int) ([]domain.ChatMessage, error)
}

type workoutRepository interface {
	GetWorkoutByID(ctx context.Context, id domain.ID) (domain.Workout, error)
}

type userPromptRepository interface {
	GetLastPromptByUserID(ctx context.Context, userID domain.ID) (domain.Prompt, error)
}

type quotaService interface {
	Reserve(ctx context.Context, userID domain.ID, n int) (bool, error)
	Confirm(ctx context.Context, userID domain.ID, reserved int, actual int) error
}

type Service struct {
	toolsService         toolsService
	chatRepository       chatRepository
	workoutRepository    workoutRepository
	userPromptRepository userPromptRepository
	llmClient            llm.CompletionProvider
	quotaService         quotaService
}

func New(
	toolsService toolsService,
	chatRepository chatRepository,
	workoutRepository workoutRepository,
	userPromptRepository userPromptRepository,
	llmClient llm.CompletionProvider,
	quotaSvc quotaService,
) *Service {
	return &Service{
		toolsService:         toolsService,
		chatRepository:       chatRepository,
		workoutRepository:    workoutRepository,
		userPromptRepository: userPromptRepository,
		llmClient:            llmClient,
		quotaService:         quotaSvc,
	}
}
