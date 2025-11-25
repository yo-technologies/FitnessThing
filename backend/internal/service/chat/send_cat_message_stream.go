package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	appconfig "fitness-trainer/internal/config"
	"fitness-trainer/internal/llm"
	"fitness-trainer/internal/logger"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/utils"

	"github.com/opentracing/opentracing-go"
)

const (
	chatHistoryLimit       = 200
	maxChatCompletionLoops = 25
	assistantErrorMessage  = "Не удалось ответить. Попробуйте ещё раз чуть позже."
)

const defaultChatSystemPrompt = `Ты — FitnessThing, персональный AI-тренер. Твоя главная задача — составлять эффективные тренировки для пользователя, учитывая его индивидуальные особенности и прогресс.

## Основные принципы работы

### Сбор информации
Перед составлением тренировки ОБЯЗАТЕЛЬНО используй инструменты для получения:
- Истории тренировок пользователя за последние 7-14 дней
- Данных о прогрессе (веса, повторения, объём нагрузки)

### Приоритет групп мышц
Приоритет групп мышц рассматривай на масштабе НЕДЕЛИ, а не одной тренировки:
- Проанализируй, какие группы мышц были нагружены за последние 7 дней
- Определи, какие приоритетные группы недополучили нагрузку
- Сбалансируй текущую тренировку так, чтобы за неделю приоритетные группы получили больше внимания
- Учитывай время восстановления: не нагружай одну группу два дня подряд

### Составление тренировки
При составлении тренировки:
1. Сначала получи всю необходимую информацию через инструменты
2. Проанализируй историю и определи оптимальную нагрузку
3. Используй инструменты для добавления упражнений в тренировку
4. Устанавливай веса и повторения на основе прогресса пользователя
5. Применяй принцип прогрессивной перегрузки (постепенное увеличение нагрузки)

### Прогрессия нагрузки
- Если пользователь успешно выполнил предыдущую тренировку — предложи небольшое увеличение (вес +2.5-5%, повторения +1-2)
- Если были сложности — сохрани текущий уровень или немного снизь
- Учитывай обратную связь пользователя о самочувствии

### Стиль общения
- Будь крайне дружелюбным, поддерживающим и мотивирующим
- Всегда объясняй свои решения кратко и понятно
- Мотивируй пользователя, отмечай прогресс
- Отвечай на русском языке

### Безопасность
- Не рекомендуй упражнения, которые могут быть опасны без присмотра
- При упоминании травм или болей — советуй консультацию с врачом
- Учитывай ограничения, указанные пользователем

## Важно
Все данные о пользователе, его тренировках и настройках получай ТОЛЬКО через предоставленные инструменты. Не придумывай информацию. Если данных недостаточно — запроси их у пользователя или через инструменты.`

type chatSession struct {
	chat     domain.Chat
	messages []llm.MessageParam
	toolDefs []llm.ToolDefinition
}

