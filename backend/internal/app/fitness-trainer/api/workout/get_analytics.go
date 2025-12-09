package workout

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *Implementation) GetAnalytics(ctx context.Context, in *desc.GetAnalyticsRequest) (*desc.GetAnalyticsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.GetAnalytics")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("user id not found in context")
		return nil, domain.ErrInternal
	}

	logger.Debugf("GetAnalytics request: muscle_group=%v, exercise_id=%v", in.MuscleGroup, in.ExerciseId)
	logger.Debugf("from=%v, to=%v, user:%v", in.From, in.To, userID)

	var muscleGroupIDs []domain.ID
	if in.MuscleGroup != nil && *in.MuscleGroup != "" {
		logger.Debugf("Parsing muscle group ID: %s", *in.MuscleGroup)
		id, err := domain.ParseID(*in.MuscleGroup)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid muscle_group: %w", domain.ErrInvalidArgument, err)
		}
		muscleGroupIDs = append(muscleGroupIDs, id)
	}

	var exerciseIDs []domain.ID
	if in.ExerciseId != nil && *in.ExerciseId != "" {
		logger.Debugf("Parsing exercise ID: %s", *in.ExerciseId)
		id, err := domain.ParseID(*in.ExerciseId)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid exercise_id: %w", domain.ErrInvalidArgument, err)
		}
		exerciseIDs = append(exerciseIDs, id)
	}

	from := in.From.AsTime()
	to := in.To.AsTime()

	series, err := i.service.GetAnalytics(ctx, userID, from, to, muscleGroupIDs, exerciseIDs, in.SplitByExercise)
	if err != nil {
		return nil, err
	}

	var protoSeries []*desc.AnalyticsSeries
	for _, s := range series {
		var points []*desc.AnalyticsPoint
		for _, p := range s.Points {
			points = append(points, &desc.AnalyticsPoint{
				Date:  timestamppb.New(p.Date),
				Value: p.Value,
			})
		}
		protoSeries = append(protoSeries, &desc.AnalyticsSeries{
			Name:   s.Name,
			Points: points,
		})
	}

	return &desc.GetAnalyticsResponse{
		Series: protoSeries,
	}, nil
}
