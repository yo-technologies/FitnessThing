package service

import (
	"context"
	"errors"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) SaveGenerationSettings(ctx context.Context, userID domain.ID, createDTO dto.CreateGenerationSettings) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.SaveGenerationSettings")
	defer span.Finish()

	settings := domain.NewGenerationSettings(userID)
	{
		settings.BasePrompt = createDTO.BasePrompt
		settings.VarietyLevel = createDTO.VarietyLevel
		settings.PrimaryGoal = createDTO.PrimaryGoal
		settings.SecondaryGoals = createDTO.SecondaryGoals
		settings.ExperienceLevel = createDTO.ExperienceLevel
		settings.DaysPerWeek = createDTO.DaysPerWeek
		settings.SessionDurationMinutes = createDTO.SessionDurationMinutes
		settings.Injuries = createDTO.Injuries
		settings.PriorityMuscleGroupsIDs = createDTO.PriorityMuscleGroupsIDs
		settings.WorkoutPlanType = createDTO.WorkoutPlanType
	}

	if err := s.unitOfWork.InTransaction(ctx, func(ctx context.Context) (err error) {
		settings, err = s.repository.CreateOrUpdateGenerationSettings(ctx, settings)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return domain.GenerationSettings{}, err
	}

	return settings, nil
}

func (s *Service) GetGenerationSettings(ctx context.Context, userID domain.ID) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetGenerationSettings")
	defer span.Finish()

	settings, err := s.repository.GetGenerationSettings(ctx, userID)

	// If the settings are found, return them
	if err == nil {
		return settings, nil
	}

	// If error is not not found, return the error
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.GenerationSettings{}, err
	}

	// Create and save new settings
	settings = domain.NewGenerationSettings(userID)
	settings, err = s.repository.CreateOrUpdateGenerationSettings(ctx, settings)
	if err != nil {
		return domain.GenerationSettings{}, err
	}

	return settings, nil
}
