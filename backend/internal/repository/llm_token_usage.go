package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fitness-trainer/internal/domain"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
)

// ReserveLLMTokens tries to reserve n tokens for (user, day) under a daily limit.
// Returns true if reservation was applied. No transaction is started here; caller controls it.
func (r *PGXRepository) ReserveLLMTokens(ctx context.Context, userID domain.ID, day time.Time, n int, dailyLimit int) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ReserveLLMTokens")
	defer span.Finish()

	engine := r.contextManager.GetEngineFromContext(ctx)

	// Single statement: attempt insert when within limit, or update existing row when within limit.
	// If limit is exceeded in both paths, no rows are affected.
	query := `
		INSERT INTO llm_token_usage (user_id, day, used_tokens, reserved_tokens)
		SELECT $1, $2, 0, $3::integer
		WHERE $3 <= $4
		ON CONFLICT (user_id, day) DO UPDATE
		SET reserved_tokens = llm_token_usage.reserved_tokens + EXCLUDED.reserved_tokens,
			updated_at = NOW()
		WHERE llm_token_usage.used_tokens + llm_token_usage.reserved_tokens + EXCLUDED.reserved_tokens <= $4
	`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(userID), day, n, dailyLimit)
	if err != nil {
		return false, fmt.Errorf("failed to reserve tokens: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

// ConfirmLLMTokenUsage applies actual usage and releases any unused reservation.
// This function is safe to call without prior reservation; it ensures the row exists.
func (r *PGXRepository) ConfirmLLMTokenUsage(ctx context.Context, userID domain.ID, day time.Time, reserved int, actual int) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ConfirmLLMTokenUsage")
	defer span.Finish()

	engine := r.contextManager.GetEngineFromContext(ctx)

	query := `
		INSERT INTO llm_token_usage (user_id, day, used_tokens, reserved_tokens)
		VALUES ($1, $2, $3, GREATEST(0 - $4, 0))
		ON CONFLICT (user_id, day) DO UPDATE
		SET used_tokens = llm_token_usage.used_tokens + EXCLUDED.used_tokens,
			reserved_tokens = GREATEST(llm_token_usage.reserved_tokens - $4, 0),
			updated_at = NOW()
	`

	// Single upsert: ensures row exists and applies usage atomically
	if _, err := engine.Exec(ctx, query, uuidToPgtype(userID), day, actual, reserved); err != nil {
		return fmt.Errorf("failed to confirm token usage: %w", err)
	}

	return nil
}

// GetLLMDailyUsage returns current used and reserved tokens for the day. If no row, returns zeros.
func (r *PGXRepository) GetLLMDailyUsage(ctx context.Context, userID domain.ID, day time.Time) (used int, reserved int, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetLLMDailyUsage")
	defer span.Finish()

	engine := r.contextManager.GetEngineFromContext(ctx)

	query := `
        SELECT used_tokens AS used, reserved_tokens AS reserved
        FROM llm_token_usage
        WHERE user_id = $1 AND day = $2
    `

	var row struct {
		Used     int `json:"used_tokens"`
		Reserved int `json:"reserved_tokens"`
	}

	if err = pgxscan.Get(ctx, engine, &row, query, uuidToPgtype(userID), day); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, nil
		}

		return 0, 0, fmt.Errorf("failed to get daily usage: %w", err)
	}

	return row.Used, row.Reserved, nil
}
