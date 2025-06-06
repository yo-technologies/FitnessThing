package mappers

import (
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	desc "fitness-trainer/pkg/workouts"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func RoutineToProto(routine domain.Routine) *desc.Routine {
	return &desc.Routine{
		Id:            routine.ID.String(),
		UserId:        routine.UserID.String(),
		Name:          routine.Name,
		Description:   routine.Description,
		CreatedAt:     timestamppb.New(routine.CreatedAt),
		UpdatedAt:     timestamppb.New(routine.UpdatedAt),
		ExerciseCount: int32(routine.ExerciseCount),
	}
}

func RoutinesToProto(routines []domain.Routine) []*desc.Routine {
	result := make([]*desc.Routine, 0, len(routines))
	for _, routine := range routines {
		result = append(result, RoutineToProto(routine))
	}

	return result
}

func ExerciseInstanceToProto(instance domain.ExerciseInstance) *desc.ExerciseInstance {
	return &desc.ExerciseInstance{
		Id:         instance.ID.String(),
		RoutineId:  instance.RoutineID.String(),
		ExerciseId: instance.ExerciseID.String(),
		CreatedAt:  timestamppb.New(instance.CreatedAt),
		UpdatedAt:  timestamppb.New(instance.UpdatedAt),
	}
}

func ExerciseInstancesToProto(instances []domain.ExerciseInstance) []*desc.ExerciseInstance {
	result := make([]*desc.ExerciseInstance, 0, len(instances))
	for _, instance := range instances {
		result = append(result, ExerciseInstanceToProto(instance))
	}

	return result
}

func ExerciseInstanceDetailToProto(instance dto.ExerciseInstanceDetailsDTO) *desc.ExerciseInstanceDetails {
	return &desc.ExerciseInstanceDetails{
		ExerciseInstance: &desc.ExerciseInstance{
			Id:         instance.ID.String(),
			RoutineId:  instance.RoutineID.String(),
			ExerciseId: instance.ExerciseID.String(),
			CreatedAt:  timestamppb.New(instance.CreatedAt),
			UpdatedAt:  timestamppb.New(instance.UpdatedAt),
		},
		Exercise: ExerciseToProto(instance.Exercise),
		Sets:     SetsToProto(instance.Sets),
	}
}

func ExerciseInstanceDetailsToProto(instances []dto.ExerciseInstanceDetailsDTO) []*desc.ExerciseInstanceDetails {
	result := make([]*desc.ExerciseInstanceDetails, 0, len(instances))
	for _, instance := range instances {
		result = append(result, ExerciseInstanceDetailToProto(instance))
	}

	return result
}

func RoutineDetailsDTOToProto(instance dto.RoutineDetailsDTO) *desc.RoutineDetailResponse {
	exerciseInstances := ExerciseInstanceDetailsToProto(instance.ExerciseInstances)

	return &desc.RoutineDetailResponse{
		Routine: &desc.Routine{
			Id:          instance.ID.String(),
			UserId:      instance.UserID.String(),
			Name:        instance.Name,
			Description: instance.Description,
			CreatedAt:   timestamppb.New(instance.CreatedAt),
			UpdatedAt:   timestamppb.New(instance.UpdatedAt),
		},
		ExerciseInstances: exerciseInstances,
	}
}
