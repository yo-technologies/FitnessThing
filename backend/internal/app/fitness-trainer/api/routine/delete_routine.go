package routine

import (
	"context"
	"fmt"

	"fitness-trainer/internal/domain"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) DeleteRoutine(ctx context.Context, in *desc.DeleteRoutineRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.DeleteRoutine")
	defer span.Finish()

	if err := in.ValidateAll(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	routineID, err := domain.ParseID(in.GetRoutineId())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	if err := i.service.DeleteRoutine(ctx, routineID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
