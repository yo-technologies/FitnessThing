package handlers

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/shared/domain/dto"
	"fitness-trainer/internal/utils"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) UpdateRoutine(ctx context.Context, in *desc.UpdateRoutineRequest) (*desc.RoutineResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.UpdateRoutine")
	defer span.Finish()

	if err := in.ValidateAll(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	routineID, err := domain.ParseID(in.GetRoutineId())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	var updateDTO dto.UpdateRoutineDTO
	{
		updateDTO.Name = utils.NewNullable(in.GetName(), in.Name != nil)
		updateDTO.Description = utils.NewNullable(in.GetDescription(), in.Description != nil)
	}

	routine, err := i.service.UpdateRoutine(ctx, routineID, updateDTO)
	if err != nil {
		return nil, err
	}

	return &desc.RoutineResponse{
		Routine: mappers.RoutineToProto(routine),
	}, nil
}
