package chat

import (
	"fmt"

	"fitness-trainer/internal/app/interceptors"
	"fitness-trainer/internal/app/mappers"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"
	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) SendChatMessageStream(in *desc.SendChatMessageRequest, stream desc.ChatService_SendChatMessageStreamServer) error {
	ctx := stream.Context()
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.SendChatMessageStream")
	defer span.Finish()

	if in == nil {
		return fmt.Errorf("%w: empty request", domain.ErrInvalidArgument)
	}

	if err := in.ValidateAll(); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		logger.Errorf("error getting user id from context")
		return domain.ErrInternal
	}

	req := dto.SendChatMessageRequest{
		Content: in.GetContent(),
	}

	if chatID := in.GetChatId(); chatID != "" {
		parsed, err := domain.ParseID(chatID)
		if err != nil {
			return fmt.Errorf("%w: invalid chat_id", domain.ErrInvalidArgument)
		}
		req.ChatID = utils.NewNullable(parsed, true)
	}

	if workoutID := in.GetWorkoutId(); workoutID != "" {
		parsed, err := domain.ParseID(workoutID)
		if err != nil {
			return fmt.Errorf("%w: invalid workout_id", domain.ErrInvalidArgument)
		}
		req.WorkoutID = utils.NewNullable(parsed, true)
	}

	send := func(resp *desc.SendChatMessageStreamResponse) error {
		if resp == nil {
			return nil
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
		return nil
	}

	callbacks := dto.ChatStreamCallbacks{
		OnContentDelta: func(delta string) error {
			if delta == "" {
				return nil
			}
			return send(&desc.SendChatMessageStreamResponse{
				Payload: &desc.SendChatMessageStreamResponse_MessageDelta{
					MessageDelta: &desc.ChatMessageDelta{Content: delta},
				},
			})
		},
		OnUsage: func(usage dto.ChatUsage) error {
			return send(&desc.SendChatMessageStreamResponse{
				Payload: &desc.SendChatMessageStreamResponse_Usage{
					Usage: mappers.ChatUsageToProto(usage),
				},
			})
		},
		OnStatus: func(status string) error {
			if status == "" {
				return nil
			}
			return send(&desc.SendChatMessageStreamResponse{
				Payload: &desc.SendChatMessageStreamResponse_Status{Status: status},
			})
		},
		OnToolEvent: func(ev dto.ToolEvent) error {
			state := desc.ToolEvent_STATE_UNSPECIFIED
			switch ev.State {
			case dto.ToolInvoking:
				state = desc.ToolEvent_INVOKING
			case dto.ToolCompleted:
				state = desc.ToolEvent_COMPLETED
			case dto.ToolError:
				state = desc.ToolEvent_ERROR
			}

			var errPtr *string
			if ev.Error != "" {
				err := ev.Error
				errPtr = &err
			}
			return send(&desc.SendChatMessageStreamResponse{
				Payload: &desc.SendChatMessageStreamResponse_ToolEvent{ToolEvent: &desc.ToolEvent{
					ToolName:   ev.ToolName,
					ToolCallId: ev.ToolCallID,
					ArgsJson:   ev.ArgsJSON,
					State:      state,
					Error:      errPtr,
				}},
			})
		},
		OnError: func(err error) error {
			// Map error to ChatError payload
			var (
				typ               = domain.ErrorTypeInternal
				code              string
				retryAfterSeconds int32
			)
			if te, ok := err.(*domain.TypedError); ok {
				typ = te.Type
				code = te.Code
				if te.RetryAfter != nil {
					retryAfterSeconds = int32(te.RetryAfter.Seconds())
				}
			}
			return send(&desc.SendChatMessageStreamResponse{
				Payload: &desc.SendChatMessageStreamResponse_Error{
					Error: &desc.ChatError{
						Type:              string(typ),
						Message:           err.Error(),
						RetryAfterSeconds: &retryAfterSeconds,
						Code:              &code,
					},
				},
			})
		},
		OnFinalResponse: func(result dto.ChatCompletionDTO) error {
			protoChat := mappers.ChatToProto(result.Chat)
			protoMessages, err := mappers.ChatMessagesToProto(result.Messages)
			if err != nil {
				logger.Errorf("failed to map final chat messages: %v", err)
				return domain.ErrInternal
			}

			finalResp := &desc.SendChatMessageResponse{
				Chat:     protoChat,
				Messages: protoMessages,
			}

			if result.Usage.IsValid {
				finalResp.Usage = mappers.ChatUsageToProto(result.Usage.V)
			}

			return send(&desc.SendChatMessageStreamResponse{
				Payload: &desc.SendChatMessageStreamResponse_Final{Final: finalResp},
			})
		},
	}

	if _, err := i.service.SendChatMessageStream(ctx, userID, req, callbacks); err != nil {
		return err
	}

	return nil
}
