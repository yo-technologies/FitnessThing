package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"fitness-trainer/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) SaveUserFact(ctx context.Context, userID domain.ID, content string) (domain.UserFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.SaveUserFact")
	defer span.Finish()

	factText := strings.TrimSpace(content)
	if factText == "" {
		return domain.UserFact{}, fmt.Errorf("%w: fact content cannot be empty", domain.ErrInvalidArgument)
	}

	if utf8.RuneCountInString(factText) > domain.MaxUserFactLength {
		return domain.UserFact{}, fmt.Errorf("%w: fact content exceeds %d characters", domain.ErrInvalidArgument, domain.MaxUserFactLength)
	}

	fact := domain.NewUserFact(userID, factText)
	var created domain.UserFact

	err := s.unitOfWork.InTransaction(ctx, func(ctx context.Context) error {
		count, err := s.repository.CountUserFactsByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if count >= domain.MaxUserFactsPerUser {
			return fmt.Errorf("%w: user facts limit reached (%d)", domain.ErrInvalidArgument, domain.MaxUserFactsPerUser)
		}

		created, err = s.repository.CreateUserFact(ctx, fact)
		return err
	})
	if err != nil {
		return domain.UserFact{}, err
	}

	return created, nil
}

func (s *Service) ListUserFacts(ctx context.Context, userID domain.ID) ([]domain.UserFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.ListUserFacts")
	defer span.Finish()

	return s.repository.ListUserFacts(ctx, userID, domain.MaxUserFactsPerUser)
}

func (s *Service) DeleteUserFact(ctx context.Context, userID, factID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.DeleteUserFact")
	defer span.Finish()

	return s.repository.DeleteUserFact(ctx, userID, factID)
}
