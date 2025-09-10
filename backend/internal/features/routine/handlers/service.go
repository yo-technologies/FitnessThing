package handlers

import (
	"context"

	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/shared/domain/dto"
	desc "fitness-trainer/pkg/workouts"
)

type Service interface {
	GetRoutines(ctx context.Context, userID domain.ID) ([]domain.Routine, error)
	CreateRoutine(ctx context.Context, dto dto.CreateRoutineDTO) (domain.Routine, error)
	GetRoutineByID(ctx context.Context, id domain.ID) (dto.RoutineDetailsDTO, error)
	UpdateRoutine(ctx context.Context, id domain.ID, dto dto.UpdateRoutineDTO) (domain.Routine, error)
	DeleteRoutine(ctx context.Context, id domain.ID) error

	AddExerciseToRoutine(ctx context.Context, routineID domain.ID, exerciseID domain.ID) (domain.ExerciseInstance, error)
	GetExerciseInstance(ctx context.Context, userID, routineID, exerciseInstanceID domain.ID) (dto.ExerciseInstanceDetailsDTO, error)
	RemoveExerciseInstanceFromRoutine(ctx context.Context, userID, routineID, exerciseInstanceID domain.ID) error
	SetExerciseOrder(ctx context.Context, userID, routineID domain.ID, exerciseInstanceIDs []domain.ID) error

	AddSetToExerciseInstance(ctx context.Context, userID, routineID, exerciseInstanceID domain.ID, dto dto.CreateSetDTO) (domain.Set, error)
	RemoveSetFromExerciseInstance(ctx context.Context, userID, routineID, exerciseInstanceID, setID domain.ID) error
	UpdateSetInExerciseInstance(ctx context.Context, userID, routineID, exerciseInstanceID, setID domain.ID, dto dto.UpdateSetDTO) (domain.Set, error)
}

type Implementation struct {
	service Service
	desc.UnimplementedRoutineServiceServer
}

func New(service Service) *Implementation {
	return &Implementation{
		service: service,
	}
}
