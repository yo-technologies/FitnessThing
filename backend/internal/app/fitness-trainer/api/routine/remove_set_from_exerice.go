package routine

import (
	"context"
	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/domain"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) RemoveSetFromExerciseInstance(ctx context.Context, in *desc.RemoveSetFromExerciseInstanceRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.RemoveSetFromExerciseInstance")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
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

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found in context")
	}

	if err := i.service.RemoveSetFromExerciseInstance(ctx, userID, routineID, exerciseInstanceID, setID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
