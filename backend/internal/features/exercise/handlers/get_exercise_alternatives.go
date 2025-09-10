package handlers

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) GetExerciseAlternatives(ctx context.Context, in *desc.GetExerciseAlternativesRequest) (*desc.GetExerciseAlternativesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.exercise.GetExerciseAlternatives")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	id, err := domain.ParseID(in.GetExerciseId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exercises, err := i.service.GetExerciseAlternatives(ctx, id)
	if err != nil {
		return nil, err
	}

	return &desc.GetExerciseAlternativesResponse{
		Alternatives: mappers.ExercisesToProto(exercises),
	}, nil
}
