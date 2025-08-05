package user

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/domain"

	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) UpdateWorkoutGenerationSettings(ctx context.Context, req *desc.UpdateWorkoutGenerationSettingsRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.user.UpdateWorkoutGenerationSettings")
	defer span.Finish()

	if err := req.ValidateAll(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	id, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user id not found in context: %w", domain.ErrUnauthorized)
	}

	createDTO, err := mappers.UpdateGenerationSettingsRequestToCreateDTO(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	if _, err := i.service.SaveGenerationSettings(ctx, id, createDTO); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
