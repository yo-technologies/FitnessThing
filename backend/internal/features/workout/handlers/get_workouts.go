package handlers

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) GetWorkouts(ctx context.Context, in *desc.GetWorkoutsRequest) (*desc.GetWorkoutsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.GetWorkouts")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("user id not found in context")
		return nil, domain.ErrInternal
	}

	var limit, offset int
	{
		if in.Limit <= 0 {
			limit = 10
		} else {
			limit = int(in.GetLimit())
		}

		if in.Offset <= 0 {
			offset = 0
		} else {
			offset = int(in.GetOffset())
		}
	}

	workoutsDTO, err := i.service.GetWorkouts(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &desc.GetWorkoutsResponse{
		Workouts: mappers.WorkoutsDTOToProto(workoutsDTO),
	}, nil
}
