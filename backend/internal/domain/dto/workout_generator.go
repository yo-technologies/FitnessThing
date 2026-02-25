package dto

import (
	"fitness-trainer/internal/domain"
	"time"
)

type SlimWorkoutDTO struct {
	ID            domain.ID
	CreatedAt     time.Time
	ExerciseNames []string
}

type SlimExerciseDTO struct {
	ID                 domain.ID
	Name               string
	TargetMuscleGroups []domain.MuscleGroup
}

type GeneratedWorkoutDTO struct {
	ExerciseIDs []domain.ID
	Reasoning   string
}
