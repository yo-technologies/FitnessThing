package handlers

import (
	"context"
	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/shared/domain/dto"
	"fitness-trainer/internal/utils"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) UpdateSetInExerciseInstance(ctx context.Context, in *desc.UpdateSetInExerciseInstanceRequest) (*desc.SetResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.UpdateSetInExerciseInstance")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found in context")
	}

	routineID, err := domain.ParseID(in.RoutineId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseInstanceID, err := domain.ParseID(in.ExerciseInstanceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	setID, err := domain.ParseID(in.SetId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	var updateDTO dto.UpdateSetDTO
	{
		updateDTO.Reps = utils.NewNullable(int(in.GetReps()), in.Reps != nil)
		updateDTO.Weight = utils.NewNullable(in.GetWeight(), in.Weight != nil)
		updateDTO.Time = utils.NewNullable(in.GetTime().AsDuration(), in.Time != nil)
	}

	set, err := i.service.UpdateSetInExerciseInstance(ctx, userID, routineID, exerciseInstanceID, setID, updateDTO)
	if err != nil {
		return nil, err
	}

	return &desc.SetResponse{
		Set: mappers.SetToProto(set),
	}, nil
}
