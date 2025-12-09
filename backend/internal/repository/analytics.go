package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fitness-trainer/internal/domain"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

func (r *PGXRepository) GetAnalyticsRawData(ctx context.Context, userID domain.ID, from, to time.Time, muscleGroupIDs []domain.ID, exerciseIDs []domain.ID) ([]domain.AnalyticsRawData, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetAnalyticsRawData")
	defer span.Finish()

	engine := r.contextManager.GetEngineFromContext(ctx)

	if muscleGroupIDs == nil {
		muscleGroupIDs = []domain.ID{}
	}

	if exerciseIDs == nil {
		exerciseIDs = []domain.ID{}
	}

	query := strings.Builder{}
	query.WriteString(`
		SELECT DISTINCT ON (sl.id, mg.id)
			w.finished_at,
			e.id as exercise_id,
			e.name as exercise_name,
			mg.id as muscle_group_id,
			mg.name as muscle_group_name,
			sl.weight,
			sl.reps
		FROM workouts w
		INNER JOIN exercise_logs el ON el.workout_id = w.id
		INNER JOIN set_logs sl ON sl.exercise_log_id = el.id
		INNER JOIN exercises e ON e.id = el.exercise_id
		INNER JOIN exercise_muscle_groups emg ON emg.exercise_id = e.id
		INNER JOIN muscle_groups mg ON mg.id = emg.muscle_group_id
		WHERE w.user_id = $1
			AND w.finished_at IS NOT NULL
			AND w.finished_at >= $2
			AND w.finished_at <= $3
	`)

	args := []any{userID, timeToPgtype(from), timeToPgtype(to)}
	nextArg := 4

	if len(muscleGroupIDs) > 0 {
		query.WriteString(fmt.Sprintf("\t\t\tAND mg.id = ANY($%d)\n", nextArg))
		args = append(args, pgtype.FlatArray[pgtype.UUID](uuidsToPgtype(muscleGroupIDs)))
		nextArg++
	}

	if len(exerciseIDs) > 0 {
		query.WriteString(fmt.Sprintf("\t\t\tAND e.id = ANY($%d)\n", nextArg))
		args = append(args, pgtype.FlatArray[pgtype.UUID](uuidsToPgtype(exerciseIDs)))
		nextArg++
	}

	query.WriteString("\t\tORDER BY sl.id, mg.id, w.finished_at DESC\n")

	var data []domain.AnalyticsRawData
	if err := pgxscan.Select(
		ctx,
		engine,
		&data,
		query.String(),
		args...,
	); err != nil {
		return nil, err
	}

	return data, nil
}
