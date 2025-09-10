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

func (i *Implementation) GetExerciseHistory(ctx context.Context, in *desc.GetExerciseHistoryRequest) (*desc.ExerciseHistoryResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.exercise.GetExerciseHistory")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("error getting user id from context")
		return nil, domain.ErrInternal
	}

	id, err := domain.ParseID(in.GetExerciseId())
	if err != nil {
		logger.Errorf("error parsing exercise id %s: %v", in.GetExerciseId(), err)
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
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

	logs, err := i.service.GetExerciseHistory(ctx, userID, id, offset, limit)
	if err != nil {
		return nil, err
	}

	return &desc.ExerciseHistoryResponse{
		ExerciseLogs: mappers.ExerciseLogDTOsToProto(logs),
	}, nil
}
