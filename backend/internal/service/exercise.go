package service

import (
	"context"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetExercises(ctx context.Context, muscleGroups, excludedExercises []domain.ID) ([]domain.Exercise, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetExercises")
	defer span.Finish()

	return s.repository.GetExercises(ctx, muscleGroups, excludedExercises)
}

func (s *Service) GetExerciseByID(ctx context.Context, id domain.ID) (domain.Exercise, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetExerciseByID")
	defer span.Finish()

	return s.repository.GetExerciseByID(ctx, id)
}

func (s *Service) GetExerciseAlternatives(ctx context.Context, id domain.ID) ([]domain.Exercise, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetExerciseAlternatives")
	defer span.Finish()

	exercise, err := s.GetExerciseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ids := make([]domain.ID, 0, len(exercise.TargetMuscleGroups))
	for _, muscleGroup := range exercise.TargetMuscleGroups {
		mg, err := s.repository.GetMuscleGroupByName(ctx, muscleGroup.String())
		if err != nil {
			return nil, err
		}
		ids = append(ids, mg.ID)
	}

	result, err := s.repository.GetExercises(ctx, ids, []domain.ID{id})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) GetExerciseHistory(ctx context.Context, userID, exerciseID domain.ID, offset, limit int) ([]dto.ExerciseLogDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetExerciseHistory")
	defer span.Finish()

	exerciseLogs, err := s.repository.GetExerciseLogsByExerciseIDAndUserID(ctx, exerciseID, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	exerciseLogDTOs := make([]dto.ExerciseLogDTO, 0, len(exerciseLogs))
	for _, exerciseLog := range exerciseLogs {
		exerciseLogDTO, err := s.GetExerciseLog(ctx, userID, exerciseLog.ID)
		if err != nil {
			return nil, err
		}

		exerciseLogDTOs = append(exerciseLogDTOs, exerciseLogDTO)
	}

	return exerciseLogDTOs, nil
}

func (s *Service) CreateExercise(ctx context.Context, exerciseDTO dto.CreateExerciseDTO) (domain.Exercise, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.CreateExercise")
	defer span.Finish()

	exercise := domain.NewExercise(
		exerciseDTO.Name,
		exerciseDTO.Description.V,
		exerciseDTO.VideoURL.V,
		[]domain.MuscleGroup{},
	)

	err := s.unitOfWork.InTransaction(ctx, func(ctx context.Context) (err error) {
		exercise, err = s.repository.CreateExercise(ctx, exercise, exerciseDTO.TargetMuscleGroups)
		return err
	})
	if err != nil {
		return domain.Exercise{}, err
	}

	return exercise, nil
}
