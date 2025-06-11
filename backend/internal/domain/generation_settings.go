package domain

import "fitness-trainer/internal/utils"

type Goal int64

const (
	GoalUnspecified Goal = iota
	GoalMuscleGain
	GoalWeightLoss
	GoalStrength
	GoalEndurance
	GoalFlexibility
)

var goalNames = map[Goal]string{
	GoalUnspecified: "Unspecified",
	GoalMuscleGain:  "Muscle Gain",
	GoalWeightLoss:  "Weight Loss",
	GoalStrength:    "Strength",
	GoalEndurance:   "Endurance",
	GoalFlexibility: "Flexibility",
}

func (g Goal) IsValid() bool {
	_, exists := goalNames[g]
	return exists
}

type ExperienceLevel int64

const (
	ExperienceLevelUnspecified ExperienceLevel = iota
	ExperienceLevelBeginner
	ExperienceLevelIntermediate
	ExperienceLevelAdvanced
)

var experienceLevelNames = map[ExperienceLevel]string{
	ExperienceLevelUnspecified:  "Unspecified",
	ExperienceLevelBeginner:     "Beginner",
	ExperienceLevelIntermediate: "Intermediate",
	ExperienceLevelAdvanced:     "Advanced",
}

func (e ExperienceLevel) String() string {
	if name, exists := experienceLevelNames[e]; exists {
		return name
	}
	return "Unknown Experience Level"
}

type WorkoutPlanType int64

const (
	WorkoutPlanTypeUnspecified WorkoutPlanType = iota
	WorkoutPlanTypeFullBody
	WorkoutPlanTypeSplit
	WorkoutPlanTypeUpperLower
	WorkoutPlanTypePushPullLegs
)

var workoutPlanTypeNames = map[WorkoutPlanType]string{
	WorkoutPlanTypeUnspecified:  "Unspecified",
	WorkoutPlanTypeFullBody:     "Full Body",
	WorkoutPlanTypeSplit:        "Split",
	WorkoutPlanTypeUpperLower:   "Upper/Lower",
	WorkoutPlanTypePushPullLegs: "Push/Pull/Legs",
}

func (w WorkoutPlanType) String() string {
	if name, exists := workoutPlanTypeNames[w]; exists {
		return name
	}
	return "Unknown Workout Plan Type"
}

type GenerationSettings struct {
	Model

	UserID                  ID
	BasePrompt              utils.Nullable[string]
	VarietyLevel            utils.Nullable[int]
	PrimaryGoal             Goal
	SecondaryGoals          []string
	ExperienceLevel         ExperienceLevel
	DaysPerWeek             utils.Nullable[int]
	SessionDurationMinutes  utils.Nullable[int]
	Injuries                utils.Nullable[string]
	PriorityMuscleGroupsIDs []ID
	WorkoutPlanType         WorkoutPlanType
}

func NewGenerationSettings(userID ID) GenerationSettings {
	return GenerationSettings{
		Model:  NewModel(),
		UserID: userID,
	}
}
