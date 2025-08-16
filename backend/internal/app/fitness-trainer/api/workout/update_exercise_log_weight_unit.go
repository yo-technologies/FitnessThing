package workout

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"

	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (s *Implementation) UpdateExerciseLogWeightUnit(ctx context.Context, in *desc.UpdateExerciseLogWeightUnitRequest) (*desc.ExerciseLogResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.UpdateExerciseLogWeightUnit")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("failed to get user id from context")
		return nil, domain.ErrInternal
	}

	workoutID, err := domain.ParseID(in.GetWorkoutId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}
	exerciseLogID, err := domain.ParseID(in.GetExerciseLogId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	logDTO, err := s.service.UpdateExerciseLogWeightUnit(
		ctx,
		userID,
		workoutID,
		exerciseLogID,
		mappers.WeightUnitFromProto(in.GetWeightUnit()),
	)
	if err != nil {
		return nil, err
	}

	return &desc.ExerciseLogResponse{ExerciseLogDetails: mappers.ExerciseLogDTOToProto(logDTO)}, nil
}
