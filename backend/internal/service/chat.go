package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	openai_client "fitness-trainer/internal/clients/openai"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/utils"

	"github.com/openai/openai-go"
	"github.com/opentracing/opentracing-go"
)

const (
	chatHistoryLimit       = 200
	maxChatCompletionLoops = 25
	assistantErrorMessage  = "Не удалось ответить. Попробуйте ещё раз чуть позже."
)

const defaultChatSystemPrompt = `Ты — виртуальный фитнес-тренер и генератор тренировок.

Твои задачи:
- Вежливо поддерживать диалог на русском языке.
- Выяснять цели, ограничения и предпочтения пользователя.
- Использовать функции-инструменты, когда нужно получить факты из базы (история, упражнения, план тренировки) или внести изменения в тренировку. Четко следуй описанию инструментов.
- Никогда не выдумывай данные и не давай медицинских рекомендаций.
- Если нужно изменить план, всегда сначала получи текущее состояние тренировки, затем примени соответствующий инструмент.
- После выполнения инструмента поясни пользователю, что было сделано, и предложи следующие шаги.
- По результату генерации плана всегда напоминай о разминке и восстановлении.`

type chatSession struct {
	chat     domain.Chat
	messages []openai.ChatCompletionMessageParamUnion
	toolDefs []openai.ChatCompletionToolParam
}

func (s *Service) SendChatMessageStream(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest, callbacks dto.ChatStreamCallbacks) (dto.ChatCompletionDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.SendChatMessageStream")
	defer span.Finish()

	session, err := s.startChatSession(ctx, userID, req)
	if err != nil {
		return dto.ChatCompletionDTO{}, err
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, len(session.messages))
	copy(messages, session.messages)
	toolDefs := session.toolDefs
	chat := session.chat
	agentCtx := agentToolContext{userID: userID, workoutID: chat.WorkoutID}

	notifyStatus := func(status string) error {
		if callbacks.OnStatus != nil {
			return callbacks.OnStatus(status)
		}
		return nil
	}

	notifyUsage := func(usage dto.ChatUsage) error {
		if callbacks.OnUsage != nil {
			return callbacks.OnUsage(usage)
		}
		return nil
	}

	notifyFinal := func(result dto.ChatCompletionDTO) error {
		if callbacks.OnFinalResponse != nil {
			return callbacks.OnFinalResponse(result)
		}
		return nil
	}

	if err := notifyStatus("assistant_thinking"); err != nil {
		return dto.ChatCompletionDTO{}, err
	}

	var (
		totalUsage         dto.ChatUsage
		assistantResponded bool
		finalResult        dto.ChatCompletionDTO
	)

	for range maxChatCompletionLoops {
		params := s.newChatCompletionParams(messages, toolDefs, s.openAIModel, true)

		stream, err := s.openAIClient.CreateChatCompletionStream(ctx, params)
		if err != nil {
			_, _ = s.handleAssistantFailure(ctx, chat.ID, err)
			return dto.ChatCompletionDTO{}, err
		}

		assistantMessage, usage, err := s.runStreamingChatCompletion(stream, &callbacks)
		if err != nil {
			_, _ = s.handleAssistantFailure(ctx, chat.ID, err)
			return dto.ChatCompletionDTO{}, err
		}

		totalUsage.PromptTokens += usage.PromptTokens
		totalUsage.CompletionTokens += usage.CompletionTokens
		totalUsage.TotalTokens += usage.TotalTokens

		if len(assistantMessage.ToolCalls) > 0 {
			if err := s.handleAssistantToolCalls(ctx, chat.ID, &messages, agentCtx, assistantMessage, &callbacks); err != nil {
				return dto.ChatCompletionDTO{}, err
			}
			continue
		}

		finalRecord := domain.NewChatMessage(
			chat.ID,
			domain.ChatMessageRoleAssistant,
			assistantMessage.Content,
			utils.Nullable[string]{},
			utils.Nullable[string]{},
			nil,
		)
		if totalUsage.TotalTokens > 0 {
			finalRecord.TokenUsage = utils.NewNullable(totalUsage.TotalTokens, true)
		}

		if _, err := s.repository.CreateChatMessage(ctx, finalRecord); err != nil {
			return dto.ChatCompletionDTO{}, fmt.Errorf("failed to save assistant message: %w", err)
		}

		updatedHistory, err := s.repository.ListChatMessages(ctx, chat.ID, chatHistoryLimit, 0)
		if err != nil {
			return dto.ChatCompletionDTO{}, fmt.Errorf("failed to refresh chat history: %w", err)
		}

		finalResult = dto.ChatCompletionDTO{Chat: chat, Messages: updatedHistory}
		if totalUsage.TotalTokens > 0 {
			finalResult.Usage = utils.NewNullable(totalUsage, true)
		}
		assistantResponded = true

		if err := notifyFinal(finalResult); err != nil {
			return dto.ChatCompletionDTO{}, err
		}

		break
	}

	if !assistantResponded {
		return dto.ChatCompletionDTO{}, fmt.Errorf("assistant did not provide a final response")
	}

	if totalUsage.TotalTokens > 0 {
		if err := notifyUsage(totalUsage); err != nil {
			return dto.ChatCompletionDTO{}, err
		}
	}

	if err := notifyStatus("assistant_completed"); err != nil {
		return dto.ChatCompletionDTO{}, err
	}

	return finalResult, nil
}

