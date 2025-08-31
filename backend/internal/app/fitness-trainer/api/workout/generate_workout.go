package workout

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

// GenerateWorkout запускает (или перезапускает) асинхронную генерацию тренировки
func (i *Implementation) GenerateWorkout(ctx context.Context, in *desc.GenerateWorkoutRequest) (*desc.GenerateWorkoutResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.GenerateWorkout")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("user id not found in context")
		return nil, domain.ErrInternal
	}

	workoutID, err := domain.ParseID(in.GetWorkoutId())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	err = i.service.GenerateWorkout(ctx, userID, workoutID, in.GetUserPrompt())
	if err != nil {
		return nil, err
	}

	return &desc.GenerateWorkoutResponse{}, nil
}
