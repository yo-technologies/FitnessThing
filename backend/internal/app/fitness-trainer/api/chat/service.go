package chat

import (
	"context"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	desc "fitness-trainer/pkg/workouts"
)

type Service interface {
	SendChatMessageStream(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest, callbacks dto.ChatStreamCallbacks) (dto.ChatCompletionDTO, error)
	GetChat(ctx context.Context, userID domain.ID, req dto.GetChatRequest) (dto.GetChatDTO, error)
}

type Implementation struct {
	service Service
	desc.UnimplementedChatServiceServer
}

func New(service Service) *Implementation {
	return &Implementation{service: service}
}