func (s *Service) startChatSession(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest) (chatSession, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.startChatSession")
	defer span.Finish()
	span.SetTag("user_id", userID.String())
	if req.ChatID.IsValid {
		span.SetTag("chat_id.requested", req.ChatID.V.String())
	}
	if req.WorkoutID.IsValid {
		span.SetTag("workout_id.requested", req.WorkoutID.V.String())
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return chatSession{}, fmt.Errorf("message content cannot be empty: %w", domain.ErrInvalidArgument)
	}

	chat, err := s.ensureChatForRequest(ctx, userID, req)
	if err != nil {
		return chatSession{}, err
	}
	span.SetTag("chat_id", chat.ID.String())
	if chat.WorkoutID.IsValid {
		span.SetTag("workout_id", chat.WorkoutID.V.String())
	}

	userMessage := domain.NewChatMessage(
		chat.ID,
		domain.ChatMessageRoleUser,
		content,
		utils.Nullable[string]{},
		utils.Nullable[string]{},
		nil,
	)

	if _, err := s.repository.CreateChatMessage(ctx, userMessage); err != nil {
		return chatSession{}, fmt.Errorf("failed to save user chat message: %w", err)
	}

	systemMessages, err := s.buildChatSystemMessages(ctx, userID, chat)
	if err != nil {
		return chatSession{}, err
	}

	history, err := s.repository.ListChatMessages(ctx, chat.ID, chatHistoryLimit, 0)
	if err != nil {
		return chatSession{}, fmt.Errorf("failed to load chat history: %w", err)
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(systemMessages)+len(history)+4)
	messages = append(messages, systemMessages...)

	for _, msg := range history {
		param, err := s.chatMessageToOpenAIParam(msg)
		if err != nil {
			return chatSession{}, err
		}
		messages = append(messages, param)
	}

	return chatSession{
		chat:     chat,
		messages: messages,
		toolDefs: s.chatToolDefinitions(),
	}, nil
}

