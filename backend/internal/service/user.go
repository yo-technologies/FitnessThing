package service

import (
	"context"
	"time"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetOrCreateUser(ctx context.Context, dto dto.CreateUserDTO) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetOrCreateUser")
	defer span.Finish()

	user := domain.NewUser(
		dto.TelegramID,
		dto.TelegramUsername,
		dto.FirstName,
		dto.LastName,
		dto.ProfilePicURL,
	)

	err := s.unitOfWork.InTransaction(ctx, func(ctx context.Context) error {
		var err error
		user, err = s.repository.GetOrCreateUser(ctx, user)
		return err
	})
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *Service) GetUserByID(ctx context.Context, id domain.ID) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetUserByID")
	defer span.Finish()

	return s.repository.GetUserByID(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, id domain.ID, dto dto.UpdateUserDTO) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.UpdateUser")
	defer span.Finish()

	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	{
		if !dto.DateOfBirth.IsZero() {
			user.DateOfBirth = dto.DateOfBirth
		}

		if dto.Height.IsValid {
			user.Height = dto.Height
		}

		if dto.Weight.IsValid {
			user.Weight = dto.Weight
		}

		user.UpdatedAt = time.Now()
	}

	err = s.unitOfWork.InTransaction(ctx, func(ctx context.Context) error {
		user, err = s.repository.UpdateUser(ctx, user)

		return err
	})
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}
