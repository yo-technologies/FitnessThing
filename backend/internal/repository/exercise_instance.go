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

type exerciseInstanceEntity struct {
	ID         pgtype.UUID
	RoutineID  pgtype.UUID
	ExerciseID pgtype.UUID
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
}

func (e exerciseInstanceEntity) toDomain() domain.ExerciseInstance {
	return domain.ExerciseInstance{
		Model: domain.Model{
			ID:        domain.ID(e.ID.Bytes),
			CreatedAt: e.CreatedAt.Time,
			UpdatedAt: e.UpdatedAt.Time,
		},
		RoutineID:  domain.ID(e.RoutineID.Bytes),
		ExerciseID: domain.ID(e.ExerciseID.Bytes),
	}
}

func exerciseInstanceFromDomain(exerciseInstance domain.ExerciseInstance) exerciseInstanceEntity {
	return exerciseInstanceEntity{
		ID:         uuidToPgtype(exerciseInstance.ID),
		RoutineID:  uuidToPgtype(exerciseInstance.RoutineID),
		ExerciseID: uuidToPgtype(exerciseInstance.ExerciseID),
		CreatedAt:  timeToPgtype(exerciseInstance.CreatedAt),
		UpdatedAt:  timeToPgtype(exerciseInstance.UpdatedAt),
	}
}

func (r *PGXRepository) GetExerciseInstanceByID(ctx context.Context, id domain.ID) (domain.ExerciseInstance, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetExerciseInstanceByID")
	defer span.Finish()

	query := `
		SELECT ei.id, ei.routine_id, ei.exercise_id, ei.created_at, ei.updated_at
		FROM exercise_instances ei
		WHERE ei.id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var exerciseInstance exerciseInstanceEntity
	err := pgxscan.Get(ctx, engine, &exerciseInstance, query, uuidToPgtype(id))
	if err != nil {
		logger.Errorf("failed to get exercise instance by id: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ExerciseInstance{}, domain.ErrNotFound
		}
		return domain.ExerciseInstance{}, err
	}

	return exerciseInstance.toDomain(), nil
}

func (r *PGXRepository) GetExerciseInstancesByRoutineID(ctx context.Context, routineID domain.ID) ([]domain.ExerciseInstance, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetExerciseInstancesByRoutineID")
	defer span.Finish()

	query := `
		SELECT ei.id, ei.routine_id, ei.exercise_id, ei.created_at, ei.updated_at
		FROM exercise_instances ei
		WHERE ei.routine_id = $1
		ORDER BY ei.position, ei.created_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var exerciseInstances []exerciseInstanceEntity
	err := pgxscan.Select(ctx, engine, &exerciseInstances, query, uuidToPgtype(routineID))
	if err != nil {
		logger.Errorf("failed to get exercise instances by routine id: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	result := make([]domain.ExerciseInstance, len(exerciseInstances))
	for i, exerciseInstance := range exerciseInstances {
		result[i] = exerciseInstance.toDomain()
	}

	return result, nil
}

func (r *PGXRepository) CreateExerciseInstance(ctx context.Context, exerciseInstance domain.ExerciseInstance) (domain.ExerciseInstance, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateExerciseInstance")
	defer span.Finish()

	query := `
		INSERT INTO exercise_instances (id, routine_id, exercise_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, routine_id, exercise_id, created_at, updated_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	entity := exerciseInstanceFromDomain(exerciseInstance)
	err := pgxscan.Get(ctx, engine, &entity, query, entity.ID, entity.RoutineID, entity.ExerciseID, entity.CreatedAt, entity.UpdatedAt)
	if err != nil {
		logger.Errorf("failed to create exercise instance: %v", err)
		return domain.ExerciseInstance{}, err
	}

	return entity.toDomain(), nil
}

func (r *PGXRepository) DeleteExerciseInstance(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteExerciseInstance")
	defer span.Finish()

	query := `
		DELETE FROM exercise_instances
		WHERE id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	_, err := engine.Exec(ctx, query, uuidToPgtype(id))
	if err != nil {
		logger.Errorf("failed to delete exercise instance: %v", err)
		return err
	}

	return nil
}

func (r *PGXRepository) SetExerciseOrder(ctx context.Context, routineID domain.ID, exerciseInstanceIDs []domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.SetExerciseOrder")
	defer span.Finish()

	query := `
		UPDATE exercise_instances ei
		SET position = idx
		FROM (SELECT unnest($1::UUID[]) AS id, generate_series(0, array_length($1, 1) - 1) AS idx) AS new_order
		WHERE ei.id = new_order.id
		AND ei.routine_id = $2
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	_, err := engine.Exec(ctx, query, uuidsToPgtype(exerciseInstanceIDs), uuidToPgtype(routineID))
	if err != nil {
		logger.Errorf("failed to set exercise order: %v", err)
		return err
	}

	return nil
}
