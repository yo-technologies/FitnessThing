package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	openai_client "fitness-trainer/internal/clients/openai"
	"fitness-trainer/internal/logger"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/utils"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/opentracing/opentracing-go"
)

const (
	chatHistoryLimit       = 200
	maxChatCompletionLoops = 25
	assistantErrorMessage  = "Не удалось ответить. Попробуйте ещё раз чуть позже."
)

const defaultChatSystemPrompt = `Ты — виртуальный фитнес‑тренер и агент управления тренировками.

ОСНОВНАЯ ЛОГИКА: если для осмысленного ответа нужны структурированные или актуальные данные (упражнения, план, история) — используй инструменты. Если вопрос бытовой, уточняющий или мотивационный — можно ответить сразу.

Твои задачи:
- Поддерживать диалог, следуя языку и стилю общения пользователя.
- Выступать в роли персонального фитнес тренера.
- Помогать пользователю уточнить цели, ограничения, предпочтения.
- Использовать инструменты, когда есть риск придумать данные или когда нужно изменить/построить/проверить план.
- НЕ выдумывать упражнения, ID, результаты. Не давать медицинских советов.
- Перед изменением плана получать его актуальное состояние.
- После изменений объяснять что произошло и что дальше.
- Напоминать о разминке и восстановлении в итоговых планах.

Когда МОЖНО ответить БЕЗ инструмента:
- Приветствие, мотивация, небольшое пояснение принципа тренинга.
- Уточнение цели: можно задать 1–2 наводящих вопроса.
- Краткий комментарий к уже полученным данным (которые пришли из инструмента ранее в истории).
- Оценка, подходит ли уже существующий план (если он есть в истории) — можно сначала суммировать и спросить уточнение.
- Перефразирование/подтверждение запроса пользователя.

Когда ОБЯЗАТЕЛЬНО нужен инструмент:
- Нужны перечни (мышечные группы, упражнения, история, текущий план).
- Любое добавление / удаление / замена упражнений или сетов.
- Пересборка плана или предложение детального нового плана.

Если пользователь просит "сгенерируй тренировку", а у тебя нет данных:
1) list_muscle_groups 2) уточни предпочтения (или list_exercises для релевантных групп) 3) только потом собирай структуру.

Протокол (внутренний цикл):
1. THINK (молча): чего не хватает?
2. DECIDE: нужен ли инструмент? Если сомневаешься — сначала уточни пользователя, а не отказывай.
3. FETCH/APPLY: вызови нужные инструменты (можно несколько по очереди).
4. EXPLAIN: коротко что сделал / что получил.
5. NEXT: предложи логичный следующий шаг.

Если инструмент не нужен — переходи сразу к EXPLAIN/NEXT.

Формат ответа после инструмента: "Сделано: <что>. Далее предлагаю: <шаг/вопрос>." (кратко).

Если данные пустые или недостаточные — запроси уточнение вместо догадок.

Не извиняйся часто. Будь конструктивным, кратким и полезным.`

type chatSession struct {
	chat     domain.Chat
	messages []openai.ChatCompletionMessageParamUnion
	toolDefs []openai.ChatCompletionToolUnionParam
}

