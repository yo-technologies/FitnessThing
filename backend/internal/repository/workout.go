package repository

import (
	"context"
	"errors"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type workoutEntity struct {
	ID         pgtype.UUID
	UserID     pgtype.UUID
	RoutineID  pgtype.UUID
	Notes      string
	Rating     int
	FinishedAt pgtype.Timestamptz
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
}

func (w workoutEntity) toDomain() domain.Workout {
	return domain.Workout{
		Model: domain.Model{
			ID:        domain.ID(w.ID.Bytes),
			CreatedAt: w.CreatedAt.Time,
			UpdatedAt: w.UpdatedAt.Time,
		},
		UserID:     domain.ID(w.UserID.Bytes),
		RoutineID:  utils.NewNullable(domain.ID(w.RoutineID.Bytes), w.RoutineID.Valid),
		Notes:      w.Notes,
		Rating:     w.Rating,
		FinishedAt: w.FinishedAt.Time,
	}
}

func workoutFromDomain(workout domain.Workout) workoutEntity {
	return workoutEntity{
		ID:         uuidToPgtype(workout.ID),
		UserID:     uuidToPgtype(workout.UserID),
		RoutineID:  pgtype.UUID{Bytes: uuid.UUID(workout.RoutineID.V), Valid: workout.RoutineID.IsValid},
		Notes:      workout.Notes,
		Rating:     workout.Rating,
		FinishedAt: timeToPgtype(workout.FinishedAt),
		CreatedAt:  timeToPgtype(workout.CreatedAt),
		UpdatedAt:  timeToPgtype(workout.UpdatedAt),
	}
}

func toWorkoutsDomain(workouts []workoutEntity) []domain.Workout {
	result := make([]domain.Workout, 0, len(workouts))
	for _, workout := range workouts {
		result = append(result, workout.toDomain())
	}

	return result
}

func (r *PGXRepository) CreateWorkout(ctx context.Context, workout domain.Workout) (domain.Workout, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateWorkout")
	defer span.Finish()

	query := `
		INSERT INTO workouts (id, user_id, routine_id, notes, rating, finished_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	entity := workoutFromDomain(workout)

	engine := r.contextManager.GetEngineFromContext(ctx)

	if err := pgxscan.Get(ctx, engine, &entity.CreatedAt, query,
		entity.ID,
		entity.UserID,
		entity.RoutineID,
		entity.Notes,
		entity.Rating,
		entity.FinishedAt,
	); err != nil {
		logger.Errorf("failed to create workout: %v", err)
		return domain.Workout{}, err
	}

	return entity.toDomain(), nil
}

func (r *PGXRepository) GetWorkoutByID(ctx context.Context, id domain.ID) (domain.Workout, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetWorkoutByID")
	defer span.Finish()

	query := `
		SELECT id, user_id, routine_id, created_at, notes, rating, finished_at, updated_at
		FROM workouts
		WHERE id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var workout workoutEntity
	if err := pgxscan.Get(ctx, engine, &workout, query, uuidToPgtype(id)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Workout{}, domain.ErrNotFound
		}
		logger.Errorf("failed to get workout by id: %v", err)
		return domain.Workout{}, domain.ErrInternal
	}

	return workout.toDomain(), nil
}

func (r *PGXRepository) GetActiveWorkouts(ctx context.Context, userID domain.ID) ([]domain.Workout, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetActiveWorkouts")
	defer span.Finish()

	query := `
		SELECT id, user_id, routine_id, created_at, notes, rating, finished_at, updated_at
		FROM workouts
		WHERE user_id = $1 AND finished_at IS NULL
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var workouts []workoutEntity
	if err := pgxscan.Select(ctx, engine, &workouts, query, uuidToPgtype(userID)); err != nil {
		logger.Errorf("failed to get active workouts: %v", err)
		return nil, domain.ErrInternal
	}

	return toWorkoutsDomain(workouts), nil
}

func (r *PGXRepository) UpdateWorkout(ctx context.Context, id domain.ID, workout domain.Workout) (domain.Workout, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateWorkout")
	defer span.Finish()

	workoutEntity := workoutFromDomain(workout)
	query := `
		UPDATE workouts
		SET notes = $1, rating = $2, finished_at = $3, updated_at = now()
		WHERE id = $4
		RETURNING updated_at
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	if err := pgxscan.Get(ctx, engine, &workoutEntity.UpdatedAt, query,
		workoutEntity.Notes,
		workoutEntity.Rating,
		workoutEntity.FinishedAt,
		uuidToPgtype(id),
	); err != nil {
		logger.Errorf("failed to update workout: %v", err)
		return domain.Workout{}, err
	}

	return workoutEntity.toDomain(), nil
}

func (r *PGXRepository) DeleteWorkout(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteWorkout")
	defer span.Finish()

	query := `
		DELETE FROM workouts
		WHERE id = $1
		RETURNING id
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var workout workoutEntity
	if err := pgxscan.Get(ctx, engine, &workout, query, uuidToPgtype(id)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		logger.Errorf("failed to delete workout: %v", err)
		return domain.ErrInternal
	}

	return nil
}

func (r *PGXRepository) GetWorkouts(ctx context.Context, userID domain.ID, limit, offset int) ([]domain.Workout, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetWorkouts")
	defer span.Finish()

	query := `
		SELECT id, user_id, routine_id, created_at, notes, rating, finished_at, updated_at
		FROM workouts
		WHERE user_id = $1 AND finished_at IS NOT NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var workouts []workoutEntity
	if err := pgxscan.Select(ctx, engine, &workouts, query, uuidToPgtype(userID), limit, offset); err != nil {
		logger.Errorf("failed to get workouts: %v", err)
		return nil, domain.ErrInternal
	}

	return toWorkoutsDomain(workouts), nil
}