func (s *Service) runStreamingChatCompletion(
	stream openai_client.ChatCompletionStream,
	callbacks *dto.ChatStreamCallbacks,
) (openai.ChatCompletionMessage, dto.ChatUsage, error) {
	defer stream.Close()

	var (
		contentBuilder strings.Builder
		totalUsage     dto.ChatUsage
		usageRecorded  bool
	)

	type toolCallAccumulator struct {
		id        string
		typ       openai.ChatCompletionMessageToolCallType
		name      string
		arguments strings.Builder
	}

	toolCalls := make(map[int64]*toolCallAccumulator)

	for stream.Next() {
		chunk := stream.Chunk()

		if !usageRecorded && (chunk.Usage.TotalTokens != 0 || chunk.Usage.PromptTokens != 0 || chunk.Usage.CompletionTokens != 0) {
			totalUsage = dto.ChatUsage{
				PromptTokens:     int(chunk.Usage.PromptTokens),
				CompletionTokens: int(chunk.Usage.CompletionTokens),
				TotalTokens:      int(chunk.Usage.TotalTokens),
			}
			usageRecorded = true
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		delta := choice.Delta

		if delta.Content != "" {
			contentBuilder.WriteString(delta.Content)
			if callbacks != nil && callbacks.OnContentDelta != nil {
				if err := callbacks.OnContentDelta(delta.Content); err != nil {
					return openai.ChatCompletionMessage{}, totalUsage, err
				}
			}
		}

		for _, toolCall := range delta.ToolCalls {
			acc := toolCalls[toolCall.Index]
			if acc == nil {
				acc = &toolCallAccumulator{}
				toolCalls[toolCall.Index] = acc
			}
			if toolCall.ID != "" {
				acc.id = toolCall.ID
			}
			if toolCall.Type != "" {
				acc.typ = openai.ChatCompletionMessageToolCallType(toolCall.Type)
			}
			if toolCall.Function.Name != "" {
				acc.name = toolCall.Function.Name
			}
			if toolCall.Function.Arguments != "" {
				acc.arguments.WriteString(toolCall.Function.Arguments)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return openai.ChatCompletionMessage{}, totalUsage, err
	}

	finalMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatCompletionMessageRoleAssistant,
		Content: contentBuilder.String(),
	}

	if len(toolCalls) > 0 {
		indices := make([]int, 0, len(toolCalls))
		for idx := range toolCalls {
			indices = append(indices, int(idx))
		}
		sort.Ints(indices)

		finalCalls := make([]openai.ChatCompletionMessageToolCall, 0, len(indices))
		for _, idx := range indices {
			acc := toolCalls[int64(idx)]
			typ := acc.typ
			if typ == "" {
				typ = openai.ChatCompletionMessageToolCallTypeFunction
			}
			finalCalls = append(finalCalls, openai.ChatCompletionMessageToolCall{
				ID:   acc.id,
				Type: typ,
				Function: openai.ChatCompletionMessageToolCallFunction{
					Name:      acc.name,
					Arguments: acc.arguments.String(),
				},
			})
		}
		finalMessage.ToolCalls = finalCalls
	}

	return finalMessage, totalUsage, nil
}

func (s *Service) ensureChatForRequest(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest) (domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.ensureChatForRequest")
	defer span.Finish()
	span.SetTag("user_id", userID.String())
	if req.ChatID.IsValid {
		span.SetTag("chat_id.requested", req.ChatID.V.String())
	}
	if req.WorkoutID.IsValid {
		span.SetTag("workout_id.requested", req.WorkoutID.V.String())
	}
	if req.ChatID.IsValid {
		chat, err := s.repository.GetChatByID(ctx, req.ChatID.V)
		if err != nil {
			return domain.Chat{}, err
		}
		if chat.UserID != userID {
			return domain.Chat{}, domain.ErrForbidden
		}
		if req.WorkoutID.IsValid && (!chat.WorkoutID.IsValid || chat.WorkoutID.V != req.WorkoutID.V) {
			return domain.Chat{}, domain.ErrInvalidArgument
		}
		span.SetTag("chat_id", chat.ID.String())
		if chat.WorkoutID.IsValid {
			span.SetTag("workout_id", chat.WorkoutID.V.String())
		}
		return chat, nil
	}

	if req.WorkoutID.IsValid {
		workout, err := s.repository.GetWorkoutByID(ctx, req.WorkoutID.V)
		if err != nil {
			return domain.Chat{}, err
		}
		if workout.UserID != userID {
			return domain.Chat{}, domain.ErrForbidden
		}

		chat, err := s.repository.GetChatByWorkoutID(ctx, req.WorkoutID.V)
		if err == nil {
			span.SetTag("chat_id", chat.ID.String())
			if chat.WorkoutID.IsValid {
				span.SetTag("workout_id", chat.WorkoutID.V.String())
			}
			return chat, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return domain.Chat{}, err
		}

		title := fmt.Sprintf("Тренировка %s", workout.CreatedAt.Format("02.01.2006"))
		newChat := domain.NewChat(userID, utils.NewNullable(req.WorkoutID.V, true), title)
		createdChat, err := s.repository.CreateChat(ctx, newChat)
		if err == nil {
			span.SetTag("chat_id", createdChat.ID.String())
			if createdChat.WorkoutID.IsValid {
				span.SetTag("workout_id", createdChat.WorkoutID.V.String())
			}
		}
		return createdChat, err
	}

	title := fmt.Sprintf("Чат %s", time.Now().Format("02.01.2006 15:04"))
	newChat := domain.NewChat(userID, utils.Nullable[domain.ID]{}, title)
	createdChat, err := s.repository.CreateChat(ctx, newChat)
	if err == nil {
		span.SetTag("chat_id", createdChat.ID.String())
	}
	return createdChat, err
}

func (s *Service) buildChatSystemMessages(ctx context.Context, userID domain.ID, chat domain.Chat) ([]openai.ChatCompletionMessageParamUnion, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.buildChatSystemMessages")
	defer span.Finish()
	span.SetTag("user_id", userID.String())
	span.SetTag("chat_id", chat.ID.String())
	if chat.WorkoutID.IsValid {
		span.SetTag("workout_id", chat.WorkoutID.V.String())
	}
	builder := strings.Builder{}
	builder.WriteString(defaultChatSystemPrompt)

	prompt, err := s.repository.GetLastPromptByUserID(ctx, userID)
	if err == nil && prompt.PromptText != "" {
		builder.WriteString("\n\nЛичные пожелания пользователя: ")
		builder.WriteString(prompt.PromptText)
	} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("failed to load generation settings: %w", err)
	}

	messages := []openai.ChatCompletionMessageParamUnion{openai.SystemMessage(builder.String())}

	if chat.WorkoutID.IsValid {
		messages = append(messages, openai.SystemMessage("Этот чат привязан к конкретной тренировке. Используй инструменты для просмотра плана и внесения правок."))
	}

	return messages, nil
}

func (s *Service) chatMessageToOpenAIParam(message domain.ChatMessage) (openai.ChatCompletionMessageParamUnion, error) {
	switch message.Role {
	case domain.ChatMessageRoleUser:
		return openai.UserMessage(message.Content), nil
	case domain.ChatMessageRoleAssistant:
		assistant := openai.AssistantMessage(message.Content)
		if message.ToolCallID.IsValid && message.ToolName.IsValid {
			argsJSON := "{}"
			if len(message.ToolArguments) > 0 {
				raw, err := json.Marshal(message.ToolArguments)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool arguments: %w", err)
				}
				argsJSON = string(raw)
			}

			assistant.ToolCalls = openai.F([]openai.ChatCompletionMessageToolCallParam{
				{
					ID:   openai.F(message.ToolCallID.V),
					Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),
					Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      openai.F(message.ToolName.V),
						Arguments: openai.F(argsJSON),
					}),
				},
			})
		}
		return assistant, nil
	case domain.ChatMessageRoleTool:
		if !message.ToolCallID.IsValid {
			return nil, fmt.Errorf("tool message missing tool call id")
		}
		return openai.ToolMessage(message.ToolCallID.V, message.Content), nil
	case domain.ChatMessageRoleSystem:
		return openai.SystemMessage(message.Content), nil
	default:
		return nil, fmt.Errorf("unsupported chat message role: %s", message.Role)
	}
}

