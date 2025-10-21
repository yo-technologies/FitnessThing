package service

import (
	"context"
	"fmt"
	"time"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/logger"

	"github.com/opentracing/opentracing-go"
)

const workoutsCount = 8

func (s *Service) StartWorkout(ctx context.Context, userID domain.ID, opts domain.StartWorkoutOpts) (domain.Workout, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.StartWorkout")
	defer span.Finish()

	txCtx, err := s.unitOfWork.Begin(ctx)
	if err != nil {
		return domain.Workout{}, err
	}
	defer s.unitOfWork.Rollback(txCtx)

	if opts.RoutineID.IsValid {
		_, err := s.repository.GetRoutineByID(txCtx, opts.RoutineID.V)
		if err != nil {
			return domain.Workout{}, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
		}
	}

	workout := domain.NewWorkout(userID, opts.RoutineID)

	workout, err = s.repository.CreateWorkout(txCtx, workout)
	if err != nil {
		return domain.Workout{}, err
	}

	if opts.RoutineID.IsValid {
		err = s.enrichWorkoutFromRoutine(txCtx, userID, workout.ID, opts.RoutineID.V)
		if err != nil {
			return domain.Workout{}, err
		}
	}

	err = s.unitOfWork.Commit(txCtx)
	if err != nil {
		return domain.Workout{}, err
	}

	return workout, nil
}

func (s *Service) enrichWorkoutFromRoutine(ctx context.Context, userID, workoutID, routineID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.assignExercisesToWorkout")
	defer span.Finish()

	routine, err := s.repository.GetRoutineByID(ctx, routineID)
	if err != nil {
		return err
	}

	exerciseInstances, err := s.repository.GetExerciseInstancesByRoutineID(ctx, routine.ID)
	if err != nil {
		return err
	}

	for _, instance := range exerciseInstances {
		exerciseLog, err := s.LogExercise(ctx, userID, workoutID, instance.ExerciseID)
		if err != nil {
			return err
		}

		sets, err := s.repository.GetSetsByExerciseInstanceID(ctx, instance.ID)
		if err != nil {
			return err
		}

		for _, set := range sets {
			expectedSet := domain.NewExpectedSet(
				exerciseLog.ID,
				set.SetType,
				set.Reps,
				set.Weight,
				set.Time,
			)

			_, err = s.repository.CreateExpectedSet(ctx, expectedSet)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) GetWorkout(ctx context.Context, userID domain.ID, workoutID domain.ID) (dto.WorkoutDetailsDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetWorkout")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return dto.WorkoutDetailsDTO{}, err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to access workout %s", userID, workoutID)
		return dto.WorkoutDetailsDTO{}, domain.ErrNotFound
	}

	exerciseLogs, err := s.repository.GetExerciseLogsByWorkoutID(ctx, workoutID)
	if err != nil {
		return dto.WorkoutDetailsDTO{}, err
	}

	exerciseLogsDTOs := make([]dto.ExerciseLogDTO, 0, len(exerciseLogs))
	for _, exerciseLog := range exerciseLogs {
		exerciseLogDTO, err := s.GetExerciseLog(ctx, userID, exerciseLog.ID)
		if err != nil {
			return dto.WorkoutDetailsDTO{}, err
		}

		exerciseLogsDTOs = append(exerciseLogsDTOs, exerciseLogDTO)
	}

	return dto.WorkoutDetailsDTO{
		Workout:      workout,
		ExerciseLogs: exerciseLogsDTOs,
	}, nil
}

