package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) SaveGenerationSettings(ctx context.Context, userID domain.ID, createDTO dto.CreateGenerationSettings) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.SaveGenerationSettings")
	defer span.Finish()

	settings, err := s.repository.GetGenerationSettings(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return domain.GenerationSettings{}, fmt.Errorf("failed to get generation settings: %w", err)
	}

	if errors.Is(err, domain.ErrNotFound) {
		settings = domain.NewGenerationSettings(userID)
	}

	{
		if createDTO.BasePrompt.IsValid {
			settings.BasePrompt = createDTO.BasePrompt
		}
		if createDTO.VarietyLevel.IsValid {
			settings.VarietyLevel = createDTO.VarietyLevel
		}
		if createDTO.PrimaryGoal.IsValid {
			settings.PrimaryGoal = createDTO.PrimaryGoal.V
		}
		if createDTO.SecondaryGoals != nil {
			settings.SecondaryGoals = createDTO.SecondaryGoals
		}
		if createDTO.ExperienceLevel.IsValid {
			settings.ExperienceLevel = createDTO.ExperienceLevel.V
		}
		if createDTO.DaysPerWeek.IsValid {
			settings.DaysPerWeek = createDTO.DaysPerWeek
		}
		if createDTO.SessionDurationMinutes.IsValid {
			settings.SessionDurationMinutes = createDTO.SessionDurationMinutes
		}
		if createDTO.Injuries.IsValid {
			settings.Injuries = createDTO.Injuries
		}
		if createDTO.PriorityMuscleGroupsIDs != nil {
			settings.PriorityMuscleGroupsIDs = createDTO.PriorityMuscleGroupsIDs
		}
		if createDTO.WorkoutPlanType.IsValid {
			settings.WorkoutPlanType = createDTO.WorkoutPlanType.V
		}
	}

	settings.Hash, err = hashGenerationSettings(settings)
	if err != nil {
		return domain.GenerationSettings{}, fmt.Errorf("failed to hash generation settings: %w", err)
	}

	if err := s.unitOfWork.InTransaction(ctx, func(ctx context.Context) (err error) {
		settings, err = s.repository.CreateOrUpdateGenerationSettings(ctx, settings)
		if err != nil {
			return fmt.Errorf("failed to create or update generation settings: %w", err)
		}

		return nil
	}); err != nil {
		return domain.GenerationSettings{}, fmt.Errorf("failed to save generation settings: %w", err)
	}

	return settings, nil
}

func (s *Service) GetGenerationSettings(ctx context.Context, userID domain.ID) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetGenerationSettings")
	defer span.Finish()

	settings, err := s.repository.GetGenerationSettings(ctx, userID)
	if err != nil {
		return domain.GenerationSettings{}, fmt.Errorf("failed to get generation settings: %w", err)
	}

	return settings, nil
}

func hashGenerationSettings(settings domain.GenerationSettings) (string, error) {
	bytes, err := json.Marshal(settings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal generation settings: %w", err)
	}

	hash := md5.Sum(bytes)
	return fmt.Sprintf("%x", hash), nil
}
