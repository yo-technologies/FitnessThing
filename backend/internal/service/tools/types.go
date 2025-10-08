package tools

import (
	"context"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fmt"
	"time"

	"github.com/samber/lo"
)

type exercise struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Targets     []domain.MuscleGroup `json:"target_muscle_groups"`
}

func exerciseFromDomain(e domain.Exercise) exercise {
	return exercise{
		ID:          e.ID.String(),
		Name:        e.Name,
		Description: e.Description,
		Targets:     e.TargetMuscleGroups,
	}
}

type expectedSet struct {
	ID         string  `json:"id"`
	SetType    string  `json:"set_type"`
	Reps       int     `json:"reps"`
	Weight     float32 `json:"weight"`
	TimeSecond int     `json:"time_seconds"`
}

func expectedSetFromDomain(s domain.ExpectedSet) expectedSet {
	return expectedSet{
		ID:         s.ID.String(),
		SetType:    s.SetType.String(),
		Reps:       s.Reps,
		Weight:     s.Weight,
		TimeSecond: int(s.Time.Seconds()),
	}
}

func expectedSetListFromDomain(s []domain.ExpectedSet) []expectedSet {
	return lo.Map(s, func(item domain.ExpectedSet, _ int) expectedSet {
		return expectedSetFromDomain(item)
	})
}

type setLog struct {
	ID         string  `json:"id"`
	Reps       int     `json:"reps"`
	Weight     float32 `json:"weight"`
	TimeSecond int     `json:"time_seconds"`
}

func setLogFromDomain(s domain.ExerciseSetLog) setLog {
	return setLog{
		ID:         s.ID.String(),
		Reps:       s.Reps,
		Weight:     s.Weight,
		TimeSecond: int(s.Time.Seconds()),
	}
}

func setLogListFromDomain(s []domain.ExerciseSetLog) []setLog {
	return lo.Map(s, func(item domain.ExerciseSetLog, _ int) setLog {
		return setLogFromDomain(item)
	})
}

type workoutPlanExercise struct {
	ExerciseLogID string        `json:"exercise_log_id"`
	Exercise      exercise      `json:"exercise"`
	ExpectedSets  []expectedSet `json:"expected_sets"`
	SetLogs       []setLog      `json:"set_logs"`
}

func workoutPlanExerciseFromDomain(e dto.ExerciseLogDTO) workoutPlanExercise {
	return workoutPlanExercise{
		ExerciseLogID: e.ExerciseLog.ID.String(),
		Exercise:      exerciseFromDomain(e.Exercise),
		ExpectedSets:  expectedSetListFromDomain(e.ExpectedSets),
		SetLogs:       setLogListFromDomain(e.SetLogs),
	}
}

func workoutPlanExerciseListFromDomain(exerciseLogs []dto.ExerciseLogDTO) []workoutPlanExercise {
	return lo.Map(exerciseLogs, func(item dto.ExerciseLogDTO, _ int) workoutPlanExercise {
		return workoutPlanExerciseFromDomain(item)
	})
}

type workoutPlan struct {
	Exercises []workoutPlanExercise `json:"exercises"`
}

func workoutPlanFromDomain(w dto.WorkoutDetailsDTO) workoutPlan {
	return workoutPlan{
		Exercises: workoutPlanExerciseListFromDomain(w.ExerciseLogs),
	}
}

// Argument structs for tool handlers
type addExercisesToWorkoutArgs struct {
	ExerciseIDs []string `json:"exercise_ids"`
}

type listExercisesArgs struct {
	MuscleGroupIDs     []string `json:"muscle_group_ids"`
	ExcludeExerciseIDs []string `json:"exclude_exercise_ids"`
	Limit              *int     `json:"limit"`
}

type updateSetLogArgs struct {
	ExerciseLogID string   `json:"exercise_log_id"`
	SetLogID      string   `json:"set_log_id"`
	Reps          *int     `json:"reps"`
	Weight        *float64 `json:"weight"`
}

type getExerciseHistoryArgs struct {
	ExerciseID string `json:"exercise_id"`
	LogsLimit  *int   `json:"logs_limit"`
}

type getWorkoutHistoryArgs struct {
	Limit int `json:"limit"`
}

type removeExerciseLogsArgs struct {
	ExerciseLogIDs []string `json:"exercise_log_ids"`
}

