package repository

import (
	"context"
	"errors"
	"fitness-trainer/internal/domain"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type promptEntity struct {
	ID        pgtype.UUID        `db:"id"`
	CreatedAt pgtype.Timestamptz `db:"created_at"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at"`

	UserID pgtype.UUID `db:"user_id"`

	PromptText   pgtype.Text `db:"prompt_text"`
	SettingsHash pgtype.Text `db:"settings_hash"`
}

func (p promptEntity) toDomain() domain.Prompt {
	return domain.Prompt{
		Model: domain.Model{
			ID:        domain.ID(p.ID.Bytes),
			CreatedAt: p.CreatedAt.Time,
			UpdatedAt: p.UpdatedAt.Time,
		},
		UserID:       domain.ID(p.UserID.Bytes),
		PromptText:   p.PromptText.String,
		SettingsHash: p.SettingsHash.String,
	}
}

func promptFromDomain(prompt domain.Prompt) promptEntity {
	return promptEntity{
		ID:        uuidToPgtype(prompt.ID),
		CreatedAt: timeToPgtype(prompt.CreatedAt),
		UpdatedAt: timeToPgtype(prompt.UpdatedAt),

		UserID:       uuidToPgtype(prompt.UserID),
		PromptText:   pgtype.Text{String: prompt.PromptText, Valid: true},
		SettingsHash: pgtype.Text{String: prompt.SettingsHash, Valid: true},
	}
}

func (r *PGXRepository) CreatePrompt(ctx context.Context, prompt domain.Prompt) (domain.Prompt, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreatePrompt")
	defer span.Finish()

	promptEntity := promptFromDomain(prompt)

	const query = `
		INSERT INTO prompts (id, user_id, prompt_text, settings_hash, created_at)
		SELECT 
			$1, $2, $3, $4, $5
		FROM llm_settings gs
		WHERE gs.user_id = $2
		  AND gs.hash = $4
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	ct, err := engine.Exec(ctx, query,
		promptEntity.ID,
		promptEntity.UserID,
		promptEntity.PromptText,
		promptEntity.SettingsHash,
		promptEntity.CreatedAt,
	)
	if err != nil {
		return domain.Prompt{}, err
	}
	if ct.RowsAffected() == 0 {
		return domain.Prompt{}, domain.ErrNotFound
	}

	return promptEntity.toDomain(), nil
}

func (r *PGXRepository) GetLastPromptByUserID(ctx context.Context, userID domain.ID) (domain.Prompt, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetLastPromptByUserID")
	defer span.Finish()

	const query = `
		SELECT id, created_at, user_id, prompt_text, settings_hash
		FROM prompts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var prompt promptEntity
	if err := pgxscan.Get(ctx, engine, &prompt, query, uuidToPgtype(userID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Prompt{}, domain.ErrNotFound
		}
		return domain.Prompt{}, err
	}

	return prompt.toDomain(), nil
}
