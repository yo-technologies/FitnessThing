package dto

import (
	"fitness-trainer/internal/utils"
)

type CreateSetLogDTO struct {
	Reps   int
	Weight float32
}

type UpdateSetLogDTO struct {
	Reps   utils.Nullable[int]
	Weight utils.Nullable[float32]
}
