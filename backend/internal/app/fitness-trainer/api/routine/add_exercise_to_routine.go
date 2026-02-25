package routine

import (
	"context"
	"fitness-trainer/internal/domain"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *Implementation) AddExerciseToRoutine(ctx context.Context, in *desc.AddExerciseToRoutineRequest) (*desc.RoutineInstanceResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.AddExerciseToRoutine")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	routineID, err := domain.ParseID(in.GetRoutineId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseID, err := domain.ParseID(in.GetExerciseId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseInstance, err := i.service.AddExerciseToRoutine(ctx, routineID, exerciseID)
	if err != nil {
		return nil, err
	}

	return &desc.RoutineInstanceResponse{
		ExerciseInstance: &desc.ExerciseInstance{
			Id:         exerciseInstance.ID.String(),
			ExerciseId: exerciseInstance.ExerciseID.String(),
			RoutineId:  exerciseInstance.RoutineID.String(),
			CreatedAt:  timestamppb.New(exerciseInstance.CreatedAt),
		},
	}, nil
}