func (s *Service) SendChatMessageStream(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest, callbacks dto.ChatStreamCallbacks) (dto.ChatCompletionDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.SendChatMessageStream")
	defer span.Finish()

	session, err := s.startChatSession(ctx, userID, req)
	if err != nil {
		return dto.ChatCompletionDTO{}, err
	}

	messages := make([]llm.MessageParam, len(session.messages))
	copy(messages, session.messages)
	toolDefs := session.toolDefs
	chat := session.chat
	agentCtx := domain.NewAgentChatContext(userID, chat.WorkoutID)

	// Reserve 1 token upfront for admission control
	allowed, err := s.quotaService.Reserve(ctx, userID, 1)
	if err != nil {
		return dto.ChatCompletionDTO{}, err
	}
	if !allowed {
		// Return a typed quota-exceeded error and try to emit a structured error event
		te := domain.QuotaExceededError(domain.ErrTooManyRequests)
		if callbacks.OnError != nil {
			_ = callbacks.OnError(te)
		}

		_, _ = s.handleAssistantFailure(ctx, chat.ID, te)
		return dto.ChatCompletionDTO{}, te
	}

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
		params := llm.ChatParams{Messages: messages, Tools: toolDefs, IncludeUsage: true}

		stream, err := s.llmClient.CreateCompletionStream(ctx, params)
		if err != nil {
			// Try to send structured error event before persisting assistant error
			if callbacks.OnError != nil {
				_ = callbacks.OnError(err)
			}
			_, _ = s.handleAssistantFailure(ctx, chat.ID, err)
			return dto.ChatCompletionDTO{}, err
		}

		assistantMessage, usage, err := s.runStreamingChatCompletion(ctx, stream, &callbacks)
		if err != nil {
			if callbacks.OnError != nil {
				_ = callbacks.OnError(err)
			}
			_, handlerErr := s.handleAssistantFailure(ctx, chat.ID, err)
			if handlerErr != nil {
				logger.Errorf("error handling assistant failure: %v", handlerErr)
			}

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

		if _, err := s.chatRepository.CreateChatMessage(ctx, finalRecord); err != nil {
			return dto.ChatCompletionDTO{}, fmt.Errorf("failed to save assistant message: %w", err)
		}

		updatedHistory, err := s.chatRepository.ListChatMessages(ctx, chat.ID, chatHistoryLimit, 0)
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

	// Finalize usage: confirm with weighted tokens
	actualUnits := totalUsage.TotalTokens
	if totalUsage.PromptTokens > 0 || totalUsage.CompletionTokens > 0 {
		cfg := appconfig.Get()
		wp := cfg.GetLLMPromptWeight()
		wc := cfg.GetLLMCompletionWeight()
		weighted := float64(totalUsage.PromptTokens)*wp + float64(totalUsage.CompletionTokens)*wc
		actualUnits = int(math.Round(weighted))
	}

	err = s.quotaService.Confirm(ctx, userID, 1, actualUnits)
	if err != nil {
		return dto.ChatCompletionDTO{}, fmt.Errorf("error confirming quota: %w", err)
	}

	return finalResult, nil
}

func (s *Service) startChatSession(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest) (chatSession, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.startChatSession")
	defer span.Finish()

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return chatSession{}, fmt.Errorf("message content cannot be empty: %w", domain.ErrInvalidArgument)
	}

	chat, err := s.ensureChatForRequest(ctx, userID, req)
	if err != nil {
		return chatSession{}, err
	}

	userMessage := domain.NewChatMessage(
		chat.ID,
		domain.ChatMessageRoleUser,
		content,
		utils.Nullable[string]{},
		utils.Nullable[string]{},
		nil,
	)

	if _, err := s.chatRepository.CreateChatMessage(ctx, userMessage); err != nil {
		return chatSession{}, fmt.Errorf("failed to save user chat message: %w", err)
	}

	systemMessages, err := s.buildChatSystemMessages(ctx, userID, chat)
	if err != nil {
		return chatSession{}, err
	}

	history, err := s.chatRepository.ListChatMessages(ctx, chat.ID, chatHistoryLimit, 0)
	if err != nil {
		return chatSession{}, fmt.Errorf("failed to load chat history: %w", err)
	}

	messages := make([]llm.MessageParam, 0, len(systemMessages)+len(history)+4)
	messages = append(messages, systemMessages...)

	for _, msg := range history {
		param, err := s.chatMessageToLLMParam(msg)
		if err != nil {
			return chatSession{}, fmt.Errorf("failed to convert chat message to OpenAI param: %w", err)
		}
		messages = append(messages, param)
	}

	return chatSession{
		chat:     chat,
		messages: messages,
		toolDefs: s.toolsService.ChatAgentToolDefinitions(),
	}, nil
}

