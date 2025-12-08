package repository

import (
	"context"
	"time"

	"fitness-trainer/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
)

func (r *PGXRepository) GetAnalyticsRawData(ctx context.Context, userID domain.ID, from, to time.Time, muscleGroupIDs []domain.ID, exerciseIDs []domain.ID) ([]domain.AnalyticsRawData, error) {
	engine := r.contextManager.GetEngineFromContext(ctx)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.Select(
		"w.finished_at",
		"e.id as exercise_id",
		"e.name as exercise_name",
		"mg.id as muscle_group_id",
		"mg.name as muscle_group_name",
		"sl.weight",
		"sl.reps",
	).
		From("workouts w").
		Join("exercise_logs el ON el.workout_id = w.id").
		Join("set_logs sl ON sl.exercise_log_id = el.id").
		Join("exercises e ON e.id = el.exercise_id").
		Join("exercise_muscle_groups emg ON emg.exercise_id = e.id").
		Join("muscle_groups mg ON mg.id = emg.muscle_group_id").
		Where(sq.Eq{"w.user_id": userID}).
		Where(sq.GtOrEq{"w.finished_at": from}).
		Where(sq.LtOrEq{"w.finished_at": to})

	if len(muscleGroupIDs) > 0 {
		query = query.Where(sq.Eq{"mg.id": muscleGroupIDs})
	}

	if len(exerciseIDs) > 0 {
		query = query.Where(sq.Eq{"e.id": exerciseIDs})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var data []domain.AnalyticsRawData
	if err := pgxscan.Select(ctx, engine, &data, sql, args...); err != nil {
		return nil, err
	}

	return data, nil
}
