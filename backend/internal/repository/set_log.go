package repository

import (
	"context"
	"fitness-trainer/internal/domain"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type setLogEntity struct {
	ID            pgtype.UUID        `db:"id"`
	ExerciseLogID pgtype.UUID        `db:"exercise_log_id"`
	Reps          int                `db:"reps"`
	Weight        float32            `db:"weight"`
	Time          pgtype.Interval    `db:"time"`
	UpdatedAt     pgtype.Timestamptz `db:"updated_at"`
	CreatedAt     pgtype.Timestamptz `db:"created_at"`
}

func (s setLogEntity) toDomain() domain.ExerciseSetLog {
	return domain.ExerciseSetLog{
		Model: domain.Model{
			ID:        domain.ID(s.ID.Bytes),
			CreatedAt: s.CreatedAt.Time,
			UpdatedAt: s.UpdatedAt.Time,
		},
		ExerciseLogID: domain.ID(s.ExerciseLogID.Bytes),
		Reps:          s.Reps,
		Weight:        s.Weight,
		Time:          durationFromPgtype(s.Time),
	}
}

func setLogFromDomain(setLog domain.ExerciseSetLog) setLogEntity {
	return setLogEntity{
		ID:            uuidToPgtype(setLog.ID),
		ExerciseLogID: uuidToPgtype(setLog.ExerciseLogID),
		Reps:          setLog.Reps,
		Weight:        setLog.Weight,
		Time:          intervalToPgtype(setLog.Time),
		CreatedAt:     timeToPgtype(setLog.CreatedAt),
		UpdatedAt:     timeToPgtype(setLog.UpdatedAt),
	}
}

func toSetLogsDomain(setLogs []setLogEntity) []domain.ExerciseSetLog {
	result := make([]domain.ExerciseSetLog, 0, len(setLogs))
	for _, setLog := range setLogs {
		result = append(result, setLog.toDomain())
	}

	return result
}

func (r *PGXRepository) GetSetLogsByExerciseLogID(ctx context.Context, exerciseLogID domain.ID) ([]domain.ExerciseSetLog, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetSetLogsByExerciseLogID")
	defer span.Finish()

	query := `
		SELECT id, created_at, exercise_log_id, reps, weight, time, updated_at
		FROM set_logs
		WHERE exercise_log_id = $1
		ORDER BY created_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var setLogs []setLogEntity
	if err := pgxscan.Select(ctx, engine, &setLogs, query, exerciseLogID); err != nil {
		return nil, err
	}

	return toSetLogsDomain(setLogs), nil
}

func (r *PGXRepository) CreateSetLog(ctx context.Context, setLog domain.ExerciseSetLog) (domain.ExerciseSetLog, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateSetLog")
	defer span.Finish()

	query := `
		INSERT INTO set_logs (id, created_at, exercise_log_id, reps, weight, time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	setLogEntity := setLogFromDomain(setLog)

	if err := pgxscan.Get(
		ctx,
		engine,
		&setLogEntity,
		query,
		setLogEntity.ID,
		setLogEntity.CreatedAt,
		setLogEntity.ExerciseLogID,
		setLogEntity.Reps,
		setLogEntity.Weight,
		setLogEntity.Time,
	); err != nil {
		return domain.ExerciseSetLog{}, err
	}

	return setLogEntity.toDomain(), nil
}

func (r *PGXRepository) GetSetLogByID(ctx context.Context, id domain.ID) (domain.ExerciseSetLog, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetSetLogByID")
	defer span.Finish()

	query := `
		SELECT id, created_at, exercise_log_id, reps, weight, time, updated_at
		FROM set_logs
		WHERE id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var setLog setLogEntity
	if err := pgxscan.Get(ctx, engine, &setLog, query, id); err != nil {
		return domain.ExerciseSetLog{}, err
	}

	return setLog.toDomain(), nil
}

func (r *PGXRepository) DeleteSetLog(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteSetLog")
	defer span.Finish()

	query := `
		DELETE FROM set_logs
		WHERE id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	_, err := engine.Exec(ctx, query, uuidToPgtype(id))
	if err != nil {
		return err
	}

	return nil
}

func (r *PGXRepository) UpdateSetLog(ctx context.Context, id domain.ID, setLog domain.ExerciseSetLog) (domain.ExerciseSetLog, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateSetLog")
	defer span.Finish()

	query := `
		UPDATE set_logs
		SET reps = $2, weight = $3, time = $4, updated_at = $5
		WHERE id = $1
		RETURNING *
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	setLogEntity := setLogFromDomain(setLog)
	if err := pgxscan.Get(
		ctx,
		engine,
		&setLogEntity,
		query,
		setLogEntity.ID,
		setLogEntity.Reps,
		setLogEntity.Weight,
		setLogEntity.Time,
		timeToPgtype(setLog.UpdatedAt),
	); err != nil {
		return domain.ExerciseSetLog{}, err
	}

	return setLogEntity.toDomain(), nil
}
