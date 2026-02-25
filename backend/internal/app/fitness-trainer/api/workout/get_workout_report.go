package workout

import (
	"context"
	"fitness-trainer/internal/domain"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) GetWorkoutReport(ctx context.Context, in *desc.GetWorkoutReportRequest) (*desc.WorkoutReportResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.workout.GetWorkoutReport")
	defer span.Finish()

	if err := in.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
