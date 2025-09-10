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

func (s *Implementation) GetExerciseLogDetails(ctx context.Context, in *desc.GetExerciseLogDetailRequest) (*desc.ExerciseLogResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.GetExerciseLogDetails")
	defer span.Finish()
	
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("failed to get user id from context")
		return nil, domain.ErrInternal
	}

	_, err := domain.ParseID(in.WorkoutId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseLogID, err := domain.ParseID(in.ExerciseLogId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseLog, err := s.service.GetExerciseLog(ctx, userID, exerciseLogID)
	if err != nil {
		return nil, err
	}

	return &desc.ExerciseLogResponse{
		ExerciseLogDetails: mappers.ExerciseLogDTOToProto(exerciseLog),
	}, nil
}
