package domain

import (
	"encoding/json"
	"fitness-trainer/internal/utils"
	"fmt"
	"math"
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

func (i ID) MarshalText() ([]byte, error) {
	return []byte(uuid.UUID(i).String()), nil
}

func (i *ID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*i = ID{}
		return nil
	}

	parsed, err := uuid.ParseBytes(text)
	if err != nil {
		return err
	}

	*i = ID(parsed)
	return nil
}

func (i ID) MarshalJSON() ([]byte, error) {
	text, err := i.MarshalText()
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(text))
}

func (i *ID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*i = ID{}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return i.UnmarshalText([]byte(s))
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

	TelegramID             int64
	TelegramUsername       utils.Nullable[string]
	FirstName              utils.Nullable[string]
	LastName               utils.Nullable[string]
	DateOfBirth            time.Time
	Height                 utils.Nullable[float32]
	Weight                 utils.Nullable[float32]
	ProfilePicURL          utils.Nullable[string]
	HasCompletedOnboarding bool
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

const (
	MaxUserFactsPerUser = 32
	MaxUserFactLength   = 512
)

type UserFact struct {
	Model

	UserID  ID
	Content string
}

func NewUserFact(userID ID, content string) UserFact {
	return UserFact{
		Model:   NewModel(),
		UserID:  userID,
		Content: content,
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

	UserID     ID
	RoutineID  utils.Nullable[ID]
	Notes      string
	Rating     utils.Nullable[int]
	FinishedAt time.Time
}

func NewWorkout(userID ID, routineID utils.Nullable[ID]) Workout {
	return Workout{
		Model:     NewModel(),
		UserID:    userID,
		RoutineID: routineID,
	}
}

type ExerciseLog struct {
	Model

	WorkoutID   ID
	ExerciseID  ID
	Notes       string
	PowerRating int
	WeightUnit  WeightUnit
}

func NewExerciseLog(workoutID, exerciseID ID, weightUnit WeightUnit) ExerciseLog {
	if weightUnit == WeightUnitUnknown {
		weightUnit = WeightUnitKG
	}
	return ExerciseLog{
		Model:      NewModel(),
		WorkoutID:  workoutID,
		ExerciseID: exerciseID,
		WeightUnit: weightUnit,
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

// UpdateWeight sets a new weight and updates the timestamp
func (s *ExerciseSetLog) UpdateWeight(weight float32) {
	s.Weight = weight
	s.UpdatedAt = time.Now()
}

type Chat struct {
	Model

	UserID    ID
	WorkoutID utils.Nullable[ID]
	Title     string
}

func NewChat(userID ID, workoutID utils.Nullable[ID], title string) Chat {
	return Chat{
		Model:     NewModel(),
		UserID:    userID,
		WorkoutID: workoutID,
		Title:     title,
	}
}

type ChatMessageRole string

const (
	ChatMessageRoleUser      ChatMessageRole = "user"
	ChatMessageRoleAssistant ChatMessageRole = "assistant"
	ChatMessageRoleTool      ChatMessageRole = "tool"
	ChatMessageRoleSystem    ChatMessageRole = "system"
)

func (r ChatMessageRole) String() string {
	return string(r)
}

type ChatMessage struct {
	Model

	ChatID        ID
	Role          ChatMessageRole
	Content       string
	ToolName      utils.Nullable[string]
	ToolCallID    utils.Nullable[string]
	ToolArguments map[string]any
	TokenUsage    utils.Nullable[int]
	Error         utils.Nullable[string]
}

func NewChatMessage(
	chatID ID,
	role ChatMessageRole,
	content string,
	toolName utils.Nullable[string],
	toolCallID utils.Nullable[string],
	toolArguments map[string]any,
) ChatMessage {
	return ChatMessage{
		Model:         NewModel(),
		ChatID:        chatID,
		Role:          role,
		Content:       content,
		ToolName:      toolName,
		ToolCallID:    toolCallID,
		ToolArguments: toolArguments,
	}
}

// WeightUnit определяет единицы измерения веса
type WeightUnit string

const (
	WeightUnitUnknown WeightUnit = ""
	WeightUnitKG      WeightUnit = "kg"
	WeightUnitLB      WeightUnit = "lb"
)

func (w WeightUnit) String() string { return string(w) }

func NewWeightUnit(s string) (WeightUnit, error) {
	switch s {
	case "kg":
		return WeightUnitKG, nil
	case "lb":
		return WeightUnitLB, nil
	case "":
		return WeightUnitUnknown, nil
	default:
		return "", fmt.Errorf("unknown weight unit: %w", ErrInvalidArgument)
	}
}

// Conversion factors between weight units
const (
	lbToKg float32 = 0.45359237
	kgToLb float32 = 2.2046226218
)

// ConversionFactor returns multiplicative factor to convert a value from current unit to target unit
func (from WeightUnit) ConversionFactor(to WeightUnit) float32 {
	if from == to || from == WeightUnitUnknown || to == WeightUnitUnknown {
		return 1
	}
	if from == WeightUnitKG && to == WeightUnitLB {
		return kgToLb
	}
	if from == WeightUnitLB && to == WeightUnitKG {
		return lbToKg
	}
	return 1
}

// roundToStep rounds value to the nearest multiple of step (e.g., 2.5)
func roundToStep(value, step float32) float32 {
	if step <= 0 {
		return value
	}
	return float32(math.Round(float64(value/step))) * step
}

// ConvertWeight converts a numeric weight value between units
func ConvertWeight(value float32, from, to WeightUnit) float32 {
	raw := value * from.ConversionFactor(to)
	return roundToStep(raw, 2.5)
}
