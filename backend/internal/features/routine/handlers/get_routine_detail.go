package handlers

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) GetRoutineDetail(ctx context.Context, in *desc.GetRoutineDetailRequest) (*desc.RoutineDetailResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.GetRoutineDetail")
	defer span.Finish()
	
	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	routineID, err := domain.ParseID(in.GetRoutineId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	routine, err := i.service.GetRoutineByID(ctx, routineID)
	if err != nil {
		return nil, err
	}

	return mappers.RoutineDetailsDTOToProto(routine), nil
}
