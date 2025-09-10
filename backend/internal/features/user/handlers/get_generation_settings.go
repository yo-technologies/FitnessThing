package handlers

import (
	"context"
	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) GetWorkoutGenerationSettings(ctx context.Context, _ *emptypb.Empty) (*desc.WorkoutGenerationSettingsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.user.GetWorkoutGenerationSettings")
	defer span.Finish()

	id, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user id not found in context: %w", domain.ErrUnauthorized)
	}

	settings, err := i.service.GetGenerationSettings(ctx, id)
	if err != nil {
		return nil, err
	}

	return &desc.WorkoutGenerationSettingsResponse{
		Settings: mappers.GenerationSettingsToProto(settings),
	}, nil
}