// UpdateExerciseLogWeightUnit changes the weight unit for an existing exercise log and converts existing set logs to the new unit
func (s *Service) UpdateExerciseLogWeightUnit(ctx context.Context, userID, workoutID, exerciseLogID domain.ID, unit domain.WeightUnit) (dto.ExerciseLogDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.UpdateExerciseLogWeightUnit")
	defer span.Finish()

	// ensure workout belongs to user
	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}
	if workout.UserID != userID {
		return dto.ExerciseLogDTO{}, domain.ErrForbidden
	}

	exerciseLog, err := s.repository.GetExerciseLogByID(ctx, exerciseLogID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}
	if exerciseLog.WorkoutID != workoutID {
		return dto.ExerciseLogDTO{}, domain.ErrInvalidArgument
	}

	if unit == domain.WeightUnitUnknown {
		unit = domain.WeightUnitKG
	}

	updateSets := func(txCtx context.Context) (err error) {
		sets, err := s.repository.GetSetLogsByExerciseLogID(txCtx, exerciseLog.ID)
		if err != nil {
			return err
		}

		for _, set := range sets {
			set.UpdateWeight(domain.ConvertWeight(set.Weight, exerciseLog.WeightUnit, unit))
			if _, err := s.repository.UpdateSetLog(txCtx, set.ID, set); err != nil {
				return err
			}
		}

		exerciseLog.WeightUnit = unit
		exerciseLog, err = s.repository.UpdateExerciseLog(txCtx, exerciseLog.ID, exerciseLog)
		if err != nil {
			return err
		}

		return nil
	}

	if unit != exerciseLog.WeightUnit {
		if err := s.unitOfWork.InTransaction(ctx, updateSets); err != nil {
			return dto.ExerciseLogDTO{}, err
		}
	}

	setLogs, err := s.repository.GetSetLogsByExerciseLogID(ctx, exerciseLog.ID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}
	expectedSets, err := s.repository.GetExpectedSetsByExerciseLogID(ctx, exerciseLog.ID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}
	exercise, err := s.repository.GetExerciseByID(ctx, exerciseLog.ExerciseID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}

	return dto.ExerciseLogDTO{
		ExerciseLog:  exerciseLog,
		SetLogs:      setLogs,
		ExpectedSets: expectedSets,
		Exercise:     exercise,
	}, nil
}

func (s *Service) GetExerciseLog(ctx context.Context, userID, exerciseLogID domain.ID) (dto.ExerciseLogDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetExerciseLog")
	defer span.Finish()

	exerciseLog, err := s.repository.GetExerciseLogByID(ctx, exerciseLogID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}

	workout, err := s.repository.GetWorkoutByID(ctx, exerciseLog.WorkoutID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to access exercise log %s", userID, exerciseLogID)
		return dto.ExerciseLogDTO{}, domain.ErrNotFound
	}

	setLogs, err := s.repository.GetSetLogsByExerciseLogID(ctx, exerciseLogID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}

	exercise, err := s.repository.GetExerciseByID(ctx, exerciseLog.ExerciseID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}

	expectedSets, err := s.repository.GetExpectedSetsByExerciseLogID(ctx, exerciseLogID)
	if err != nil {
		return dto.ExerciseLogDTO{}, err
	}

	return dto.ExerciseLogDTO{
		ExerciseLog:  exerciseLog,
		Exercise:     exercise,
		SetLogs:      setLogs,
		ExpectedSets: expectedSets,
	}, nil
}

func (s *Service) GetExerciseLogByID(ctx context.Context, id domain.ID) (domain.ExerciseLog, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetExerciseLogByID")
	defer span.Finish()

	return s.repository.GetExerciseLogByID(ctx, id)
}

func (s *Service) LogExercise(ctx context.Context, userID, workoutID, exerciseID domain.ID) (domain.ExerciseLog, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.LogExercise")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return domain.ExerciseLog{}, err
	}

	if !workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to log exercise for finished workout %s", userID, workoutID)
		return domain.ExerciseLog{}, fmt.Errorf("%w: workout %s is already finished", domain.ErrInvalidArgument, workoutID)
	}

	_, err = s.repository.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		return domain.ExerciseLog{}, err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to log exercise for workout %s", userID, workoutID)
		return domain.ExerciseLog{}, domain.ErrNotFound
	}

	exerciseLog := domain.NewExerciseLog(workoutID, exerciseID, domain.WeightUnitKG)

	exerciseLog, err = s.repository.CreateExerciseLog(ctx, exerciseLog)
	if err != nil {
		return domain.ExerciseLog{}, err
	}

	return exerciseLog, nil
}

