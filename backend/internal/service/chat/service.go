package chat

import (
	"context"

	openai_client "fitness-trainer/internal/clients/openai"
	"fitness-trainer/internal/config"

	"fitness-trainer/internal/domain"

	"github.com/openai/openai-go/v3"
)

type toolsService interface {
	ChatAgentToolDefinitions() []openai.ChatCompletionToolUnionParam
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

type Service struct {
	toolsService         toolsService
	chatRepository       chatRepository
	workoutRepository    workoutRepository
	userPromptRepository userPromptRepository

	openAIClient openai_client.ChatClient
	config       *config.Config
}

func New(
	toolsService toolsService,
	chatRepository chatRepository,
	workoutRepository workoutRepository,
	userPromptRepository userPromptRepository,
	openAIClient openai_client.ChatClient,
	config *config.Config,
) *Service {
	return &Service{
		toolsService:         toolsService,
		chatRepository:       chatRepository,
		workoutRepository:    workoutRepository,
		userPromptRepository: userPromptRepository,
		openAIClient:         openAIClient,
		config:               config,
	}
}