func (s *Service) runStreamingChatCompletion(
	ctx context.Context,
	stream llm.ChatStream,
	callbacks *dto.ChatStreamCallbacks,
) (llm.ChatMessage, dto.ChatUsage, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "service.chat.runStreamingChatCompletion")
	defer span.Finish()

	defer stream.Close()

	var (
		contentBuilder strings.Builder
		totalUsage     dto.ChatUsage
		usageRecorded  bool
	)

	type toolCallAccumulator struct {
		id        string
		name      string
		arguments strings.Builder
	}

	toolCalls := make(map[int]*toolCallAccumulator)

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

		if chunk.Content != "" {
			contentBuilder.WriteString(chunk.Content)
			if callbacks != nil && callbacks.OnContentDelta != nil {
				if err := callbacks.OnContentDelta(chunk.Content); err != nil {
					return llm.ChatMessage{}, totalUsage, err
				}
			}
		}

		for _, toolCall := range chunk.ToolCalls {
			acc := toolCalls[toolCall.Index]
			if acc == nil {
				acc = &toolCallAccumulator{}
				toolCalls[toolCall.Index] = acc
			}
			if toolCall.ID != "" {
				acc.id = toolCall.ID
			}
			if toolCall.Name != "" {
				acc.name = toolCall.Name
			}
			if toolCall.Arguments != "" {
				acc.arguments.WriteString(toolCall.Arguments)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return llm.ChatMessage{}, totalUsage, err
	}

	finalMessage := llm.ChatMessage{
		Role:    llm.RoleAssistant,
		Content: contentBuilder.String(),
	}

	if len(toolCalls) > 0 {
		indices := make([]int, 0, len(toolCalls))
		for idx := range toolCalls {
			indices = append(indices, idx)
		}
		sort.Ints(indices)

		finalCalls := make([]llm.ToolCall, 0, len(indices))
		for _, idx := range indices {
			acc := toolCalls[idx]
			finalCalls = append(finalCalls, llm.ToolCall{
				ID:        acc.id,
				Name:      acc.name,
				Arguments: acc.arguments.String(),
			})
		}
		finalMessage.ToolCalls = finalCalls
	}

	return finalMessage, totalUsage, nil
}

func (s *Service) ensureChatForRequest(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest) (domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.ensureChatForRequest")
	defer span.Finish()

	if req.ChatID.IsValid {
		chat, err := s.chatRepository.GetChatByID(ctx, req.ChatID.V)
		if err != nil {
			return domain.Chat{}, err
		}
		if chat.UserID != userID {
			return domain.Chat{}, domain.ErrForbidden
		}
		if req.WorkoutID.IsValid && (!chat.WorkoutID.IsValid || chat.WorkoutID.V != req.WorkoutID.V) {
			return domain.Chat{}, domain.ErrInvalidArgument
		}

		return chat, nil
	}

	if req.WorkoutID.IsValid {
		workout, err := s.workoutRepository.GetWorkoutByID(ctx, req.WorkoutID.V)
		if err != nil {
			return domain.Chat{}, err
		}
		if workout.UserID != userID {
			return domain.Chat{}, domain.ErrForbidden
		}

		chat, err := s.chatRepository.GetChatByWorkoutID(ctx, req.WorkoutID.V)
		if err == nil {
			return chat, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return domain.Chat{}, fmt.Errorf("failed to get chat by workout ID: %w", err)
		}

		title := fmt.Sprintf("Тренировка %s", workout.CreatedAt.Format("02.01.2006"))
		newChat := domain.NewChat(userID, utils.NewNullable(req.WorkoutID.V, true), title)

		createdChat, err := s.chatRepository.CreateChat(ctx, newChat)
		if err != nil {
			return domain.Chat{}, fmt.Errorf("failed to create chat for workout: %w", err)
		}

		return createdChat, nil
	}

	title := fmt.Sprintf("Чат %s", time.Now().Format("02.01.2006 15:04"))
	newChat := domain.NewChat(userID, utils.Nullable[domain.ID]{}, title)

	createdChat, err := s.chatRepository.CreateChat(ctx, newChat)
	if err != nil {
		return domain.Chat{}, fmt.Errorf("failed to create chat: %w", err)
	}

	return createdChat, nil
}

