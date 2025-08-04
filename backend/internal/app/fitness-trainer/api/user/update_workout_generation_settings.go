package user

import (
	"context"
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/utils"

	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) UpdateWorkoutGenerationSettings(ctx context.Context, req *desc.UpdateWorkoutGenerationSettingsRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.user.UpdateWorkoutGenerationSettings")
	defer span.Finish()

	if err := req.ValidateAll(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	id, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user id not found in context: %w", domain.ErrUnauthorized)
	}

	var createDTO dto.CreateGenerationSettings
	{
		createDTO.BasePrompt = utils.NewNullable(req.GetBasePrompt(), req.BasePrompt != nil)
		createDTO.VarietyLevel = utils.NewNullable(int(req.GetVarietyLevel()), req.VarietyLevel != nil)

		createDTO.PrimaryGoal = utils.NewNullable(mappers.GoalFromProto(req.GetPrimaryGoal()), req.PrimaryGoal != nil)
		createDTO.SecondaryGoals = req.GetSecondaryGoals() 

		createDTO.ExperienceLevel = utils.NewNullable(mappers.ExperienceLevelFromProto(req.GetExperienceLevel()), req.ExperienceLevel != nil)
		createDTO.WorkoutPlanType = utils.NewNullable(mappers.WorkoutPlanTypeFromProto(req.GetWorkoutPlanType()), req.WorkoutPlanType != nil)

		createDTO.DaysPerWeek = utils.NewNullable(int(req.GetDaysPerWeek()), req.DaysPerWeek != nil)
		createDTO.SessionDurationMinutes = utils.NewNullable(int(req.GetSessionDurationMinutes()), req.SessionDurationMinutes != nil)

		createDTO.Injuries = utils.NewNullable(req.GetInjuries(), req.Injuries != nil)

		createDTO.PriorityMuscleGroupsIDs = make([]domain.ID, 0, len(req.GetPriorityMuscleGroupsIds()))
		for _, id := range req.GetPriorityMuscleGroupsIds() {
			id, err := domain.ParseID(id)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, fmt.Errorf("invalid muscle group id: %v", id))
			}
			createDTO.PriorityMuscleGroupsIDs = append(createDTO.PriorityMuscleGroupsIDs, id)
		}
	}

	if _, err := i.service.SaveGenerationSettings(ctx, id, createDTO); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
