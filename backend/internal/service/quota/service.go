package quota

import (
	"context"
	"time"

	"fitness-trainer/internal/config"
	"fitness-trainer/internal/domain"
)

// TokenLimitProvider allows customizing per-user token limits.
// If nil, default from config is used.
type TokenLimitProvider func(ctx context.Context, userID domain.ID) int

type repo interface {
	ReserveLLMTokens(ctx context.Context, userID domain.ID, day time.Time, n int, dailyLimit int) (bool, error)
	ConfirmLLMTokenUsage(ctx context.Context, userID domain.ID, day time.Time, reserved int, actual int) error
	GetLLMDailyUsage(ctx context.Context, userID domain.ID, day time.Time) (used int, reserved int, err error)
}

type Service struct {
	repo      repo
	limitProv TokenLimitProvider
}

func New(r repo, cfg *config.Config, lp TokenLimitProvider) *Service {
	return &Service{repo: r, limitProv: lp}
}

func (s *Service) today() time.Time {
	now := time.Now().UTC()
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Reserve reserves n tokens if within daily limit. Returns true if allowed.
func (s *Service) Reserve(ctx context.Context, userID domain.ID, n int) (bool, error) {
	day := s.today()
	limit := s.limitProv(ctx, userID)
	return s.repo.ReserveLLMTokens(ctx, userID, day, n, limit)
}

// Confirm adjusts reserved and increments used by actual tokens consumed.
// If actual < reserved, the remainder is released. We don't enforce limit here strictly
// to avoid failing post-consumption; enforcement happens at Reserve time.
func (s *Service) Confirm(ctx context.Context, userID domain.ID, reserved int, actual int) error {
	day := s.today()
	return s.repo.ConfirmLLMTokenUsage(ctx, userID, day, reserved, actual)
}

// GetLLMDailyUsage возвращает использованные и зарезервированные токены на выбранный день
func (s *Service) GetLLMDailyUsage(ctx context.Context, userID domain.ID, day time.Time) (used int, reserved int, err error) {
	// нормализуем день к полуночи UTC
	d := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	return s.repo.GetLLMDailyUsage(ctx, userID, d)
}

// DailyLimit сообщает актуальный дневной лимит токенов для пользователя
func (s *Service) DailyLimit(ctx context.Context, userID domain.ID) int {
	return s.limitProv(ctx, userID)
}