func (s *Service) buildChatSystemMessages(ctx context.Context, userID domain.ID, chat domain.Chat) ([]llm.MessageParam, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.buildChatSystemMessages")
	defer span.Finish()

	messages := []llm.MessageParam{{Role: llm.RoleSystem, Content: defaultChatSystemPrompt}}

	prompt, err := s.userPromptRepository.GetLastPromptByUserID(ctx, userID)
	if err == nil {
		messages = append(messages, llm.MessageParam{Role: llm.RoleSystem, Content: fmt.Sprintf("\n\nЛичные пожелания пользователя: %s", prompt.PromptText)})
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("failed to load generation settings: %w", err)
	}

	if chat.WorkoutID.IsValid {
		messages = append(messages, llm.MessageParam{Role: llm.RoleSystem, Content: "Этот чат привязан к конкретной тренировке. Используй инструменты для просмотра плана и внесения правок."})
	}

	return messages, nil
}

func (s *Service) chatMessageToLLMParam(message domain.ChatMessage) (llm.MessageParam, error) {
	switch message.Role {
	case domain.ChatMessageRoleUser:
		return llm.MessageParam{Role: llm.RoleUser, Content: message.Content}, nil
	case domain.ChatMessageRoleAssistant:
		// Build assistant message with optional tool call
		assistant := llm.MessageParam{Role: llm.RoleAssistant, Content: message.Content}
		if message.ToolCallID.IsValid && message.ToolName.IsValid {
			argsJSON := "{}"
			if len(message.ToolArguments) > 0 {
				raw, err := json.Marshal(message.ToolArguments)
				if err != nil {
					return llm.MessageParam{}, fmt.Errorf("failed to marshal tool arguments: %w", err)
				}
				argsJSON = string(raw)
			}
			assistant.ToolCalls = []llm.ToolCall{{ID: message.ToolCallID.V, Name: message.ToolName.V, Arguments: argsJSON}}
		}
		return assistant, nil
	case domain.ChatMessageRoleTool:
		if !message.ToolCallID.IsValid {
			return llm.MessageParam{}, fmt.Errorf("tool message missing tool call id")
		}
		content := message.Content
		if strings.TrimSpace(content) == "" && message.Error.IsValid {
			content = fmt.Sprintf("error: %s", message.Error.V)
		}
		return llm.MessageParam{Role: llm.RoleTool, Content: content, ToolCallID: message.ToolCallID.V}, nil
	case domain.ChatMessageRoleSystem:
		return llm.MessageParam{Role: llm.RoleSystem, Content: message.Content}, nil
	default:
		return llm.MessageParam{}, fmt.Errorf("unsupported chat message role: %s", message.Role)
	}
}

