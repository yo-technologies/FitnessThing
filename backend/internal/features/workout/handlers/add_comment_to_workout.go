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

func (i *Implementation) AddCommentToWorkout(ctx context.Context, in *desc.AddCommentToWorkoutRequest) (*desc.WorkoutResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.AddCommentToWorkout")
	defer span.Finish()

	if err := in.ValidateAll(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("error getting user id from context")
		return nil, domain.ErrInternal
	}

	workoutID, err := domain.ParseID(in.GetWorkoutId())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	comment := in.GetComment()

	workout, err := i.service.AddCommentToWorkout(ctx, userID, workoutID, comment)
	if err != nil {
		return nil, err
	}

	return &desc.WorkoutResponse{
		Workout: mappers.WorkoutToProto(workout),
	}, nil
}
