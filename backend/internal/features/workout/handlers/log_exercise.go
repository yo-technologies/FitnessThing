package handlers

import (
	"context"
	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) LogExercise(ctx context.Context, in *desc.LogExerciseRequest) (*desc.ExerciseLog, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.LogExercise")
	defer span.Finish()
	
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("failed to get user id from context")
		return nil, domain.ErrInternal
	}

	workoutID, err := domain.ParseID(in.WorkoutId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseID, err := domain.ParseID(in.ExerciseId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseLog, err := i.service.LogExercise(ctx, userID, workoutID, exerciseID)
	if err != nil {
		return nil, err
	}

	return mappers.ExerciseLogToProto(exerciseLog), nil
}