type replaceExpectedSetsArgs struct {
	ExerciseLogID string             `json:"exercise_log_id"`
	Sets          []expectedSetInput `json:"sets"`
}

type expectedSetInput struct {
	SetType     string   `json:"set_type"`
	Reps        *int     `json:"reps"`
	Weight      *float64 `json:"weight"`
	TimeSeconds *int     `json:"time_seconds"`
}

type exerciseLogHistory struct {
	ID        string          `json:"id"`
	WorkoutID string          `json:"workout_id"`
	CreatedAt time.Time       `json:"created_at"`
	Notes     string          `json:"notes"`
	SetLogs   []setLogHistory `json:"set_logs"`
}

func exerciseLogHistoryFromDomain(log dto.ExerciseLogDTO) exerciseLogHistory {
	return exerciseLogHistory{
		ID:        log.ExerciseLog.ID.String(),
		WorkoutID: log.ExerciseLog.WorkoutID.String(),
		CreatedAt: log.ExerciseLog.CreatedAt,
		Notes:     log.ExerciseLog.Notes,
		SetLogs:   setLogHistoryListFromDomain(log.SetLogs),
	}
}

func exerciseLogHistoryListFromDomain(logs []dto.ExerciseLogDTO) []exerciseLogHistory {
	return lo.Map(logs, func(item dto.ExerciseLogDTO, _ int) exerciseLogHistory {
		return exerciseLogHistoryFromDomain(item)
	})
}

type setLogHistory struct {
	ID      string  `json:"id"`
	Reps    int     `json:"reps"`
	Weight  float32 `json:"weight"`
	TimeSec int     `json:"time_seconds"`
}

func setLogHistoryFromDomain(s domain.ExerciseSetLog) setLogHistory {
	return setLogHistory{
		ID:      s.ID.String(),
		Reps:    s.Reps,
		Weight:  s.Weight,
		TimeSec: int(s.Time.Seconds()),
	}
}

func setLogHistoryListFromDomain(s []domain.ExerciseSetLog) []setLogHistory {
	return lo.Map(s, func(item domain.ExerciseSetLog, _ int) setLogHistory {
		return setLogHistoryFromDomain(item)
	})
}

type exerciseLogHistoryPayload struct {
	ExerciseLogs []exerciseLogHistory `json:"exercise_logs"`
}

func exerciseLogHistoryPayloadFromDomain(logs []dto.ExerciseLogDTO) exerciseLogHistoryPayload {
	return exerciseLogHistoryPayload{
		ExerciseLogs: exerciseLogHistoryListFromDomain(logs),
	}
}

type workoutPayload struct {
	ID         string     `json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt time.Time  `json:"finished_at"`
	Exercises  []exercise `json:"exercises"`
}

type listExercisesResponse struct {
	Exercises []exercise `json:"exercises"`
}

type getWorkoutHistoryResponse struct {
	Workouts []workoutPayload `json:"workouts"`
}

type listMuscleGroupsResponse struct {
	MuscleGroups []dto.MuscleGroupDTO `json:"muscle_groups"`
}

func (t *Tools) convertWorkoutsToHistoryResponse(ctx context.Context, workoutsDTO []dto.WorkoutDTO) (getWorkoutHistoryResponse, error) {
	response := getWorkoutHistoryResponse{Workouts: make([]workoutPayload, 0, len(workoutsDTO))}

	for _, workoutDTO := range workoutsDTO {
		exercises := make([]exercise, 0, len(workoutDTO.ExerciseLogs))
		for _, log := range workoutDTO.ExerciseLogs {
			exercise, err := t.service.GetExerciseByID(ctx, log.ExerciseID)
			if err != nil {
				return response, fmt.Errorf("failed to load exercise %s: %w", log.ExerciseID, err)
			}
			exercises = append(exercises, exerciseFromDomain(exercise))
		}

		response.Workouts = append(response.Workouts, workoutPayload{
			ID:         workoutDTO.Workout.ID.String(),
			CreatedAt:  workoutDTO.Workout.CreatedAt,
			FinishedAt: workoutDTO.Workout.FinishedAt,
			Exercises:  exercises,
		})
	}

	return response, nil
}

func convertExercisesToListResponse(exercises []domain.Exercise) listExercisesResponse {
	response := listExercisesResponse{Exercises: make([]exercise, 0, len(exercises))}
	for _, ex := range exercises {
		response.Exercises = append(response.Exercises, exerciseFromDomain(ex))
	}
	return response
}
