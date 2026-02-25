package dto

import "fitness-trainer/internal/domain"

type WorkoutDTO struct {
	Workout      domain.Workout
	ExerciseLogs []domain.ExerciseLog
}