func (s *Service) LogSet(ctx context.Context, userID, workoutID, exerciseLogID domain.ID, setlogDTO dto.CreateSetLogDTO) (domain.ExerciseSetLog, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.LogSet")
	defer span.Finish()

	exerciseLog, err := s.repository.GetExerciseLogByID(ctx, exerciseLogID)
	if err != nil {
		return domain.ExerciseSetLog{}, err
	}

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return domain.ExerciseSetLog{}, err
	}

	if !workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to log exercise for finished workout %s", userID, workoutID)
		return domain.ExerciseSetLog{}, fmt.Errorf("%w: workout %s is already finished", domain.ErrInvalidArgument, workoutID)
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to log set for workout %s", userID, workoutID)
		return domain.ExerciseSetLog{}, domain.ErrNotFound
	}

	if exerciseLog.WorkoutID != workoutID {
		logger.Errorf("user %s tried to log set for exercise log %s for workout %s", userID, exerciseLogID, workoutID)
		return domain.ExerciseSetLog{}, domain.ErrNotFound
	}

	setLog := domain.NewExerciseSetLog(
		exerciseLogID,
		setlogDTO.Reps,
		setlogDTO.Weight,
		time.Duration(0),
	)

	setLog, err = s.repository.CreateSetLog(ctx, setLog)
	if err != nil {
		return domain.ExerciseSetLog{}, err
	}

	return setLog, nil
}

func (s *Service) GetActiveWorkouts(ctx context.Context, userID domain.ID) ([]domain.Workout, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetActiveWorkouts")
	defer span.Finish()

	workouts, err := s.repository.GetActiveWorkouts(ctx, userID)
	if err != nil {
		return nil, err
	}

	return workouts, nil
}

