package service

import (
	"context"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/logger"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetRoutines(ctx context.Context, userID domain.ID) ([]domain.Routine, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetRoutines")
	defer span.Finish()

	return s.repository.GetRoutines(ctx, userID)
}

func (s *Service) CreateRoutine(ctx context.Context, dto dto.CreateRoutineDTO) (domain.Routine, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.CreateRoutine")
	defer span.Finish()

	ctx, err := s.unitOfWork.Begin(ctx)
	if err != nil {
		return domain.Routine{}, err
	}
	defer s.unitOfWork.Rollback(ctx)

	routine := domain.NewRoutine(dto.UserID, dto.Name, dto.Description)

	routine, err = s.repository.CreateRoutine(ctx, routine)
	if err != nil {
		return domain.Routine{}, err
	}

	if dto.WorkoutID.IsValid {
		err = s.fillRoutineWithWorkout(ctx, routine, dto.WorkoutID.V)
		if err != nil {
			return domain.Routine{}, err
		}
	}

	if err := s.unitOfWork.Commit(ctx); err != nil {
		return domain.Routine{}, err
	}

	return routine, nil
}

func (s *Service) fillRoutineWithWorkout(ctx context.Context, routine domain.Routine, workoutID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.fillRoutineWithWorkout")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return err
	}

	if routine.UserID != workout.UserID {
		return err
	}

	exerciseLogs, err := s.repository.GetExerciseLogsByWorkoutID(ctx, workoutID)
	if err != nil {
		return err
	}

	for _, exerciseLog := range exerciseLogs {
		exerciseInstance := domain.NewExerciseInstance(routine.ID, exerciseLog.ExerciseID)
		exerciseInstance, err := s.repository.CreateExerciseInstance(ctx, exerciseInstance)
		if err != nil {
			return err
		}

		sets, err := s.repository.GetSetLogsByExerciseLogID(ctx, exerciseLog.ID)
		if err != nil {
			return err
		}

		for _, set := range sets {
			newSet := domain.NewSet(
				exerciseInstance.ID,
				domain.SetTypeReps,
				set.Reps,
				set.Weight,
				set.Time,
			)
			if _, err := s.repository.CreateSet(ctx, newSet); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) GetRoutineByID(ctx context.Context, id domain.ID) (dto.RoutineDetailsDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetRoutineByID")
	defer span.Finish()

	routine, err := s.repository.GetRoutineByID(ctx, id)
	if err != nil {
		return dto.RoutineDetailsDTO{}, err
	}

	exerciseInstances, err := s.repository.GetExerciseInstancesByRoutineID(ctx, id)
	if err != nil {
		return dto.RoutineDetailsDTO{}, err
	}

	result := dto.RoutineDetailsDTO{
		ID:                routine.ID,
		UserID:            routine.UserID,
		Name:              routine.Name,
		Description:       routine.Description,
		CreatedAt:         routine.CreatedAt,
		UpdatedAt:         routine.UpdatedAt,
		ExerciseInstances: make([]dto.ExerciseInstanceDetailsDTO, len(exerciseInstances)),
	}

	for i, instance := range exerciseInstances {
		exercise, err := s.repository.GetExerciseByID(ctx, instance.ExerciseID)
		if err != nil {
			return dto.RoutineDetailsDTO{}, err
		}

		sets, err := s.repository.GetSetsByExerciseInstanceID(ctx, instance.ID)
		if err != nil {
			return dto.RoutineDetailsDTO{}, err
		}

		result.ExerciseInstances[i] = dto.ExerciseInstanceDetailsDTO{
			ID:         instance.ID,
			RoutineID:  instance.RoutineID,
			ExerciseID: instance.ExerciseID,
			CreatedAt:  instance.CreatedAt,
			UpdatedAt:  instance.UpdatedAt,
			Exercise:   exercise,
			Sets:       sets,
		}
	}

	return result, nil
}

func (s *Service) AddExerciseToRoutine(ctx context.Context, routineID domain.ID, exerciseID domain.ID) (domain.ExerciseInstance, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.AddExerciseToRoutine")
	defer span.Finish()

	exerciseInstance := domain.NewExerciseInstance(routineID, exerciseID)
	return s.repository.CreateExerciseInstance(ctx, exerciseInstance)
}

func (s *Service) GetExerciseInstance(ctx context.Context, userID, routineID, exerciseInstanceID domain.ID) (dto.ExerciseInstanceDetailsDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetExerciseInstance")
	defer span.Finish()

	exerciseInstance, err := s.repository.GetExerciseInstanceByID(ctx, exerciseInstanceID)
	if err != nil {
		return dto.ExerciseInstanceDetailsDTO{}, err
	}

	if exerciseInstance.RoutineID != routineID {
		logger.Errorf("exercise instance %s does not belong to routine %s", exerciseInstanceID, routineID)
		return dto.ExerciseInstanceDetailsDTO{}, domain.ErrNotFound
	}

	exercise, err := s.repository.GetExerciseByID(ctx, exerciseInstance.ExerciseID)
	if err != nil {
		return dto.ExerciseInstanceDetailsDTO{}, err
	}

	sets, err := s.repository.GetSetsByExerciseInstanceID(ctx, exerciseInstanceID)
	if err != nil {
		return dto.ExerciseInstanceDetailsDTO{}, err
	}

	return dto.ExerciseInstanceDetailsDTO{
		ID:         exerciseInstance.ID,
		RoutineID:  exerciseInstance.RoutineID,
		ExerciseID: exerciseInstance.ExerciseID,
		CreatedAt:  exerciseInstance.CreatedAt,
		UpdatedAt:  exerciseInstance.UpdatedAt,
		Exercise:   exercise,
		Sets:       sets,
	}, nil
}

func (s *Service) RemoveExerciseInstanceFromRoutine(ctx context.Context, userID domain.ID, routineID domain.ID, exerciseInstanceID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.RemoveExerciseInstanceFromRoutine")
	defer span.Finish()

	routine, err := s.repository.GetRoutineByID(ctx, routineID)
	if err != nil {
		return err
	}

	if routine.UserID != userID {
		return domain.ErrUnauthorized
	}

	return s.repository.DeleteExerciseInstance(ctx, exerciseInstanceID)
}

func (s *Service) DeleteRoutine(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.DeleteRoutine")
	defer span.Finish()

	return s.repository.DeleteRoutine(ctx, id)
}

func (s *Service) UpdateRoutine(ctx context.Context, id domain.ID, dto dto.UpdateRoutineDTO) (domain.Routine, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.UpdateRoutine")
	defer span.Finish()

	routine, err := s.repository.GetRoutineByID(ctx, id)
	if err != nil {
		return domain.Routine{}, err
	}

	if dto.Name.IsValid {
		routine.Name = dto.Name.V
	}

	if dto.Description.IsValid {
		routine.Description = dto.Description.V
	}

	return s.repository.UpdateRoutine(ctx, id, routine)
}

func (s *Service) AddSetToExerciseInstance(ctx context.Context, userID, routineID, exerciseInstanceID domain.ID, dto dto.CreateSetDTO) (domain.Set, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.AddSetToExerciseInstance")
	defer span.Finish()

	set := domain.NewSet(exerciseInstanceID, dto.SetType, dto.Reps.V, dto.Weight.V, dto.Time.V)
	return s.repository.CreateSet(ctx, set)
}

func (s *Service) RemoveSetFromExerciseInstance(ctx context.Context, userID, routineID, exerciseInstanceID, setID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.RemoveSetFromExerciseInstance")
	defer span.Finish()

	routine, err := s.repository.GetRoutineByID(ctx, routineID)
	if err != nil {
		return err
	}

	if routine.UserID != userID {
		logger.Errorf("user %s is not authorized to remove set from exercise instance %s", userID, exerciseInstanceID)
		return domain.ErrNotFound
	}

	exerciseInstance, err := s.repository.GetExerciseInstanceByID(ctx, exerciseInstanceID)
	if err != nil {
		return err
	}

	if exerciseInstance.RoutineID != routineID {
		logger.Errorf("exercise instance %s does not belong to routine %s", exerciseInstanceID, routineID)
		return domain.ErrNotFound
	}

	set, err := s.repository.GetSetByID(ctx, setID)
	if err != nil {
		return err
	}

	if set.ExerciseInstanceID != exerciseInstanceID {
		logger.Errorf("set %s does not belong to exercise instance %s", setID, exerciseInstanceID)
		return domain.ErrNotFound
	}

	return s.repository.DeleteSet(ctx, setID)
}

func (s *Service) UpdateSetInExerciseInstance(ctx context.Context, userID, routineID, exerciseInstanceID, setID domain.ID, dto dto.UpdateSetDTO) (domain.Set, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.UpdateSet")
	defer span.Finish()

	routine, err := s.repository.GetRoutineByID(ctx, routineID)
	if err != nil {
		return domain.Set{}, err
	}

	if routine.UserID != userID {
		logger.Errorf("user %s is not authorized to update set %s", userID, setID)
		return domain.Set{}, domain.ErrUnauthorized
	}

	exerciseInstance, err := s.repository.GetExerciseInstanceByID(ctx, exerciseInstanceID)
	if err != nil {
		return domain.Set{}, err
	}

	if exerciseInstance.RoutineID != routineID {
		logger.Errorf("exercise instance %s does not belong to routine %s", exerciseInstanceID, routineID)
		return domain.Set{}, domain.ErrNotFound
	}

	set, err := s.repository.GetSetByID(ctx, setID)
	if err != nil {
		return domain.Set{}, err
	}

	if set.ExerciseInstanceID != exerciseInstanceID {
		logger.Errorf("set %s does not belong to exercise instance %s", setID, exerciseInstanceID)
		return domain.Set{}, domain.ErrNotFound
	}

	if dto.Reps.IsValid {
		set.Reps = dto.Reps.V
	}

	if dto.Weight.IsValid {
		set.Weight = dto.Weight.V
	}

	if dto.Time.IsValid {
		set.Time = dto.Time.V
	}

	return s.repository.UpdateSet(ctx, setID, set)
}

func (s *Service) SetExerciseOrder(ctx context.Context, userID, routineID domain.ID, exerciseInstanceIDs []domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.SetExerciseOrder")
	defer span.Finish()

	routine, err := s.repository.GetRoutineByID(ctx, routineID)
	if err != nil {
		return err
	}

	if routine.UserID != userID {
		logger.Errorf("user %s is not authorized to update exercise order", userID)
		return domain.ErrUnauthorized
	}

	// TODO: check if all exercise instances belong to the routine

	return s.repository.SetExerciseOrder(ctx, routineID, exerciseInstanceIDs)
}
