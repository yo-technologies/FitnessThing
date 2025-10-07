package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type chatEntity struct {
	ID        pgtype.UUID
	UserID    pgtype.UUID
	WorkoutID pgtype.UUID
	Title     pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (c chatEntity) toDomain() domain.Chat {
	return domain.Chat{
		Model: domain.Model{
			ID:        domain.ID(c.ID.Bytes),
			CreatedAt: c.CreatedAt.Time,
			UpdatedAt: c.UpdatedAt.Time,
		},
		UserID:    domain.ID(c.UserID.Bytes),
		WorkoutID: utils.NewNullable(domain.ID(c.WorkoutID.Bytes), c.WorkoutID.Valid),
		Title:     c.Title.String,
	}
}

func chatFromDomain(chat domain.Chat) chatEntity {
	return chatEntity{
		ID:        uuidToPgtype(chat.ID),
		UserID:    uuidToPgtype(chat.UserID),
		WorkoutID: nullableIDToPgtype(chat.WorkoutID),
		Title: pgtype.Text{
			String: chat.Title,
			Valid:  chat.Title != "",
		},
	}
}

type chatMessageEntity struct {
	ID            pgtype.UUID
	ChatID        pgtype.UUID
	Role          pgtype.Text
	Content       pgtype.Text
	ToolName      pgtype.Text
	ToolCallID    pgtype.Text
	ToolArguments pgtype.Text
	TokenUsage    pgtype.Int4
	Error         pgtype.Text
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
}

func (m chatMessageEntity) toDomain() (domain.ChatMessage, error) {
	var toolArgs map[string]any
	if m.ToolArguments.Valid {
		if err := json.Unmarshal([]byte(m.ToolArguments.String), &toolArgs); err != nil {
			return domain.ChatMessage{}, err
		}
	}

	var tokenUsage utils.Nullable[int]
	if m.TokenUsage.Valid {
		tokenUsage = utils.NewNullable(int(m.TokenUsage.Int32), true)
	}

	return domain.ChatMessage{
		Model: domain.Model{
			ID:        domain.ID(m.ID.Bytes),
			CreatedAt: m.CreatedAt.Time,
			UpdatedAt: m.UpdatedAt.Time,
		},
		ChatID:        domain.ID(m.ChatID.Bytes),
		Role:          domain.ChatMessageRole(m.Role.String),
		Content:       m.Content.String,
		ToolName:      utils.NewNullable(m.ToolName.String, m.ToolName.Valid),
		ToolCallID:    utils.NewNullable(m.ToolCallID.String, m.ToolCallID.Valid),
		ToolArguments: toolArgs,
		TokenUsage:    tokenUsage,
		Error:         utils.NewNullable(m.Error.String, m.Error.Valid),
	}, nil
}

func chatMessageFromDomain(message domain.ChatMessage) (chatMessageEntity, error) {
	entity := chatMessageEntity{
		ID:      uuidToPgtype(message.ID),
		ChatID:  uuidToPgtype(message.ChatID),
		Role:    pgtype.Text{String: message.Role.String(), Valid: true},
		Content: pgtype.Text{String: message.Content, Valid: true},
		ToolName: pgtype.Text{
			String: message.ToolName.V,
			Valid:  message.ToolName.IsValid,
		},
		ToolCallID: pgtype.Text{
			String: message.ToolCallID.V,
			Valid:  message.ToolCallID.IsValid,
		},
		TokenUsage: pgtype.Int4{Int32: int32(message.TokenUsage.V), Valid: message.TokenUsage.IsValid},
		Error: pgtype.Text{
			String: message.Error.V,
			Valid:  message.Error.IsValid,
		},
	}

	if message.ToolArguments != nil {
		raw, err := json.Marshal(message.ToolArguments)
		if err != nil {
			return chatMessageEntity{}, err
		}
		entity.ToolArguments = pgtype.Text{String: string(raw), Valid: true}
	}

	return entity, nil
}

func (r *PGXRepository) CreateChat(ctx context.Context, chat domain.Chat) (domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateChat")
	defer span.Finish()

	query := `
		INSERT INTO llm_chats (id, user_id, workout_id, title)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, workout_id, title, created_at, updated_at
	`

	entity := chatFromDomain(chat)

	engine := r.contextManager.GetEngineFromContext(ctx)

	if err := pgxscan.Get(ctx, engine, &entity, query,
		entity.ID,
		entity.UserID,
		entity.WorkoutID,
		entity.Title,
	); err != nil {
		logger.Errorf("failed to create chat: %v", err)
		return domain.Chat{}, err
	}

	return entity.toDomain(), nil
}

func (r *PGXRepository) GetChatByWorkoutID(ctx context.Context, workoutID domain.ID) (domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetChatByWorkoutID")
	defer span.Finish()

	query := `
		SELECT id, user_id, workout_id, title, created_at, updated_at
		FROM llm_chats
		WHERE workout_id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var entity chatEntity
	if err := pgxscan.Get(ctx, engine, &entity, query, uuidToPgtype(workoutID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Chat{}, domain.ErrNotFound
		}
		logger.Errorf("failed to get chat by workout id: %v", err)
		return domain.Chat{}, domain.ErrInternal
	}

	return entity.toDomain(), nil
}

func (r *PGXRepository) GetChatByID(ctx context.Context, chatID domain.ID) (domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetChatByID")
	defer span.Finish()

	query := `
		SELECT id, user_id, workout_id, title, created_at, updated_at
		FROM llm_chats
		WHERE id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var entity chatEntity
	if err := pgxscan.Get(ctx, engine, &entity, query, uuidToPgtype(chatID)); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Chat{}, domain.ErrNotFound
		}
		logger.Errorf("failed to get chat by id: %v", err)
		return domain.Chat{}, domain.ErrInternal
	}

	return entity.toDomain(), nil
}

func (r *PGXRepository) CreateChatMessage(ctx context.Context, message domain.ChatMessage) (domain.ChatMessage, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateChatMessage")
	defer span.Finish()

	entity, err := chatMessageFromDomain(message)
	if err != nil {
		return domain.ChatMessage{}, err
	}

	query := `
		INSERT INTO llm_chat_messages (id, chat_id, role, content, tool_name, tool_call_id, tool_arguments, token_usage, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, chat_id, role, content, tool_name, tool_call_id, tool_arguments, token_usage, error, created_at, updated_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	if err := pgxscan.Get(ctx, engine, &entity, query,
		entity.ID,
		entity.ChatID,
		entity.Role,
		entity.Content,
		entity.ToolName,
		entity.ToolCallID,
		entity.ToolArguments,
		entity.TokenUsage,
		entity.Error,
	); err != nil {
		logger.Errorf("failed to create chat message: %v", err)
		return domain.ChatMessage{}, err
	}

	return entity.toDomain()
}

func (r *PGXRepository) ListChatMessages(ctx context.Context, chatID domain.ID, limit, offset int) ([]domain.ChatMessage, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListChatMessages")
	defer span.Finish()

	query := `
		SELECT id, chat_id, role, content, tool_name, tool_call_id, tool_arguments, token_usage, error, created_at, updated_at
		FROM llm_chat_messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var entities []chatMessageEntity
	if err := pgxscan.Select(ctx, engine, &entities, query, uuidToPgtype(chatID), limit, offset); err != nil {
		logger.Errorf("failed to list chat messages: %v", err)
		return nil, domain.ErrInternal
	}

	messages := make([]domain.ChatMessage, 0, len(entities))
	for _, entity := range entities {
		msg, err := entity.toDomain()
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
