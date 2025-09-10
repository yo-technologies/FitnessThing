package handlers

import (
	"context"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/shared/domain/dto"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) CreateExercise(ctx context.Context, in *desc.CreateExerciseRequest) (*desc.ExerciseResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.CreateExercise")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	var exerciseDTO dto.CreateExerciseDTO
	{
		exerciseDTO.Name = in.Name
		
		if in.Description != nil {
			exerciseDTO.Description = utils.NewNullable(in.GetDescription(), true)
		}

		if in.VideoUrl != nil {
			exerciseDTO.VideoURL = utils.NewNullable(in.GetVideoUrl(), true)
		}

		for _, muscleGroupID := range in.TargetMuscleGroupIds {
			id, err := domain.ParseID(muscleGroupID)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
			}

			exerciseDTO.TargetMuscleGroups = append(exerciseDTO.TargetMuscleGroups, id)
		}
	}

	exercise, err := i.service.CreateExercise(
		ctx,
		exerciseDTO,
	)
	if err != nil {
		logger.Errorf("error creating exercise: %v", err)
		return nil, err
	}

	return &desc.ExerciseResponse{
		Exercise: mappers.ExerciseToProto(exercise),
	}, nil
}
