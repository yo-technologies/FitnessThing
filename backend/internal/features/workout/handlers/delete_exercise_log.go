package handlers

import (
	"context"
	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/logger"
	"fmt"

	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) DeleteExerciseLog(ctx context.Context, in *desc.DeleteExerciseLogRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.DeleteExerciseLog")
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

	exerciseLogID, err := domain.ParseID(in.GetExerciseLogId())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	if err := i.service.DeleteExerciseLog(ctx, userID, workoutID, exerciseLogID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
