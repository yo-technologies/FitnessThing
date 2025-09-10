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

func (i *Implementation) GetWorkout(ctx context.Context, in *desc.GetWorkoutRequest) (*desc.GetWorkoutResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.GetWorkout")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("user id not found in context")
		return nil, domain.ErrInternal
	}

	workoutID, err := domain.ParseID(in.WorkoutId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	workouDTO, err := i.service.GetWorkout(ctx, userID, workoutID)
	if err != nil {
		return nil, err
	}

	return &desc.GetWorkoutResponse{
		Workout:      mappers.WorkoutToProto(workouDTO.Workout),
		ExerciseLogs: mappers.ExerciseLogDTOsToProto(workouDTO.ExerciseLogs),
	}, nil
}
