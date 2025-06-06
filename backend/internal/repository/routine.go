package repository

import (
	"context"
	"errors"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type routineEntity struct {
	ID            pgtype.UUID
	Name          string
	Description   string
	UserID        pgtype.UUID
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	ExerciseCount int `db:"exercise_count"`
}

func (r routineEntity) toDomain() domain.Routine {
	return domain.Routine{
		Model: domain.Model{
			ID:        domain.ID(r.ID.Bytes),
			CreatedAt: r.CreatedAt.Time,
			UpdatedAt: r.UpdatedAt.Time,
		},
		Name:          r.Name,
		Description:   r.Description,
		UserID:        domain.ID(r.UserID.Bytes),
		ExerciseCount: r.ExerciseCount,
	}
}

func routineFromDomain(routine domain.Routine) routineEntity {
	return routineEntity{
		ID:          uuidToPgtype(routine.ID),
		Name:        routine.Name,
		Description: routine.Description,
		UserID:      uuidToPgtype(routine.UserID),
		CreatedAt:   timeToPgtype(routine.CreatedAt),
		UpdatedAt:   timeToPgtype(routine.UpdatedAt),
	}
}

func (r *PGXRepository) GetRoutines(ctx context.Context, userID domain.ID) ([]domain.Routine, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetRoutines")
	defer span.Finish()

	query := `
		SELECT 
			r.id,
			r.name,
			r.description,
			r.user_id,
			r.created_at,
			r.updated_at,
			COUNT(ei.id) AS exercise_count
		FROM routines r
		LEFT JOIN exercise_instances ei ON ei.routine_id = r.id
		WHERE r.user_id = $1
		GROUP BY r.id, r.name, r.description, r.user_id, r.created_at, r.updated_at
		ORDER BY r.created_at DESC
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var routines []routineEntity
	err := pgxscan.Select(ctx, engine, &routines, query, uuidToPgtype(userID))
	if err != nil {
		logger.Errorf("failed to get routines: %v", err)
		return nil, err
	}

	result := make([]domain.Routine, len(routines))
	for i, routine := range routines {
		result[i] = routine.toDomain()
	}

	return result, nil
}

func (r *PGXRepository) CreateRoutine(ctx context.Context, routine domain.Routine) (domain.Routine, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateRoutine")
	defer span.Finish()

	query := `
		INSERT INTO routines (id, name, description, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING 
			id, name, description, user_id, created_at, updated_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	entity := routineFromDomain(routine)
	err := pgxscan.Get(ctx, engine, &entity, query, entity.ID, entity.Name, entity.Description, entity.UserID)
	if err != nil {
		logger.Errorf("failed to create routine: %v", err)
		return domain.Routine{}, err
	}

	return entity.toDomain(), nil
}

func (r *PGXRepository) GetRoutineByID(ctx context.Context, id domain.ID) (domain.Routine, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetRoutineByID")
	defer span.Finish()

	query := `
		SELECT
			r.id,
			r.name,
			r.description,
			r.user_id,
			r.created_at,
			r.updated_at,
			COUNT(ei.id) AS exercise_count
		FROM routines r
		LEFT JOIN exercise_instances ei ON ei.routine_id = r.id
		WHERE r.id = $1
		GROUP BY r.id, r.name, r.description, r.user_id, r.created_at, r.updated_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var routine routineEntity
	err := pgxscan.Get(ctx, engine, &routine, query, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Routine{}, domain.ErrNotFound
		}
		logger.Errorf("failed to get routine by id: %v", err)
		return domain.Routine{}, err
	}

	return routine.toDomain(), nil
}

func (r *PGXRepository) DeleteRoutine(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteRoutine")
	defer span.Finish()

	query := `
		DELETE FROM routines r
		WHERE r.id = $1
		RETURNING id
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var routine routineEntity
	err := pgxscan.Get(ctx, engine, &routine, query, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		logger.Errorf("failed to delete routine: %v", err)
		return err
	}

	return nil
}

func (r *PGXRepository) UpdateRoutine(ctx context.Context, id domain.ID, routine domain.Routine) (domain.Routine, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateRoutine")
	defer span.Finish()

	query := `
		UPDATE routines
		SET name = $2, description = $3
		WHERE id = $1
		RETURNING id, name, description, user_id, created_at, updated_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	entity := routineFromDomain(routine)
	err := pgxscan.Get(ctx, engine, &entity, query, entity.ID, entity.Name, entity.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Routine{}, domain.ErrNotFound
		}
		logger.Errorf("failed to update routine: %v", err)
		return domain.Routine{}, err
	}

	return entity.toDomain(), nil
}
