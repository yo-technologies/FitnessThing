package handlers

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

// func (i *Implementation) SetExerciseOrder
func (i *Implementation) SetExerciseOrder(ctx context.Context, in *desc.SetExerciseOrderRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.SetExerciseOrder")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("failed to get user id from context")
		return nil, domain.ErrInternal
	}

	routineID, err := domain.ParseID(in.GetRoutineId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	ids := make([]domain.ID, 0, len(in.GetExerciseInstanceIds()))
	for _, id := range in.GetExerciseInstanceIds() {
		parsedID, err := domain.ParseID(id)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
		}

		ids = append(ids, parsedID)
	}

	if err := i.service.SetExerciseOrder(ctx, userID, routineID, ids); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
