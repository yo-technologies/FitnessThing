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

type userFactEntity struct {
	ID        pgtype.UUID        `db:"id"`
	UserID    pgtype.UUID        `db:"user_id"`
	Content   pgtype.Text        `db:"content"`
	CreatedAt pgtype.Timestamptz `db:"created_at"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at"`
}

func (f userFactEntity) toDomain() domain.UserFact {
	return domain.UserFact{
		Model: domain.Model{
			ID:        domain.ID(f.ID.Bytes),
			CreatedAt: timeFromPgtype(f.CreatedAt),
			UpdatedAt: timeFromPgtype(f.UpdatedAt),
		},
		UserID:  domain.ID(f.UserID.Bytes),
		Content: f.Content.String,
	}
}

func userFactFromDomain(f domain.UserFact) userFactEntity {
	return userFactEntity{
		ID:        uuidToPgtype(f.ID),
		UserID:    uuidToPgtype(f.UserID),
		Content:   pgtype.Text{String: f.Content, Valid: true},
		CreatedAt: timeToPgtype(f.CreatedAt),
		UpdatedAt: timeToPgtype(f.UpdatedAt),
	}
}

func (r *PGXRepository) CreateUserFact(ctx context.Context, fact domain.UserFact) (domain.UserFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateUserFact")
	defer span.Finish()

	factEntity := userFactFromDomain(fact)

	const query = `
		INSERT INTO user_facts (id, user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, content, created_at, updated_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	if err := pgxscan.Get(ctx, engine, &factEntity, query,
		factEntity.ID,
		factEntity.UserID,
		factEntity.Content,
		factEntity.CreatedAt,
		factEntity.UpdatedAt,
	); err != nil {
		return domain.UserFact{}, err
	}

	return factEntity.toDomain(), nil
}

func (r *PGXRepository) ListUserFacts(ctx context.Context, userID domain.ID, limit int) ([]domain.UserFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListUserFacts")
	defer span.Finish()

	if limit <= 0 || limit > domain.MaxUserFactsPerUser {
		limit = domain.MaxUserFactsPerUser
	}

	const query = `
		SELECT id, user_id, content, created_at, updated_at
		FROM user_facts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	engine := r.contextManager.GetEngineFromContext(ctx)
	facts := make([]userFactEntity, 0)
	if err := pgxscan.Select(ctx, engine, &facts, query, uuidToPgtype(userID), limit); err != nil {
		return nil, err
	}

	result := make([]domain.UserFact, 0, len(facts))
	for _, fact := range facts {
		result = append(result, fact.toDomain())
	}

	return result, nil
}

func (r *PGXRepository) DeleteUserFact(ctx context.Context, userID, factID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteUserFact")
	defer span.Finish()

	const query = `
		DELETE FROM user_facts
		WHERE id = $1 AND user_id = $2
	`

	engine := r.contextManager.GetEngineFromContext(ctx)
	ct, err := engine.Exec(ctx, query, uuidToPgtype(factID), uuidToPgtype(userID))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *PGXRepository) CountUserFactsByUserID(ctx context.Context, userID domain.ID) (int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CountUserFactsByUserID")
	defer span.Finish()

	const query = `
		SELECT COUNT(*)
		FROM user_facts
		WHERE user_id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)
	var count int
	if err := pgxscan.Get(ctx, engine, &count, query, uuidToPgtype(userID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}