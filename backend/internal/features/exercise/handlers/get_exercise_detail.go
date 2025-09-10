package handlers

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) GetExerciseDetail(ctx context.Context, in *desc.GetExerciseDetailRequest) (*desc.ExerciseResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.exercise.GetExerciseDetail")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	id, err := domain.ParseID(in.GetExerciseId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exercise, err := i.service.GetExerciseByID(ctx, id)
	if err != nil {
		logger.Errorf("error getting exercise detail: %v", err)
		return nil, err
	}

	return &desc.ExerciseResponse{
		Exercise: mappers.ExerciseToProto(exercise),
	}, nil
}
