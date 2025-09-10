package handlers

import (
	"context"
	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) StartWorkout(ctx context.Context, in *desc.StartWorkoutRequest) (*desc.WorkoutResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.StartWorkout")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("user id not found in context")
		return nil, domain.ErrInternal
	}

	var (
		opts domain.StartWorkoutOpts
		err  error
	)

	if in.RoutineId != nil {
		parsedID, err := domain.ParseID(*in.RoutineId)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
		}
		opts.RoutineID = utils.NewNullable(parsedID, true)
	}

	opts.GenerateWorkout = in.GetGenerateWorkout()
	opts.UserPrompt = in.GetUserPrompt()

	workout, err := i.service.StartWorkout(ctx, userID, opts)
	if err != nil {
		return nil, err
	}

	return &desc.WorkoutResponse{
		Workout: mappers.WorkoutToProto(workout),
	}, nil
}
