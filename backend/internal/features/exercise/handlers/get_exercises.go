package handlers

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) GetExercises(ctx context.Context, in *desc.GetExercisesRequest) (*desc.GetExercisesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.exercise.GetExercises")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	muscleGroupIDs := make([]domain.ID, 0, len(in.GetMuscleGroupIds()))
	for _, mg := range in.GetMuscleGroupIds() {
		muscleGroupID, err := domain.ParseID(mg)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
		}
		muscleGroupIDs = append(muscleGroupIDs, muscleGroupID)
	}

	excludedExerciseIDs := make([]domain.ID, 0, len(in.GetExcludeExerciseIds()))
	for _, exerciseID := range in.GetExcludeExerciseIds() {
		id, err := domain.ParseID(exerciseID)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
		}
		excludedExerciseIDs = append(excludedExerciseIDs, id)
	}

	exercises, err := i.service.GetExercises(ctx, muscleGroupIDs, excludedExerciseIDs)
	if err != nil {
		return nil, err
	}

	return &desc.GetExercisesResponse{
		Exercises: mappers.ExercisesToProto(exercises),
	}, nil
}
