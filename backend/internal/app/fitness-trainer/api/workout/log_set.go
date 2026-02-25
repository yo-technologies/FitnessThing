package workout

import (
	"context"
	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) LogSet(ctx context.Context, in *desc.LogSetRequest) (*desc.SetLogResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.LogSet")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("failed to get user id from context")
		return nil, domain.ErrInternal
	}

	workoutID, err := domain.ParseID(in.WorkoutId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	exerciseLogID, err := domain.ParseID(in.ExerciseLogId)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	var setLogDTO dto.CreateSetLogDTO
	{
		setLogDTO.Reps = int(in.Reps)
		setLogDTO.Weight = in.Weight
	}

	setLog, err := i.service.LogSet(ctx, userID, workoutID, exerciseLogID, setLogDTO)
	if err != nil {
		return nil, err
	}

	return &desc.SetLogResponse{
		SetLog: mappers.SetLogToProto(setLog),
	}, nil
}
