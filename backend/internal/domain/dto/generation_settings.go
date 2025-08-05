package dto

import (
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/utils"
)

type CreateGenerationSettings struct {
	BasePrompt              utils.Nullable[string]
	VarietyLevel            utils.Nullable[int]
	PrimaryGoal             utils.Nullable[domain.Goal]
	SecondaryGoals          []string
	ExperienceLevel         utils.Nullable[domain.ExperienceLevel]
	DaysPerWeek             utils.Nullable[int]
	SessionDurationMinutes  utils.Nullable[int]
	Injuries                utils.Nullable[string]
	PriorityMuscleGroupsIDs []domain.ID
	WorkoutPlanType         utils.Nullable[domain.WorkoutPlanType]
}