func (s *Service) handleAssistantToolCalls(
	ctx context.Context,
	chatID domain.ID,
	messages *[]llm.MessageParam,
	chatCtx domain.AgentChatContext,
	assistantMessage llm.ChatMessage,
	callbacks *dto.ChatStreamCallbacks,
) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.handleAssistantToolCalls")
	defer span.Finish()

	// Сохраняем отдельный assistant message для каждого tool call
	// Это упрощает восстановление истории и UI логику
	contentForFirst := assistantMessage.Content

	for idx, toolCall := range assistantMessage.ToolCalls {
		toolName := toolCall.Name
		toolArgsJSON := toolCall.Arguments

		var toolArgs map[string]any
		trimmedArgs := strings.TrimSpace(toolArgsJSON)
		if trimmedArgs != "" {
			// Попытка парсинга как есть
			if err := json.Unmarshal([]byte(trimmedArgs), &toolArgs); err != nil {
				// Если не удалось, пытаемся извлечь первый валидный JSON объект
				// (иногда OpenAI возвращает несколько объектов подряд)
				decoder := json.NewDecoder(strings.NewReader(trimmedArgs))
				if decodeErr := decoder.Decode(&toolArgs); decodeErr != nil {
					return fmt.Errorf("failed to parse tool arguments for %s: %w (raw: %q)", toolName, err, trimmedArgs)
				}
				logger.Warnf("tool %s had malformed arguments, using first valid JSON object: %v", toolName, toolArgs)
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

		savedAssistant, err := s.chatRepository.CreateChatMessage(ctx, assistantRecord)
		if err != nil {
			return fmt.Errorf("failed to persist assistant tool call: %w", err)
		}

		param, err := s.chatMessageToLLMParam(savedAssistant)
		if err != nil {
			return err
		}
		*messages = append(*messages, param)

		if callbacks.OnToolEvent != nil {
			_ = callbacks.OnToolEvent(dto.ToolEvent{
				ToolName:   toolName,
				ToolCallID: toolCall.ID,
				ArgsJSON:   toolArgsJSON,
				State:      dto.ToolInvoking,
			})
		}

		result, err := s.toolsService.ExecuteChatAgentTool(ctx, chatCtx, toolName, toolArgsJSON)
		if err != nil {
			// Не прерываем цикл: сохраняем ошибку как tool-сообщение, чтобы LLM увидела её и могла скорректировать стратегию.
			toolRecord := domain.NewChatMessage(
				chatID,
				domain.ChatMessageRoleTool,
				"",
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
			// Сохраним сам текст ошибки отдельным полем, чтобы UI мог маркировать, а LLM получала её через синтез контента
			toolRecord.Error = utils.NewNullable(err.Error(), true)

			savedTool, perr := s.chatRepository.CreateChatMessage(ctx, toolRecord)
			if perr != nil {
				return fmt.Errorf("failed to persist tool error message: %w", perr)
			}

			toolParam, perr := s.chatMessageToLLMParam(savedTool)
			if perr != nil {
				return perr
			}
			*messages = append(*messages, toolParam)

			if callbacks.OnToolEvent != nil {
				_ = callbacks.OnToolEvent(dto.ToolEvent{
					ToolName:   toolName,
					ToolCallID: toolCall.ID,
					ArgsJSON:   toolArgsJSON,
					State:      dto.ToolError,
					Error:      err.Error(),
				})
				_ = callbacks.OnToolEvent(dto.ToolEvent{
					ToolName:   toolName,
					ToolCallID: toolCall.ID,
					ArgsJSON:   toolArgsJSON,
					State:      dto.ToolCompleted,
				})
			}

			// Продолжаем к следующему tool call
			continue
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

		savedTool, err := s.chatRepository.CreateChatMessage(ctx, toolRecord)
		if err != nil {
			return fmt.Errorf("failed to persist tool message: %w", err)
		}

		toolParam, err := s.chatMessageToLLMParam(savedTool)
		if err != nil {
			return err
		}
		*messages = append(*messages, toolParam)

		if callbacks.OnToolEvent != nil {
			_ = callbacks.OnToolEvent(dto.ToolEvent{
				ToolName:   toolName,
				ToolCallID: toolCall.ID,
				ArgsJSON:   toolArgsJSON,
				State:      dto.ToolCompleted,
			})
		}
	}

	return nil
}

func (s *Service) handleAssistantFailure(ctx context.Context, chatID domain.ID, originalErr error) (dto.ChatCompletionDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.handleAssistantFailure")
	defer span.Finish()

	errMsg := domain.NewChatMessage(
		chatID,
		domain.ChatMessageRoleAssistant,
		assistantErrorMessage,
		utils.Nullable[string]{},
		utils.Nullable[string]{},
		nil,
	)
	errMsg.Error = utils.NewNullable(originalErr.Error(), true)

	if _, err := s.chatRepository.CreateChatMessage(ctx, errMsg); err != nil {
		return dto.ChatCompletionDTO{}, fmt.Errorf("failed to persist assistant error message: %w", err)
	}

	return dto.ChatCompletionDTO{}, errors.Join(domain.ErrInternal, originalErr)
}