func (s *Service) handleAssistantToolCalls(
	ctx context.Context,
	chatID domain.ID,
	messages *[]openai.ChatCompletionMessageParamUnion,
	toolCtx agentToolContext,
	assistantMessage openai.ChatCompletionMessage,
	callbacks *dto.ChatStreamCallbacks,
) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.handleAssistantToolCalls")
	defer span.Finish()
	span.SetTag("chat_id", chatID.String())
	span.SetTag("tool_call_count", len(assistantMessage.ToolCalls))

	var statusFn func(string) error
	if callbacks != nil {
		statusFn = callbacks.OnStatus
	}

	// Сохраняем отдельный assistant message для каждого tool call
	// Это упрощает восстановление истории и UI логику
	contentForFirst := assistantMessage.Content

	for idx, toolCall := range assistantMessage.ToolCalls {
		toolName := toolCall.Function.Name
		toolArgsJSON := toolCall.Function.Arguments
		span.LogKV("event", "tool_call_start", "tool", toolName, "index", idx)

		var toolArgs map[string]any
		trimmedArgs := strings.TrimSpace(toolArgsJSON)
		if trimmedArgs != "" {
			if err := json.Unmarshal([]byte(trimmedArgs), &toolArgs); err != nil {
				span.SetTag("error", true)
				span.LogKV("event", "json_parse_error", "tool", toolName, "raw_args", trimmedArgs, "error", err.Error())
				return fmt.Errorf("failed to parse tool arguments for %s: %w (raw: %q)", toolName, err, trimmedArgs)
			}
		}

		assistantRecord := domain.NewChatMessage(
			chatID,
			domain.ChatMessageRoleAssistant,
			"",
			utils.Nullable[string]{},
			utils.Nullable[string]{},
			toolArgs,
		)

		// Только первый assistant message содержит текстовый контент
		if idx == 0 {
			assistantRecord.Content = contentForFirst
		}

		if toolName != "" {
			assistantRecord.ToolName = utils.NewNullable(toolName, true)
		}
		if toolCall.ID != "" {
			assistantRecord.ToolCallID = utils.NewNullable(toolCall.ID, true)
		}

		savedAssistant, err := s.repository.CreateChatMessage(ctx, assistantRecord)
		if err != nil {
			return fmt.Errorf("failed to persist assistant tool call: %w", err)
		}

		param, err := s.chatMessageToOpenAIParam(savedAssistant)
		if err != nil {
			return err
		}
		*messages = append(*messages, param)

		if statusFn != nil {
			if err := statusFn(fmt.Sprintf("invoking tool %s", toolName)); err != nil {
				return err
			}
		}

		result, err := s.executeTool(ctx, toolCtx, toolName, toolArgsJSON)
		if err != nil {
			span.SetTag("error", true)
			span.LogKV("event", "tool_call_error", "tool", toolName, "index", idx, "error.object", err)
			_, failureErr := s.handleAssistantFailure(ctx, chatID, fmt.Errorf("tool %s error: %w", toolName, err))
			if failureErr != nil {
				return failureErr
			}
			return err
		}

		toolRecord := domain.NewChatMessage(
			chatID,
			domain.ChatMessageRoleTool,
			result,
			utils.Nullable[string]{},
			utils.Nullable[string]{},
			nil,
		)

		if toolName != "" {
			toolRecord.ToolName = utils.NewNullable(toolName, true)
		}
		if toolCall.ID != "" {
			toolRecord.ToolCallID = utils.NewNullable(toolCall.ID, true)
		}

		savedTool, err := s.repository.CreateChatMessage(ctx, toolRecord)
		if err != nil {
			return fmt.Errorf("failed to persist tool message: %w", err)
		}

		toolParam, err := s.chatMessageToOpenAIParam(savedTool)
		if err != nil {
			return err
		}
		*messages = append(*messages, toolParam)

		if statusFn != nil {
			if err := statusFn(fmt.Sprintf("tool %s completed", toolName)); err != nil {
				return err
			}
		}

		span.LogKV("event", "tool_call_success", "tool", toolName, "index", idx)
	}

	return nil
}

