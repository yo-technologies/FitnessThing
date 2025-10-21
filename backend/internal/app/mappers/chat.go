package mappers

import (
	"fmt"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	desc "fitness-trainer/pkg/workouts"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ChatToProto(chat domain.Chat) *desc.Chat {
	var workoutID *string
	if chat.WorkoutID.IsValid {
		id := chat.WorkoutID.V.String()
		workoutID = &id
	}

	return &desc.Chat{
		Id:        chat.ID.String(),
		UserId:    chat.UserID.String(),
		WorkoutId: workoutID,
		Title:     chat.Title,
		CreatedAt: timestamppb.New(chat.CreatedAt),
		UpdatedAt: timestamppb.New(chat.UpdatedAt),
	}
}

func ChatMessagesToProto(messages []domain.ChatMessage) ([]*desc.ChatMessage, error) {
	result := make([]*desc.ChatMessage, 0, len(messages))
	for _, message := range messages {
		protoMessage, err := ChatMessageToProto(message)
		if err != nil {
			return nil, err
		}
		result = append(result, protoMessage)
	}
	return result, nil
}

func ChatMessageToProto(message domain.ChatMessage) (*desc.ChatMessage, error) {
	var (
		toolName   *string
		toolCallID *string
		tokenUsage *int32
		errText    *string
		toolArgs   *structpb.Struct
		err        error
	)

	if message.ToolName.IsValid {
		value := message.ToolName.V
		toolName = &value
	}

	if message.ToolCallID.IsValid {
		value := message.ToolCallID.V
		toolCallID = &value
	}

	if message.TokenUsage.IsValid {
		value := int32(message.TokenUsage.V)
		tokenUsage = &value
	}

	if message.Error.IsValid {
		value := message.Error.V
		errText = &value
	}

	if message.ToolArguments != nil {
		toolArgs, err = structpb.NewStruct(message.ToolArguments)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool arguments to struct: %w", err)
		}
	}

	return &desc.ChatMessage{
		Id:            message.ID.String(),
		ChatId:        message.ChatID.String(),
		Role:          chatMessageRoleToProto(message.Role),
		Content:       message.Content,
		ToolName:      toolName,
		ToolCallId:    toolCallID,
		ToolArguments: toolArgs,
		TokenUsage:    tokenUsage,
		Error:         errText,
		CreatedAt:     timestamppb.New(message.CreatedAt),
		UpdatedAt:     timestamppb.New(message.UpdatedAt),
	}, nil
}

func ChatUsageToProto(usage dto.ChatUsage) *desc.ChatUsage {
	return &desc.ChatUsage{
		PromptTokens:     int32(usage.PromptTokens),
		CompletionTokens: int32(usage.CompletionTokens),
		TotalTokens:      int32(usage.TotalTokens),
	}
}

func chatMessageRoleToProto(role domain.ChatMessageRole) desc.ChatMessageRole {
	switch role {
	case domain.ChatMessageRoleUser:
		return desc.ChatMessageRole_CHAT_MESSAGE_ROLE_USER
	case domain.ChatMessageRoleAssistant:
		return desc.ChatMessageRole_CHAT_MESSAGE_ROLE_ASSISTANT
	case domain.ChatMessageRoleTool:
		return desc.ChatMessageRole_CHAT_MESSAGE_ROLE_TOOL
	case domain.ChatMessageRoleSystem:
		return desc.ChatMessageRole_CHAT_MESSAGE_ROLE_SYSTEM
	default:
		return desc.ChatMessageRole_CHAT_MESSAGE_ROLE_UNSPECIFIED
	}
}
