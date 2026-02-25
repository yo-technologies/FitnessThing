package routine

import (
	"context"
	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) RemoveExerciseInstanceFromRoutine(ctx context.Context, in *desc.RemoveExerciseInstanceFromRoutineRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.RemoveExerciseInstanceFromRoutine")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	routineID, err := domain.ParseID(in.RoutineId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseInstanceID, err := domain.ParseID(in.ExerciseInstanceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Error("user id not found in context")
		return nil, domain.ErrUnauthorized
	}

	if err := i.service.RemoveExerciseInstanceFromRoutine(ctx, userID, routineID, exerciseInstanceID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
