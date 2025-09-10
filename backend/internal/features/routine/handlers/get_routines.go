package handlers

import (
	"context"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) GetRoutines(ctx context.Context, _ *emptypb.Empty) (*desc.RoutineListResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.routine.GetRoutines")
	defer span.Finish()

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("user id not found in context")
		return nil, domain.ErrUnauthorized
	}

	routines, err := i.service.GetRoutines(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &desc.RoutineListResponse{
		Routines: mappers.RoutinesToProto(routines),
	}, nil
}
