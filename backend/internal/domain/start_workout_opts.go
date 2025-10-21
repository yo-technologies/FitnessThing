package domain

import "fitness-trainer/internal/utils"

type StartWorkoutOpts struct {
	RoutineID utils.Nullable[ID]
}
