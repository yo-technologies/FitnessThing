package handlers

import (
	"context"
	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/shared/domain/dto"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"
	"fmt"

	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) UpdateSetLog(ctx context.Context, in *desc.UpdateSetLogRequest) (*desc.SetLogResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.UpdateSetLog")
	defer span.Finish()

	if err := in.ValidateAll(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("error getting user id from context")
		return nil, domain.ErrInternal
	}

	workoutID, err := domain.ParseID(in.GetWorkoutId())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	exerciseLogID, err := domain.ParseID(in.GetExerciseLogId())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	setLogID, err := domain.ParseID(in.GetSetId())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	var updateSetLogDTO dto.UpdateSetLogDTO
	{
		if in.Reps != nil {
			updateSetLogDTO.Reps = utils.NewNullable(int(*in.Reps), in.Reps != nil)
		}
		if in.Weight != nil {
			updateSetLogDTO.Weight = utils.NewNullable(float32(*in.Weight), in.Weight != nil)
		}
	}

	setLog, err := i.service.UpdateSetLog(ctx, userID, workoutID, exerciseLogID, setLogID, updateSetLogDTO)
	if err != nil {
		return nil, err
	}

	return &desc.SetLogResponse{
		SetLog: mappers.SetLogToProto(setLog),
	}, nil
}
