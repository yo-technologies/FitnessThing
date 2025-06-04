package user

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

func (i *Implementation) UpdateUser(ctx context.Context, in *desc.UpdateUserRequest) (*desc.UserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.user.Update")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	id, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("user id not found in context")
		return nil, domain.ErrUnauthorized
	}

	var input dto.UpdateUserDTO
	{
		input.DateOfBirth = in.GetDateOfBirth().AsTime()

		input.Height = utils.NewNullable(in.GetHeight(), in.GetHeight() != 0)
		input.Weight = utils.NewNullable(in.GetWeight(), in.GetWeight() != 0)
	}

	user, err := i.service.UpdateUser(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return &desc.UserResponse{
		User: mappers.UserToProto(user),
	}, nil
}
