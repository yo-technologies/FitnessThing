package dto

import "fitness-trainer/internal/domain"

type GenerateWorkoutOptions struct {
	UserID     domain.ID
	Workouts   []SlimWorkoutDTO
	Exercises  []SlimExerciseDTO
	UserPrompt string
	Settings   domain.GenerationSettings
}
