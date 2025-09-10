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

func (i *Implementation) GetExerciseInstanceDetails(ctx context.Context, in *desc.GetExerciseInstanceDetailsRequest) (*desc.GetExerciseInstanceDetailsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.GetExerciseInstanceDetails")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("failed to get user id from context")
		return nil, domain.ErrInternal
	}

	routineID, err := domain.ParseID(in.GetRoutineId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseInstanceID, err := domain.ParseID(in.GetExerciseInstanceId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseInstance, err := i.service.GetExerciseInstance(ctx, userID, routineID, exerciseInstanceID)
	if err != nil {
		return nil, err
	}

	return &desc.GetExerciseInstanceDetailsResponse{
		ExerciseInstanceDetails: mappers.ExerciseInstanceDetailToProto(exerciseInstance),
	}, nil
}
