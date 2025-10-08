package chat

import (
	"context"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetChat(ctx context.Context, userID domain.ID, req dto.GetChatRequest) (dto.GetChatDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetChat")
	defer span.Finish()

	if !req.ChatID.IsValid && !req.WorkoutID.IsValid {
		return dto.GetChatDTO{}, fmt.Errorf("either chat_id or workout_id must be provided")
	}

	var chat domain.Chat
	var err error

	// Определяем, как искать чат: по workout_id или по chat_id
	if req.ChatID.IsValid {
		chat, err = s.chatRepository.GetChatByID(ctx, req.ChatID.V)
		if err != nil {
			return dto.GetChatDTO{}, fmt.Errorf("failed to get chat by id: %w", err)
		}
	}
	if req.WorkoutID.IsValid {
		chat, err = s.chatRepository.GetChatByWorkoutID(ctx, req.WorkoutID.V)
		if err != nil {
			return dto.GetChatDTO{}, fmt.Errorf("failed to get chat by workout id: %w", err)
		}
	}

	// Проверяем, что пользователь имеет доступ к этому чату
	if chat.UserID != userID {
		return dto.GetChatDTO{}, domain.ErrForbidden
	}

	messages, err := s.chatRepository.ListChatMessages(ctx, chat.ID, 1000, 0) // limit 1000, offset 0
	if err != nil {
		return dto.GetChatDTO{}, fmt.Errorf("failed to list chat messages: %w", err)
	}

	return dto.GetChatDTO{
		Chat:     chat,
		Messages: messages,
	}, nil
}