func (s *Service) handleAssistantFailure(ctx context.Context, chatID domain.ID, originalErr error) (dto.ChatCompletionDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.handleAssistantFailure")
	defer span.Finish()
	span.SetTag("chat_id", chatID.String())
	if originalErr != nil {
		span.SetTag("error", true)
		span.LogKV("event", "assistant_failure", "error.object", originalErr)
	}
	errMsg := domain.NewChatMessage(
		chatID,
		domain.ChatMessageRoleAssistant,
		assistantErrorMessage,
		utils.Nullable[string]{},
		utils.Nullable[string]{},
		nil,
	)
	errMsg.Error = utils.NewNullable(originalErr.Error(), true)

	if _, err := s.repository.CreateChatMessage(ctx, errMsg); err != nil {
		return dto.ChatCompletionDTO{}, fmt.Errorf("failed to persist assistant error message: %w", err)
	}

	return dto.ChatCompletionDTO{}, errors.Join(domain.ErrInternal, originalErr)
}

func (s *Service) GetChat(ctx context.Context, userID domain.ID, req dto.GetChatRequest) (dto.GetChatDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetChat")
	defer span.Finish()

	if !req.ChatID.IsValid && !req.WorkoutID.IsValid {
		return dto.GetChatDTO{}, fmt.Errorf("either chat_id or workout_id must be provided")
	}

	var chat domain.Chat
	var err error

	// Определяем, как искать чат: по workout_id или по chat_id
	if req.ChatID.IsValid {
		chat, err = s.repository.GetChatByID(ctx, req.ChatID.V)
		if err != nil {
			return dto.GetChatDTO{}, fmt.Errorf("failed to get chat by id: %w", err)
		}
	}
	if req.WorkoutID.IsValid {
		chat, err = s.repository.GetChatByWorkoutID(ctx, req.WorkoutID.V)
		if err != nil {
			return dto.GetChatDTO{}, fmt.Errorf("failed to get chat by workout id: %w", err)
		}
	}

	// Проверяем, что пользователь имеет доступ к этому чату
	if chat.UserID != userID {
		return dto.GetChatDTO{}, domain.ErrForbidden
	}

	messages, err := s.repository.ListChatMessages(ctx, chat.ID, 1000, 0) // limit 1000, offset 0
	if err != nil {
		return dto.GetChatDTO{}, fmt.Errorf("failed to list chat messages: %w", err)
	}

	return dto.GetChatDTO{
		Chat:     chat,
		Messages: messages,
	}, nil
}
