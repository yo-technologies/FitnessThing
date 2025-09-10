package handlers

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/logger"

	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) GetUser(ctx context.Context, in *desc.GetUserRequest) (*desc.UserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.user.Get")
	defer span.Finish()

	span.SetTag("user_id", in.GetUserId())

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	id, err := domain.ParseID(in.GetUserId())
	if err != nil {
		logger.Errorf("error parsing user id: %v", err)
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	user, err := i.service.GetUserByID(ctx, id)
	if err != nil {
		logger.Errorf("error getting user: %v", err)
		return nil, err
	}

	return &desc.UserResponse{
		User: mappers.UserToProto(user),
	}, nil
}
