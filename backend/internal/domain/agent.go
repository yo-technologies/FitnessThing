package domain

import "fitness-trainer/internal/utils"

type AgentChatContext struct {
	UserID    ID
	WorkoutID utils.Nullable[ID]
}

func NewAgentChatContext(userID ID, workoutID utils.Nullable[ID]) AgentChatContext {
	return AgentChatContext{
		UserID:    userID,
		WorkoutID: workoutID,
	}
}
