package service

import (
	"context"
	"fitness-trainer/internal/domain/dto"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetMuscleGroups(ctx context.Context) ([]dto.MuscleGroupDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetMuscleGroups")
	defer span.Finish()

	return s.repository.GetMuscleGroups(ctx)
}