func (s *Service) CompleteWorkout(ctx context.Context, userID, workoutID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.FinishWorkout")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to finish workout %s", userID, workoutID)
		return domain.ErrNotFound
	}

	if !workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to finish already finished workout %s", userID, workoutID)
		return fmt.Errorf("%w: workout %s is already finished", domain.ErrInvalidArgument, workoutID)
	}

	workout.FinishedAt = time.Now()

	_, err = s.repository.UpdateWorkout(ctx, workoutID, workout)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteWorkout(ctx context.Context, userID, workoutID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.DeleteWorkout")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to delete workout %s", userID, workoutID)
		return domain.ErrNotFound
	}

	err = s.repository.DeleteWorkout(ctx, workoutID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteExerciseLog(ctx context.Context, userID, workoutID, exerciseLogID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.DeleteExerciseLog")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to delete exercise log %s for workout %s", userID, exerciseLogID, workoutID)
		return domain.ErrNotFound
	}

	if !workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to delete exercise log %s for finished workout %s", userID, exerciseLogID, workoutID)
		return fmt.Errorf("%w: workout %s is already finished", domain.ErrInvalidArgument, workoutID)
	}

	exerciseLog, err := s.repository.GetExerciseLogByID(ctx, exerciseLogID)
	if err != nil {
		return err
	}

	if exerciseLog.WorkoutID != workoutID {
		logger.Errorf("user %s tried to delete exercise log %s for workout %s", userID, exerciseLogID, workoutID)
		return domain.ErrNotFound
	}

	err = s.repository.DeleteExerciseLog(ctx, exerciseLogID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ReplaceExpectedSets(
	ctx context.Context,
	userID, workoutID, exerciseLogID domain.ID,
	sets []dto.ExpectedSetInput,
) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.ReplaceExpectedSets")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return err
	}

	if workout.UserID != userID {
		return domain.ErrNotFound
	}

	if !workout.FinishedAt.IsZero() {
		return fmt.Errorf("%w: workout %s is already finished", domain.ErrInvalidArgument, workoutID)
	}

	exerciseLog, err := s.repository.GetExerciseLogByID(ctx, exerciseLogID)
	if err != nil {
		return err
	}

	if exerciseLog.WorkoutID != workoutID {
		return domain.ErrNotFound
	}

	return s.unitOfWork.InTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repository.DeleteExpectedSetsByExerciseLogID(txCtx, exerciseLogID); err != nil {
			return err
		}

		for _, set := range sets {
			domainSet := domain.NewExpectedSet(
				exerciseLogID,
				set.SetType,
				set.Reps,
				set.Weight,
				set.Time,
			)

			if _, err := s.repository.CreateExpectedSet(txCtx, domainSet); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) DeleteSetLog(ctx context.Context, userID, workoutID, exerciseLogID domain.ID, setLogID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.DeleteSetLog")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to delete set log %s for exercise log %s in workout %s", userID, setLogID, exerciseLogID, workoutID)
		return domain.ErrNotFound
	}

	if !workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to delete set log %s for exercise log %s in finished workout %s", userID, setLogID, exerciseLogID, workoutID)
		return fmt.Errorf("%w: workout %s is already finished", domain.ErrInvalidArgument, workoutID)
	}

	exerciseLog, err := s.repository.GetExerciseLogByID(ctx, exerciseLogID)
	if err != nil {
		return err
	}

	if exerciseLog.WorkoutID != workoutID {
		logger.Errorf("user %s tried to delete set log %s for exercise log %s in workout %s", userID, setLogID, exerciseLogID, workoutID)
		return domain.ErrNotFound
	}

	setLog, err := s.repository.GetSetLogByID(ctx, setLogID)
	if err != nil {
		return err
	}

	if setLog.ExerciseLogID != exerciseLogID {
		logger.Errorf("user %s tried to delete set log %s for exercise log %s in workout %s", userID, setLogID, exerciseLogID, workoutID)
		return domain.ErrNotFound
	}

	err = s.repository.DeleteSetLog(ctx, setLogID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateSetLog(ctx context.Context, userID, workoutID, exerciseLogID, setLogID domain.ID, setlogDTO dto.UpdateSetLogDTO) (domain.ExerciseSetLog, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.UpdateSetLog")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return domain.ExerciseSetLog{}, err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to update set log %s for exercise log %s in workout %s", userID, setLogID, exerciseLogID, workoutID)
		return domain.ExerciseSetLog{}, domain.ErrNotFound
	}

	if !workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to update set log %s for exercise log %s in finished workout %s", userID, setLogID, exerciseLogID, workoutID)
		return domain.ExerciseSetLog{}, fmt.Errorf("%w: workout %s is already finished", domain.ErrInvalidArgument, workoutID)
	}

	exerciseLog, err := s.repository.GetExerciseLogByID(ctx, exerciseLogID)
	if err != nil {
		return domain.ExerciseSetLog{}, err
	}

	if exerciseLog.WorkoutID != workoutID {
		logger.Errorf("user %s tried to update set log %s for exercise log %s in workout %s", userID, setLogID, exerciseLogID, workoutID)
		return domain.ExerciseSetLog{}, domain.ErrNotFound
	}

	setLog, err := s.repository.GetSetLogByID(ctx, setLogID)
	if err != nil {
		return domain.ExerciseSetLog{}, err
	}

	if setLog.ExerciseLogID != exerciseLogID {
		logger.Errorf("user %s tried to update set log %s for exercise log %s in workout %s", userID, setLogID, exerciseLogID, workoutID)
		return domain.ExerciseSetLog{}, domain.ErrNotFound
	}

	if setlogDTO.Reps.IsValid {
		setLog.Reps = setlogDTO.Reps.V
	}

	if setlogDTO.Weight.IsValid {
		setLog.Weight = setlogDTO.Weight.V
	}

	setLog.UpdatedAt = time.Now()

	setLog, err = s.repository.UpdateSetLog(ctx, setLogID, setLog)
	if err != nil {
		return domain.ExerciseSetLog{}, err
	}

	return setLog, nil
}

func (s *Service) RateWorkout(ctx context.Context, userID, workoutID domain.ID, rating int) (domain.Workout, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.RateWorkout")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return domain.Workout{}, err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to rate workout %s", userID, workoutID)
		return domain.Workout{}, domain.ErrNotFound
	}

	if workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to rate unfinished workout %s", userID, workoutID)
		return domain.Workout{}, fmt.Errorf("%w: workout %s is not finished", domain.ErrInvalidArgument, workoutID)
	}

	workout.Rating = rating

	workout, err = s.repository.UpdateWorkout(ctx, workoutID, workout)
	if err != nil {
		return domain.Workout{}, err
	}

	return workout, nil
}

