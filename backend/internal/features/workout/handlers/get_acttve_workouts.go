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

func (i *Implementation) GetActiveWorkouts(ctx context.Context, in *emptypb.Empty) (*desc.WorkoutsListResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.GetActiveWorkouts")
	defer span.Finish()

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("failed to get user id from context")
		return nil, domain.ErrInternal
	}

	workouts, err := i.service.GetActiveWorkouts(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mappers.WorkoutsToProto(workouts), nil
}
