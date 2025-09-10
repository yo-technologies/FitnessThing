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

func (i *Implementation) AddSetToExerciseInstance(ctx context.Context, in *desc.AddSetToExerciseInstanceRequest) (*desc.SetResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.AddSetToExerciseInstance")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id is not found in context")
	}

	routineID, err := domain.ParseID(in.RoutineId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseInstanceID, err := domain.ParseID(in.ExerciseInstanceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	var createSetDTO dto.CreateSetDTO
	{
		createSetDTO.ExerciseInstanceID = exerciseInstanceID

		createSetDTO.SetType = mappers.SetTypeFromProto(in.GetSetType())
		if createSetDTO.SetType == domain.SetTypeUnknown {
			return nil, fmt.Errorf("%w: unknown set type", domain.ErrInvalidArgument)
		}

		createSetDTO.Reps = utils.NewNullable(int(in.GetReps()), in.GetReps() != 0)
		createSetDTO.Weight = utils.NewNullable(float32(in.GetWeight()), in.GetWeight() != 0)
		createSetDTO.Time = utils.NewNullable(in.GetTime().AsDuration(), in.GetTime().AsDuration() != 0)
	}

	set, err := i.service.AddSetToExerciseInstance(ctx, userID, routineID, exerciseInstanceID, createSetDTO)
	if err != nil {
		return nil, err
	}

	return &desc.SetResponse{
		Set: mappers.SetToProto(set),
	}, nil
}
