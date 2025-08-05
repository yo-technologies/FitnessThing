package mappers

import (
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/utils"
	desc "fitness-trainer/pkg/workouts"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func GoalToProto(goal domain.Goal) desc.Goal {
	switch goal {
	case domain.GoalEndurance:
		return desc.Goal_GOAL_ENDURANCE
	case domain.GoalWeightLoss:
		return desc.Goal_GOAL_WEIGHT_LOSS
	case domain.GoalMuscleGain:
		return desc.Goal_GOAL_MUSCLE_GAIN
	case domain.GoalFlexibility:
		return desc.Goal_GOAL_FLEXIBILITY
	case domain.GoalStrength:
		return desc.Goal_GOAL_STRENGTH
	default:
		return desc.Goal_GOAL_UNSPECIFIED
	}
}

func GoalFromProto(goal desc.Goal) domain.Goal {
	switch goal {
	case desc.Goal_GOAL_ENDURANCE:
		return domain.GoalEndurance
	case desc.Goal_GOAL_WEIGHT_LOSS:
		return domain.GoalWeightLoss
	case desc.Goal_GOAL_MUSCLE_GAIN:
		return domain.GoalMuscleGain
	case desc.Goal_GOAL_FLEXIBILITY:
		return domain.GoalFlexibility
	case desc.Goal_GOAL_STRENGTH:
		return domain.GoalStrength
	default:
		return domain.GoalUnspecified
	}
}

func ExperienceLevelToProto(level domain.ExperienceLevel) desc.ExperienceLevel {
	switch level {
	case domain.ExperienceLevelBeginner:
		return desc.ExperienceLevel_EXPERIENCE_LEVEL_BEGINNER
	case domain.ExperienceLevelIntermediate:
		return desc.ExperienceLevel_EXPERIENCE_LEVEL_INTERMEDIATE
	case domain.ExperienceLevelAdvanced:
		return desc.ExperienceLevel_EXPERIENCE_LEVEL_ADVANCED
	default:
		return desc.ExperienceLevel_EXPERIENCE_LEVEL_UNSPECIFIED
	}
}

func ExperienceLevelFromProto(level desc.ExperienceLevel) domain.ExperienceLevel {
	switch level {
	case desc.ExperienceLevel_EXPERIENCE_LEVEL_BEGINNER:
		return domain.ExperienceLevelBeginner
	case desc.ExperienceLevel_EXPERIENCE_LEVEL_INTERMEDIATE:
		return domain.ExperienceLevelIntermediate
	case desc.ExperienceLevel_EXPERIENCE_LEVEL_ADVANCED:
		return domain.ExperienceLevelAdvanced
	default:
		return domain.ExperienceLevelUnspecified
	}
}

func WorkoutPlanTypeToProto(planType domain.WorkoutPlanType) desc.WorkoutPlanType {
	switch planType {
	case domain.WorkoutPlanTypeFullBody:
		return desc.WorkoutPlanType_WORKOUT_PLAN_TYPE_FULL_BODY
	case domain.WorkoutPlanTypeSplit:
		return desc.WorkoutPlanType_WORKOUT_PLAN_TYPE_SPLIT
	case domain.WorkoutPlanTypeUpperLower:
		return desc.WorkoutPlanType_WORKOUT_PLAN_TYPE_UPPER_LOWER
	case domain.WorkoutPlanTypePushPullLegs:
		return desc.WorkoutPlanType_WORKOUT_PLAN_TYPE_PUSH_PULL_LEGS
	default:
		return desc.WorkoutPlanType_WORKOUT_PLAN_TYPE_UNSPECIFIED
	}
}

func WorkoutPlanTypeFromProto(planType desc.WorkoutPlanType) domain.WorkoutPlanType {
	switch planType {
	case desc.WorkoutPlanType_WORKOUT_PLAN_TYPE_FULL_BODY:
		return domain.WorkoutPlanTypeFullBody
	case desc.WorkoutPlanType_WORKOUT_PLAN_TYPE_SPLIT:
		return domain.WorkoutPlanTypeSplit
	case desc.WorkoutPlanType_WORKOUT_PLAN_TYPE_UPPER_LOWER:
		return domain.WorkoutPlanTypeUpperLower
	case desc.WorkoutPlanType_WORKOUT_PLAN_TYPE_PUSH_PULL_LEGS:
		return domain.WorkoutPlanTypePushPullLegs
	default:
		return domain.WorkoutPlanTypeUnspecified
	}
}

func GenerationSettingsToProto(settings domain.GenerationSettings) *desc.WorkoutGenerationSettings {
	return &desc.WorkoutGenerationSettings{
		BasePrompt:              nullableStringToOptionalProto(settings.BasePrompt),
		VarietyLevel:            nullableIntToOptionalProto(settings.VarietyLevel),
		PrimaryGoal:             GoalToProto(settings.PrimaryGoal),
		SecondaryGoals:          settings.SecondaryGoals,
		ExperienceLevel:         ExperienceLevelToProto(settings.ExperienceLevel),
		DaysPerWeek:             nullableIntToOptionalProto(settings.DaysPerWeek),
		SessionDurationMinutes:  nullableIntToOptionalProto(settings.SessionDurationMinutes),
		Injuries:                nullableStringToOptionalProto(settings.Injuries),
		PriorityMuscleGroupsIds: idsToStrings(settings.PriorityMuscleGroupsIDs),
		WorkoutPlanType:         WorkoutPlanTypeToProto(settings.WorkoutPlanType),
		UpdatedAt:               timestamppb.New(settings.UpdatedAt),
	}
}

func UpdateGenerationSettingsRequestToCreateDTO(req *desc.UpdateWorkoutGenerationSettingsRequest) (dto.CreateGenerationSettings, error) {

	var createDTO dto.CreateGenerationSettings
	{
		createDTO.BasePrompt = utils.NewNullable(req.GetBasePrompt(), req.BasePrompt != nil)
		createDTO.VarietyLevel = utils.NewNullable(int(req.GetVarietyLevel()), req.VarietyLevel != nil)

		createDTO.PrimaryGoal = utils.NewNullable(GoalFromProto(req.GetPrimaryGoal()), req.PrimaryGoal != nil)
		createDTO.SecondaryGoals = req.GetSecondaryGoals()

		createDTO.ExperienceLevel = utils.NewNullable(ExperienceLevelFromProto(req.GetExperienceLevel()), req.ExperienceLevel != nil)
		createDTO.WorkoutPlanType = utils.NewNullable(WorkoutPlanTypeFromProto(req.GetWorkoutPlanType()), req.WorkoutPlanType != nil)

		createDTO.DaysPerWeek = utils.NewNullable(int(req.GetDaysPerWeek()), req.DaysPerWeek != nil)
		createDTO.SessionDurationMinutes = utils.NewNullable(int(req.GetSessionDurationMinutes()), req.SessionDurationMinutes != nil)

		createDTO.Injuries = utils.NewNullable(req.GetInjuries(), req.Injuries != nil)

		for _, rawID := range req.GetPriorityMuscleGroupsIds() {
			parsedID, err := domain.ParseID(rawID)
			if err != nil {
				return dto.CreateGenerationSettings{}, fmt.Errorf(
					"%w: %w",
					domain.ErrInvalidArgument,
					fmt.Errorf("invalid muscle group id: %v", rawID),
				)
			}
			createDTO.PriorityMuscleGroupsIDs = append(createDTO.PriorityMuscleGroupsIDs, parsedID)
		}
	}

	return createDTO, nil
}
