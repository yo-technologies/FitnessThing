package mappers

import (
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	desc "fitness-trainer/pkg/workouts"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func WorkoutsToProto(workouts []domain.Workout) *desc.WorkoutsListResponse {
	workoutsList := make([]*desc.Workout, 0, len(workouts))
	for _, workout := range workouts {
		workoutsList = append(workoutsList, WorkoutToProto(workout))
	}

	return &desc.WorkoutsListResponse{
		Workouts: workoutsList,
	}
}

func WorkoutToProto(workout domain.Workout) *desc.Workout {
	var routineID *string
	if workout.RoutineID.IsValid {
		routineIDValue := workout.RoutineID.V.String()
		routineID = &routineIDValue
	}
	return &desc.Workout{
		Id:               workout.ID.String(),
		RoutineId:        routineID,
		UserId:           workout.UserID.String(),
		CreatedAt:        timestamppb.New(workout.CreatedAt),
		Notes:            workout.Notes,
		Rating:           int32(workout.Rating),
		FinishedAt:       timestamppb.New(workout.FinishedAt),
		UpdatedAt:        timestamppb.New(workout.UpdatedAt),
		Reasoning:        workout.Reasoning,
		IsAiGenerated:    workout.IsAIGenerated,
		GenerationStatus: generationStatusToProto(workout.GenerationStatus),
	}
}

func generationStatusToProto(s domain.WorkoutGenerationStatus) desc.GenerationStatus {
	switch s {
	case domain.WorkoutGenerationStatusRunning:
		return desc.GenerationStatus_GENERATION_STATUS_RUNNING
	case domain.WorkoutGenerationStatusFailed:
		return desc.GenerationStatus_GENERATION_STATUS_FAILED
	case domain.WorkoutGenerationStatusCompleted:
		return desc.GenerationStatus_GENERATION_STATUS_COMPLETED
	case domain.WorkoutGenerationStatusUnspecified:
		return desc.GenerationStatus_GENERATION_STATUS_UNSPECIFIED
	default:
		return desc.GenerationStatus_GENERATION_STATUS_UNSPECIFIED
	}
}

func SetLogToProto(setLog domain.ExerciseSetLog) *desc.SetLog {
	return &desc.SetLog{
		Id:        setLog.ID.String(),
		Reps:      int32(setLog.Reps),
		Weight:    setLog.Weight,
		CreatedAt: timestamppb.New(setLog.CreatedAt),
		UpdatedAt: timestamppb.New(setLog.UpdatedAt),
	}
}

func ExerciseLogToProto(exerciseLog domain.ExerciseLog) *desc.ExerciseLog {
	return &desc.ExerciseLog{
		Id:          exerciseLog.ID.String(),
		WorkoutId:   exerciseLog.WorkoutID.String(),
		ExerciseId:  exerciseLog.ExerciseID.String(),
		PowerRating: int32(exerciseLog.PowerRating),
		Notes:       exerciseLog.Notes,
		CreatedAt:   timestamppb.New(exerciseLog.CreatedAt),
		UpdatedAt:   timestamppb.New(exerciseLog.UpdatedAt),
		WeightUnit:  weightUnitToProto(exerciseLog.WeightUnit),
	}
}

func ExerciseLogDTOsToProto(in []dto.ExerciseLogDTO) []*desc.ExerciseLogDetails {
	out := make([]*desc.ExerciseLogDetails, 0, len(in))
	for _, ex := range in {
		out = append(out, ExerciseLogDTOToProto(ex))
	}
	return out
}

func ExpectedSetsToProto(expectedSets []domain.ExpectedSet) []*desc.ExpectedSet {
	out := make([]*desc.ExpectedSet, 0, len(expectedSets))
	for _, expectedSet := range expectedSets {
		out = append(out, ExpectedSetToProto(expectedSet))
	}
	return out
}

func ExpectedSetToProto(expectedSet domain.ExpectedSet) *desc.ExpectedSet {
	return &desc.ExpectedSet{
		Id:            expectedSet.ID.String(),
		ExerciseLogId: expectedSet.ExerciseLogID.String(),
		Reps:          int32(expectedSet.Reps),
		Weight:        expectedSet.Weight,
		Time:          durationpb.New(expectedSet.Time),
		CreatedAt:     timestamppb.New(expectedSet.CreatedAt),
		UpdatedAt:     timestamppb.New(expectedSet.UpdatedAt),
	}
}

func ExerciseLogDTOToProto(in dto.ExerciseLogDTO) *desc.ExerciseLogDetails {
	return &desc.ExerciseLogDetails{
		ExerciseLog:  ExerciseLogToProto(in.ExerciseLog),
		Exercise:     ExerciseToProto(in.Exercise),
		SetLogs:      SetLogsToProto(in.SetLogs),
		ExpectedSets: ExpectedSetsToProto(in.ExpectedSets),
	}
}

func SetLogsToProto(setLogs []domain.ExerciseSetLog) []*desc.SetLog {
	setLogsList := make([]*desc.SetLog, 0, len(setLogs))
	for _, setLog := range setLogs {
		setLogsList = append(setLogsList, SetLogToProto(setLog))
	}

	return setLogsList
}

func ExerciseLogsToProto(exerciseLogs []domain.ExerciseLog) []*desc.ExerciseLog {
	result := make([]*desc.ExerciseLog, 0, len(exerciseLogs))
	for _, exerciseLog := range exerciseLogs {
		result = append(result, ExerciseLogToProto(exerciseLog))
	}

	return result
}

func WorkoutDTOToProto(workoutDTO dto.WorkoutDTO) *desc.GetWorkoutsResponse_WorkoutDetails {
	return &desc.GetWorkoutsResponse_WorkoutDetails{
		Workout:      WorkoutToProto(workoutDTO.Workout),
		ExerciseLogs: ExerciseLogsToProto(workoutDTO.ExerciseLogs),
	}
}

func WorkoutsDTOToProto(workoutDTOs []dto.WorkoutDTO) []*desc.GetWorkoutsResponse_WorkoutDetails {
	result := make([]*desc.GetWorkoutsResponse_WorkoutDetails, 0, len(workoutDTOs))
	for _, workoutDTO := range workoutDTOs {
		result = append(result, WorkoutDTOToProto(workoutDTO))
	}

	return result
}
