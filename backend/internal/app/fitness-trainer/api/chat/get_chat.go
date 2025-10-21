package chat

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) GetChat(ctx context.Context, in *desc.GetChatRequest) (*desc.GetChatResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.GetChat")
	defer span.Finish()

	if in == nil {
		return nil, fmt.Errorf("%w: empty request", domain.ErrInvalidArgument)
	}

	if err := in.ValidateAll(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("error getting user id from context")
		return nil, domain.ErrInternal
	}

	// Создаём запрос с поддержкой обоих параметров
	req := dto.GetChatRequest{}

	if in.GetChatId() != "" {
		chatID, err := domain.ParseID(in.GetChatId())
		if err != nil {
			return nil, fmt.Errorf("%w: invalid chat_id", domain.ErrInvalidArgument)
		}
		req.ChatID = utils.NewNullable(chatID, true)
	}

	if in.GetWorkoutId() != "" {
		workoutID, err := domain.ParseID(in.GetWorkoutId())
		if err != nil {
			return nil, fmt.Errorf("%w: invalid workout_id", domain.ErrInvalidArgument)
		}
		req.WorkoutID = utils.NewNullable(workoutID, true)
	}

	// Проверяем, что хотя бы один из параметров указан
	if !req.ChatID.IsValid && !req.WorkoutID.IsValid {
		return nil, fmt.Errorf("%w: either chat_id or workout_id must be provided", domain.ErrInvalidArgument)
	}

	result, err := i.service.GetChat(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	chat := mappers.ChatToProto(result.Chat)

	messages := make([]*desc.ChatMessage, len(result.Messages))
	for i, msg := range result.Messages {
		protoMsg, err := mappers.ChatMessageToProto(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to map chat message: %w", err)
		}
		messages[i] = protoMsg
	}

	return &desc.GetChatResponse{
		Chat:     chat,
		Messages: messages,
	}, nil
}