func (s *Service) AddCommentToWorkout(ctx context.Context, userID, workoutID domain.ID, comment string) (domain.Workout, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.AddCommentToWorkout")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return domain.Workout{}, err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to add comment to workout %s", userID, workoutID)
		return domain.Workout{}, domain.ErrNotFound
	}

	if workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to add comment to unfinished workout %s", userID, workoutID)
		return domain.Workout{}, fmt.Errorf("%w: workout %s is not finished", domain.ErrInvalidArgument, workoutID)
	}

	workout.Notes = comment

	workout, err = s.repository.UpdateWorkout(ctx, workoutID, workout)
	if err != nil {
		return domain.Workout{}, err
	}

	return workout, nil
}

func (s *Service) GetWorkouts(ctx context.Context, userID domain.ID, limit, offset int) ([]dto.WorkoutDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetWorkouts")
	defer span.Finish()

	workouts, err := s.repository.GetWorkouts(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	workoutsDTO := make([]dto.WorkoutDTO, 0, len(workouts))
	for _, workout := range workouts {
		exerciseLogs, err := s.repository.GetExerciseLogsByWorkoutID(ctx, workout.ID)
		if err != nil {
			return nil, err
		}

		workoutsDTO = append(workoutsDTO, dto.WorkoutDTO{
			Workout:      workout,
			ExerciseLogs: exerciseLogs,
		})
	}

	return workoutsDTO, nil
}

func (s *Service) AddNotesToExerciseLog(ctx context.Context, userID, workoutID, exerciseLogID domain.ID, notes string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.AddNotesToExerciseLog")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to add notes to exercise log %s for workout %s", userID, exerciseLogID, workoutID)
		return domain.ErrNotFound
	}

	if !workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to add notes to exercise instance of finished workout %s", userID, workoutID)
		return fmt.Errorf("%w: workout %s is already finished", domain.ErrInvalidArgument, workoutID)
	}

	exerciseLog, err := s.repository.GetExerciseLogByID(ctx, exerciseLogID)
	if err != nil {
		return err
	}

	if exerciseLog.WorkoutID != workoutID {
		logger.Errorf("user %s tried to add notes to exercise log %s for workout %s", userID, exerciseLogID, workoutID)
		return domain.ErrNotFound
	}

	exerciseLog.Notes = notes

	_, err = s.repository.UpdateExerciseLog(ctx, exerciseLogID, exerciseLog)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) AddPowerRatingToExerciseLog(ctx context.Context, userID, workoutID, exerciseLogID domain.ID, powerRating int) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.AddPowerRatingToExerciseLog")
	defer span.Finish()

	workout, err := s.repository.GetWorkoutByID(ctx, workoutID)
	if err != nil {
		return err
	}

	if workout.UserID != userID {
		logger.Errorf("user %s tried to add power rating to exercise log %s of workout %s", userID, exerciseLogID, workoutID)
		return domain.ErrNotFound
	}

	if !workout.FinishedAt.IsZero() {
		logger.Errorf("user %s tried to add power rating to exercise instance of finished workout %s", userID, workoutID)
		return fmt.Errorf("%w: workout %s is already finished", domain.ErrInvalidArgument, workoutID)
	}

	exerciseLog, err := s.repository.GetExerciseLogByID(ctx, exerciseLogID)
	if err != nil {
		return err
	}

	if exerciseLog.WorkoutID != workoutID {
		logger.Errorf("user %s tried to add notes to exercise log %s for workout %s", userID, exerciseLogID, workoutID)
		return domain.ErrNotFound
	}

	exerciseLog.PowerRating = powerRating

	_, err = s.repository.UpdateExerciseLog(ctx, exerciseLogID, exerciseLog)
	if err != nil {
		return err
	}

	return nil
}
