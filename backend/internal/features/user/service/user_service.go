package service

import (
	"context"
	"time"

	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/shared/domain/dto"

	"github.com/opentracing/opentracing-go"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id domain.ID) (domain.User, error)
	GetOrCreateUser(ctx context.Context, user domain.User) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
}

type GenerationSettingsRepository interface {
	CreateOrUpdateGenerationSettings(ctx context.Context, settings domain.GenerationSettings) (domain.GenerationSettings, error)
	GetGenerationSettings(ctx context.Context, userID domain.ID) (domain.GenerationSettings, error)
}

type UnitOfWork interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	InTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

type UserService struct {
	userRepo               UserRepository
	generationSettingsRepo GenerationSettingsRepository
	unitOfWork             UnitOfWork
}

func NewUserService(
	userRepo UserRepository,
	generationSettingsRepo GenerationSettingsRepository,
	unitOfWork UnitOfWork,
) *UserService {
	return &UserService{
		userRepo:               userRepo,
		generationSettingsRepo: generationSettingsRepo,
		unitOfWork:             unitOfWork,
	}
}

func (s *UserService) GetOrCreateUser(ctx context.Context, dto dto.CreateUserDTO) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserService.GetOrCreateUser")
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
		user, err = s.userRepo.GetOrCreateUser(ctx, user)
		return err
	})
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id domain.ID) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserService.GetUserByID")
	defer span.Finish()

	return s.userRepo.GetUserByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id domain.ID, dto dto.UpdateUserDTO) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserService.UpdateUser")
	defer span.Finish()

	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	if dto.FirstName.IsValid {
		user.FirstName = dto.FirstName
	}

	if dto.LastName.IsValid {
		user.LastName = dto.LastName
	}

	if dto.PreferredLanguage.IsValid {
		user.PreferredLanguage = dto.PreferredLanguage
	}

	if dto.ExperienceLevel.IsValid {
		user.ExperienceLevel = dto.ExperienceLevel.V
	}

	if dto.WeightUnit.IsValid {
		user.WeightUnit = dto.WeightUnit.V
	}

	if dto.ProfilePicURL.IsValid {
		user.ProfilePicURL = dto.ProfilePicURL
	}

	user.UpdatedAt = time.Now()

	err = s.unitOfWork.InTransaction(ctx, func(ctx context.Context) error {
		var err error
		user, err = s.userRepo.UpdateUser(ctx, user)
		return err
	})
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *UserService) GetGenerationSettings(ctx context.Context, userID domain.ID) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserService.GetGenerationSettings")
	defer span.Finish()

	return s.generationSettingsRepo.GetGenerationSettings(ctx, userID)
}

func (s *UserService) SaveGenerationSettings(ctx context.Context, userID domain.ID, createDTO dto.CreateGenerationSettings) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserService.SaveGenerationSettings")
	defer span.Finish()

	settings := domain.NewGenerationSettings(
		userID,
		createDTO.Goals,
		createDTO.AvailableTime,
		createDTO.ExperienceLevel,
		createDTO.PreferredExercises,
		createDTO.AvoidedExercises,
		createDTO.Equipment,
		createDTO.InjuryConsiderations,
		createDTO.SpecialRequests,
	)

	var err error
	err = s.unitOfWork.InTransaction(ctx, func(ctx context.Context) error {
		settings, err = s.generationSettingsRepo.CreateOrUpdateGenerationSettings(ctx, settings)
		return err
	})
	if err != nil {
		return domain.GenerationSettings{}, err
	}

	return settings, nil
}