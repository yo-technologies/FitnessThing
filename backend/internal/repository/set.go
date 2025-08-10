package repository

import (
	"context"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type setEntity struct {
	ID                 pgtype.UUID        `db:"id"`
	ExerciseInstanceID pgtype.UUID        `db:"exercise_instance_id"`
	SetType            pgtype.Text        `db:"set_type"`
	Reps               pgtype.Int8        `db:"reps"`
	Weight             pgtype.Float4      `db:"weight"`
	Time               pgtype.Interval    `db:"time"`
	UpdatedAt          pgtype.Timestamptz `db:"updated_at"`
	CreatedAt          pgtype.Timestamptz `db:"created_at"`
}

func setTypeToDomain(setType string) domain.SetType {
	switch setType {
	case "reps":
		return domain.SetTypeReps
	case "time":
		return domain.SetTypeTime
	case "weight":
		return domain.SetTypeWeight
	default:
		return domain.SetTypeUnknown
	}
}

func (s setEntity) toDomain() domain.Set {
	return domain.Set{
		Model: domain.Model{
			ID:        domain.ID(s.ID.Bytes),
			CreatedAt: s.CreatedAt.Time,
			UpdatedAt: s.UpdatedAt.Time,
		},
		ExerciseInstanceID: domain.ID(s.ExerciseInstanceID.Bytes),
		SetType:            setTypeToDomain(s.SetType.String),
		Reps:               int(s.Reps.Int64),
		Weight:             s.Weight.Float32,
		Time:               durationFromPgtype(s.Time),
	}
}

func setFromDomain(set domain.Set) setEntity {
	return setEntity{
		ID:                 uuidToPgtype(set.ID),
		ExerciseInstanceID: uuidToPgtype(set.ExerciseInstanceID),
		SetType:            pgtype.Text{String: set.SetType.String(), Valid: set.SetType != domain.SetTypeUnknown},
		Reps:               pgtype.Int8{Int64: int64(set.Reps), Valid: set.Reps != 0},
		Weight:             pgtype.Float4{Float32: set.Weight, Valid: set.Weight != 0},
		Time:               intervalToPgtype(set.Time),
		CreatedAt:          timeToPgtype(set.CreatedAt),
		UpdatedAt:          timeToPgtype(set.UpdatedAt),
	}
}

func toSetsDomain(sets []setEntity) []domain.Set {
	result := make([]domain.Set, 0, len(sets))
	for _, set := range sets {
		result = append(result, set.toDomain())
	}

	return result
}

func (r *PGXRepository) GetSetsByExerciseInstanceID(ctx context.Context, exerciseInstanceID domain.ID) ([]domain.Set, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetSetsByExerciseInstanceID")
	defer span.Finish()

	query := `
		SELECT id, exercise_instance_id, reps, weight, time, set_type, updated_at, created_at
		FROM sets
		WHERE exercise_instance_id = $1
		ORDER BY created_at ASC
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var sets []setEntity
	if err := pgxscan.Select(ctx, engine, &sets, query, exerciseInstanceID); err != nil {
		return nil, err
	}

	return toSetsDomain(sets), nil
}

func (r *PGXRepository) CreateSet(ctx context.Context, set domain.Set) (domain.Set, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateSet")
	defer span.Finish()

	query := `
		INSERT INTO sets (id, exercise_instance_id, reps, weight, time, set_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	entity := setFromDomain(set)
	if err := pgxscan.Get(ctx, engine, &entity.CreatedAt, query, entity.ID, entity.ExerciseInstanceID, entity.Reps, entity.Weight, entity.Time, entity.SetType); err != nil {
		logger.Errorf("failed to create set: %v", err)
		return domain.Set{}, err
	}

	return entity.toDomain(), nil
}

func (r *PGXRepository) GetSetByID(ctx context.Context, id domain.ID) (domain.Set, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetSetByID")
	defer span.Finish()

	query := `
		SELECT id, exercise_instance_id, reps, weight, time, set_type, updated_at, created_at
		FROM sets
		WHERE id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var set setEntity
	if err := pgxscan.Get(ctx, engine, &set, query, id); err != nil {
		return domain.Set{}, err
	}

	return set.toDomain(), nil
}

func (r *PGXRepository) DeleteSet(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteSet")
	defer span.Finish()

	query := `
		DELETE FROM sets
		WHERE id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	if _, err := engine.Exec(ctx, query, id); err != nil {
		return err
	}

	return nil
}

func (r *PGXRepository) UpdateSet(ctx context.Context, id domain.ID, set domain.Set) (domain.Set, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateSet")
	defer span.Finish()

	query := `
		UPDATE sets
		SET reps = $1, weight = $2, time = $3, set_type = $4
		WHERE id = $5
		RETURNING updated_at
	`

	entity := setFromDomain(set)

	engine := r.contextManager.GetEngineFromContext(ctx)

	if err := pgxscan.Get(ctx, engine, &entity.UpdatedAt, query, entity.Reps, entity.Weight, entity.Time, entity.SetType, entity.ID); err != nil {
		return domain.Set{}, err
	}

	return entity.toDomain(), nil
}
