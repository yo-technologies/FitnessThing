package handlers

import (
	"context"

	"fitness-trainer/internal/app/mappers"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) GetMuscleGroups(ctx context.Context, in *emptypb.Empty) (*desc.GetMuscleGroupsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.exercise.GetMuscleGroups")
	defer span.Finish()
	
	muscleGroups, err := i.service.GetMuscleGroups(ctx)
	if err != nil {
		return nil, err
	}

	return &desc.GetMuscleGroupsResponse{
		MuscleGroups: mappers.MuscleGroupDTOsToProto(muscleGroups),
	}, nil
}
