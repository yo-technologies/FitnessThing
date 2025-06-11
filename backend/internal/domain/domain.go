package domain

import (
	"fitness-trainer/internal/utils"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ID uuid.UUID

func NewID() ID {
	return ID(uuid.New())
}

func (i ID) String() string {
	return uuid.UUID(i).String()
}

func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ID{}, err
	}
	return ID(id), nil
}

type Model struct {
	ID        ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewModel() Model {
	return Model{
		ID:        NewID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

type User struct {
	Model

	TelegramID       int64
	TelegramUsername utils.Nullable[string]
	FirstName        utils.Nullable[string]
	LastName         utils.Nullable[string]
	DateOfBirth      time.Time
	Height           utils.Nullable[float32]
	Weight           utils.Nullable[float32]
	ProfilePicURL    utils.Nullable[string]
}

func NewUser(
	TelegramID int64,
	TelegramUsername utils.Nullable[string],
	FirstName utils.Nullable[string],
	LastName utils.Nullable[string],
	ProfilePicURL utils.Nullable[string],
) User {
	return User{
		Model:            NewModel(),
		TelegramID:       TelegramID,
		TelegramUsername: TelegramUsername,
		FirstName:        FirstName,
		LastName:         LastName,
		ProfilePicURL:    ProfilePicURL,
	}
}

type MuscleGroup string

const (
	MuscleGroupChest      MuscleGroup = "chest"
	MuscleGroupBack       MuscleGroup = "lats"
	MuscleGroupShoulders  MuscleGroup = "shoulders"
	MuscleGroupBiceps     MuscleGroup = "biceps"
	MuscleGroupTriceps    MuscleGroup = "triceps"
	MuscleGroupForearms   MuscleGroup = "forearms"
	MuscleGroupAbs        MuscleGroup = "abs"
	MuscleGroupQuads      MuscleGroup = "quads"
	MuscleGroupHamstrings MuscleGroup = "hamstrings"
	MuscleGroupCalves     MuscleGroup = "calves"
	MuscleGroupGlutes     MuscleGroup = "glutes"
	MuscleGroupLowerBack  MuscleGroup = "lower-back"
	MuscleGroupTraps      MuscleGroup = "traps"
)

func (m MuscleGroup) String() string {
	return string(m)
}

func NewMuscleGroup(m string) (MuscleGroup, error) {
	switch m {
	case "chest":
		return MuscleGroupChest, nil
	case "lats":
		return MuscleGroupBack, nil
	case "shoulders":
		return MuscleGroupShoulders, nil
	case "biceps":
		return MuscleGroupBiceps, nil
	case "triceps":
		return MuscleGroupTriceps, nil
	case "forearms":
		return MuscleGroupForearms, nil
	case "abs":
		return MuscleGroupAbs, nil
	case "quads":
		return MuscleGroupQuads, nil
	case "hamstrings":
		return MuscleGroupHamstrings, nil
	case "calves":
		return MuscleGroupCalves, nil
	case "glutes":
		return MuscleGroupGlutes, nil
	case "lower back":
		return MuscleGroupLowerBack, nil
	case "traps":
		return MuscleGroupTraps, nil
	default:
		return "", fmt.Errorf("unknown muscle group: %w", ErrInvalidArgument)
	}
}

type Exercise struct {
	Model

	Name               string
	Description        string
	VideoURL           string
	TargetMuscleGroups []MuscleGroup
}

func NewExercise(name, description, videoURL string, targetMuscleGroups []MuscleGroup) Exercise {
	return Exercise{
		Model:              NewModel(),
		Name:               name,
		Description:        description,
		VideoURL:           videoURL,
		TargetMuscleGroups: targetMuscleGroups,
	}
}

type Routine struct {
	Model

	UserID        ID
	Name          string
	Description   string
	ExerciseCount int
}

func NewRoutine(userID ID, name, description string) Routine {
	return Routine{
		Model:       NewModel(),
		UserID:      userID,
		Name:        name,
		Description: description,
	}
}

type SetType string

const (
	SetTypeUnknown SetType = ""
	SetTypeReps    SetType = "reps"
	SetTypeWeight  SetType = "weight"
	SetTypeTime    SetType = "time"
)

func (s SetType) String() string {
	return string(s)
}

func NewSetType(s string) (SetType, error) {
	switch s {
	case "reps":
		return SetTypeReps, nil
	case "weight":
		return SetTypeWeight, nil
	case "time":
		return SetTypeTime, nil
	default:
		return "", fmt.Errorf("unknown set type: %w", ErrInvalidArgument)
	}
}

type ExerciseInstance struct {
	Model

	RoutineID  ID
	ExerciseID ID
}

func NewExerciseInstance(routineID, exerciseID ID) ExerciseInstance {
	return ExerciseInstance{
		Model:      NewModel(),
		RoutineID:  routineID,
		ExerciseID: exerciseID,
	}
}

type Set struct {
	Model

	ExerciseInstanceID ID
	SetType            SetType
	Reps               int
	Weight             float32
	Time               time.Duration
}

func NewSet(exerciseInstanceID ID, setType SetType, reps int, weight float32, time time.Duration) Set {
	return Set{
		Model:              NewModel(),
		ExerciseInstanceID: exerciseInstanceID,
		SetType:            setType,
		Reps:               reps,
		Weight:             weight,
		Time:               time,
	}
}

type Workout struct {
	Model

	UserID        ID
	RoutineID     utils.Nullable[ID]
	Notes         string
	Rating        int
	FinishedAt    time.Time
	IsAIGenerated bool
	Reasoning     string
}

func NewWorkout(userID ID, routineID utils.Nullable[ID], isAIGenerated bool) Workout {
	return Workout{
		Model:         NewModel(),
		UserID:        userID,
		RoutineID:     routineID,
		IsAIGenerated: isAIGenerated,
	}
}

type ExerciseLog struct {
	Model

	WorkoutID   ID
	ExerciseID  ID
	Notes       string
	PowerRating int
}

func NewExerciseLog(workoutID, exerciseID ID) ExerciseLog {
	return ExerciseLog{
		Model:      NewModel(),
		WorkoutID:  workoutID,
		ExerciseID: exerciseID,
	}
}

type ExpectedSet struct {
	Model

	ExerciseLogID ID
	SetType       SetType
	Reps          int
	Weight        float32
	Time          time.Duration
}

func NewExpectedSet(exerciseLogID ID, setType SetType, reps int, weight float32, time time.Duration) ExpectedSet {
	return ExpectedSet{
		Model:         NewModel(),
		ExerciseLogID: exerciseLogID,
		SetType:       setType,
		Reps:          reps,
		Weight:        weight,
		Time:          time,
	}
}

type ExerciseSetLog struct {
	Model

	ExerciseLogID ID
	Reps          int
	Weight        float32
	Time          time.Duration
}

func NewExerciseSetLog(exerciseLogID ID, reps int, weight float32, time time.Duration) ExerciseSetLog {
	return ExerciseSetLog{
		Model:         NewModel(),
		ExerciseLogID: exerciseLogID,
		Reps:          reps,
		Weight:        weight,
		Time:          time,
	}
}
