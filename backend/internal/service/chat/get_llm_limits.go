package chat

import (
	"context"
	"fitness-trainer/internal/domain"
	"time"

	"github.com/opentracing/opentracing-go"
)

// GetLLMLimits возвращает агрегированные лимиты LLM на текущий день
func (s *Service) GetLLMLimits(ctx context.Context, userID domain.ID) (domain.LLMLimits, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetLLMLimits")
	defer span.Finish()

	now := time.Now().UTC()
	used, reserved, err := s.quotaService.GetLLMDailyUsage(ctx, userID, now)
	if err != nil {
		return domain.LLMLimits{}, err
	}

	daily := s.quotaService.DailyLimit(ctx, userID)

	return domain.NewLLMLimits(daily, used, reserved), nil
}
