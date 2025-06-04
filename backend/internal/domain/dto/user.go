package dto

import (
	"fitness-trainer/internal/utils"
	"time"
)

type CreateUserDTO struct {
	TelegramID       int64
	TelegramUsername utils.Nullable[string]
	FirstName        utils.Nullable[string]
	LastName         utils.Nullable[string]
	ProfilePicURL    utils.Nullable[string]
}

type UpdateUserDTO struct {
	Height      utils.Nullable[float32]
	Weight      utils.Nullable[float32]
	DateOfBirth time.Time
}