func (s *Service) SendChatMessageStream(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest, callbacks dto.ChatStreamCallbacks) (dto.ChatCompletionDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.SendChatMessageStream")
	defer span.Finish()

	session, err := s.startChatSession(ctx, userID, req)
	if err != nil {
		return dto.ChatCompletionDTO{}, err
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, len(session.messages))
	copy(messages, session.messages)
	toolDefs := session.toolDefs
	chat := session.chat
	agentCtx := domain.NewAgentChatContext(userID, chat.WorkoutID)

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
		params := s.newChatCompletionParams(messages, toolDefs, s.openAIModel, s.reasoningEffort, true)

		stream, err := s.openAIClient.CreateChatCompletionStream(ctx, params)
		if err != nil {
			_, _ = s.handleAssistantFailure(ctx, chat.ID, err)
			return dto.ChatCompletionDTO{}, err
		}

		assistantMessage, usage, err := s.runStreamingChatCompletion(stream, &callbacks)
		if err != nil {
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

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(systemMessages)+len(history)+4)
	messages = append(messages, systemMessages...)

	for _, msg := range history {
		param, err := s.chatMessageToOpenAIParam(msg)
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
		Role:    "assistant",
		Content: contentBuilder.String(),
	}

	if len(toolCalls) > 0 {
		indices := make([]int, 0, len(toolCalls))
		for idx := range toolCalls {
			indices = append(indices, int(idx))
		}
		sort.Ints(indices)

		finalCalls := make([]openai.ChatCompletionMessageToolCallUnion, 0, len(indices))
		for _, idx := range indices {
			acc := toolCalls[int64(idx)]
			finalCalls = append(finalCalls, openai.ChatCompletionMessageToolCallUnion{
				// ID is shared across variants
				ID: acc.id,
				// Function variant payload
				Function: openai.ChatCompletionMessageFunctionToolCallFunction{
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

func (s *Service) buildChatSystemMessages(ctx context.Context, userID domain.ID, chat domain.Chat) ([]openai.ChatCompletionMessageParamUnion, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.buildChatSystemMessages")
	defer span.Finish()

	messages := []openai.ChatCompletionMessageParamUnion{openai.SystemMessage(defaultChatSystemPrompt)}

	prompt, err := s.userPromptRepository.GetLastPromptByUserID(ctx, userID)
	if err == nil {
		messages = append(messages, openai.SystemMessage(fmt.Sprintf("\n\nЛичные пожелания пользователя: %s", prompt.PromptText)))
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("failed to load generation settings: %w", err)
	}

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
		// Build assistant message with optional tool call
		assistant := openai.ChatCompletionAssistantMessageParam{
			Content: openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(message.Content),
			},
		}
		if message.ToolCallID.IsValid && message.ToolName.IsValid {
			argsJSON := "{}"
			if len(message.ToolArguments) > 0 {
				raw, err := json.Marshal(message.ToolArguments)
				if err != nil {
					return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("failed to marshal tool arguments: %w", err)
				}
				argsJSON = string(raw)
			}
			assistant.ToolCalls = []openai.ChatCompletionMessageToolCallUnionParam{
				{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: message.ToolCallID.V,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      message.ToolName.V,
							Arguments: argsJSON,
						},
					},
				},
			}
		}
		return openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant}, nil
	case domain.ChatMessageRoleTool:
		if !message.ToolCallID.IsValid {
			return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("tool message missing tool call id")
		}
		content := message.Content
		if strings.TrimSpace(content) == "" && message.Error.IsValid {
			content = fmt.Sprintf("error: %s", message.Error.V)
		}
		return openai.ToolMessage(content, message.ToolCallID.V), nil
	case domain.ChatMessageRoleSystem:
		return openai.SystemMessage(message.Content), nil
	default:
		return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("unsupported chat message role: %s", message.Role)
	}
}

func (s *Service) handleAssistantToolCalls(
	ctx context.Context,
	chatID domain.ID,
	messages *[]openai.ChatCompletionMessageParamUnion,
	chatCtx domain.AgentChatContext,
	assistantMessage openai.ChatCompletionMessage,
	callbacks *dto.ChatStreamCallbacks,
) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.handleAssistantToolCalls")
	defer span.Finish()

	// Сохраняем отдельный assistant message для каждого tool call
	// Это упрощает восстановление истории и UI логику
	contentForFirst := assistantMessage.Content

	for idx, toolCall := range assistantMessage.ToolCalls {
		toolName := toolCall.Function.Name
		toolArgsJSON := toolCall.Function.Arguments

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

		param, err := s.chatMessageToOpenAIParam(savedAssistant)
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

			toolParam, perr := s.chatMessageToOpenAIParam(savedTool)
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

		toolParam, err := s.chatMessageToOpenAIParam(savedTool)
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

func (s *Service) newChatCompletionParams(messages []openai.ChatCompletionMessageParamUnion, tools []openai.ChatCompletionToolUnionParam, model string, reasoningEffort string, stream bool) openai.ChatCompletionNewParams {
	params := openai.ChatCompletionNewParams{
		Model:           shared.ChatModel(model),
		Messages:        messages,
		ReasoningEffort: shared.ReasoningEffort(reasoningEffort),
	}

	if len(tools) > 0 {
		params.Tools = tools
		params.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{OfAuto: openai.String("auto")}
	}

	if stream {
		params.StreamOptions = openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)}
	}

	return params
}
