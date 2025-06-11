package dto

import (
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/utils"
)

type CreateGenerationSettings struct {
	BasePrompt              utils.Nullable[string]
	VarietyLevel            utils.Nullable[int]
	PrimaryGoal             domain.Goal
	SecondaryGoals          []string
	ExperienceLevel         domain.ExperienceLevel
	DaysPerWeek             utils.Nullable[int]
	SessionDurationMinutes  utils.Nullable[int]
	Injuries                utils.Nullable[string]
	PriorityMuscleGroupsIDs []domain.ID
	WorkoutPlanType         domain.WorkoutPlanType
}
