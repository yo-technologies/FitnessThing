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

func (i *Implementation) SendChatMessage(ctx context.Context, in *desc.SendChatMessageRequest) (*desc.SendChatMessageResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.SendChatMessage")
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

	req := dto.SendChatMessageRequest{
		Content: in.GetContent(),
		Stream:  in.GetStream(),
	}

	if chatID := in.GetChatId(); chatID != "" {
		parsed, err := domain.ParseID(chatID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid chat_id", domain.ErrInvalidArgument)
		}
		req.ChatID = utils.NewNullable(parsed, true)
	}

	if workoutID := in.GetWorkoutId(); workoutID != "" {
		parsed, err := domain.ParseID(workoutID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid workout_id", domain.ErrInvalidArgument)
		}
		req.WorkoutID = utils.NewNullable(parsed, true)
	}

	completion, err := i.service.SendChatMessage(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	protoChat := mappers.ChatToProto(completion.Chat)

	protoMessages, err := mappers.ChatMessagesToProto(completion.Messages)
	if err != nil {
		logger.Errorf("failed to map chat messages: %v", err)
		return nil, domain.ErrInternal
	}

	resp := &desc.SendChatMessageResponse{
		Chat:     protoChat,
		Messages: protoMessages,
	}

	if completion.Usage.IsValid {
		resp.Usage = mappers.ChatUsageToProto(completion.Usage.V)
	}

	return resp, nil
}
