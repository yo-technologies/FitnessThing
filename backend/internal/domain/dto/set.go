package dto

import (
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/utils"
	"time"
)

type CreateSetDTO struct {
	ExerciseInstanceID domain.ID
	SetType            domain.SetType
	Reps               utils.Nullable[int]
	Weight             utils.Nullable[float32]
	Time               utils.Nullable[time.Duration]
}

type UpdateSetDTO struct {
	Reps   utils.Nullable[int]
	Weight utils.Nullable[float32]
	Time   utils.Nullable[time.Duration]
}

type ExpectedSetInput struct {
	SetType domain.SetType
	Reps    int
	Weight  float32
	Time    time.Duration
}
