package repository

import (
	"context"
	"fitness-trainer/internal/domain"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type expectedSetEntity struct {
	ID            pgtype.UUID        `db:"id"`
	ExerciseLogID pgtype.UUID        `db:"exercise_log_id"`
	SetType       pgtype.Text        `db:"set_type"`
	Reps          pgtype.Int8        `db:"reps"`
	Weight        pgtype.Float4      `db:"weight"`
	Time          pgtype.Interval    `db:"time"`
	CreatedAt     pgtype.Timestamptz `db:"created_at"`
	UpdatedAt     pgtype.Timestamptz `db:"updated_at"`
}

func (s expectedSetEntity) toDomain() domain.ExpectedSet {
	return domain.ExpectedSet{
		Model: domain.Model{
			ID:        domain.ID(s.ID.Bytes),
			CreatedAt: s.CreatedAt.Time,
			UpdatedAt: s.UpdatedAt.Time,
		},
		ExerciseLogID: domain.ID(s.ExerciseLogID.Bytes),
		SetType:       setTypeToDomain(s.SetType.String),
		Reps:          int(s.Reps.Int64),
		Weight:        s.Weight.Float32,
		Time:          durationFromPgtype(s.Time),
	}
}

func expectedSetFromDomain(expectedSet domain.ExpectedSet) expectedSetEntity {
	return expectedSetEntity{
		ID:            uuidToPgtype(expectedSet.ID),
		ExerciseLogID: uuidToPgtype(expectedSet.ExerciseLogID),
		SetType:       pgtype.Text{String: expectedSet.SetType.String(), Valid: expectedSet.SetType != domain.SetTypeUnknown},
		Reps:          pgtype.Int8{Int64: int64(expectedSet.Reps), Valid: expectedSet.Reps != 0},
		Weight:        pgtype.Float4{Float32: expectedSet.Weight, Valid: expectedSet.Weight != 0},
		Time:          intervalToPgtype(expectedSet.Time),
		CreatedAt:     timeToPgtype(expectedSet.CreatedAt),
		UpdatedAt:     timeToPgtype(expectedSet.UpdatedAt),
	}
}

func expectedSetsToDomain(expectedSets []expectedSetEntity) []domain.ExpectedSet {
	result := make([]domain.ExpectedSet, 0, len(expectedSets))
	for _, expectedSet := range expectedSets {
		result = append(result, expectedSet.toDomain())
	}

	return result
}

func (r *PGXRepository) GetExpectedSetsByExerciseLogID(ctx context.Context, exerciseLogID domain.ID) ([]domain.ExpectedSet, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetExpectedSetsByExerciseLogID")
	defer span.Finish()
	
	query := `
		SELECT id, exercise_log_id, set_type, reps, weight, time, created_at, updated_at
		FROM expected_sets es
		WHERE es.exercise_log_id = $1
		ORDER BY es.created_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var expectedSets []expectedSetEntity
	err := pgxscan.Select(ctx, engine, &expectedSets, query, uuidToPgtype(exerciseLogID))
	if err != nil {
		return nil, err
	}

	return expectedSetsToDomain(expectedSets), nil
}

func (r *PGXRepository) CreateExpectedSet(ctx context.Context, expectedSet domain.ExpectedSet) (domain.ExpectedSet, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateExpectedSet")
	defer span.Finish()

	query := `
		INSERT INTO expected_sets (id, exercise_log_id, set_type, reps, weight, time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, exercise_log_id, set_type, reps, weight, time, created_at, updated_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	expectedSetEntity := expectedSetFromDomain(expectedSet)
	err := pgxscan.Get(ctx, engine, &expectedSetEntity, query, expectedSet.ID, expectedSetEntity.ExerciseLogID, expectedSetEntity.SetType, expectedSetEntity.Reps, expectedSetEntity.Weight, expectedSetEntity.Time)
	if err != nil {
		return domain.ExpectedSet{}, err
	}

	return expectedSetEntity.toDomain(), nil
}

func (r *PGXRepository) DeleteExpectedSetsByExerciseLogID(ctx context.Context, exerciseLogID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteExpectedSetsByExerciseLogID")
	defer span.Finish()

	query := `
		DELETE FROM expected_sets
		WHERE exercise_log_id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	_, err := engine.Exec(ctx, query, uuidToPgtype(exerciseLogID))
	return err
}
