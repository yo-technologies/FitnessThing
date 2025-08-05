package repository

import (
	"context"
	"errors"
	"fitness-trainer/internal/domain"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type llmSettings struct {
	ID        pgtype.UUID        `db:"id"`
	CreatedAt pgtype.Timestamptz `db:"created_at"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at"`

	UserID pgtype.UUID `db:"user_id"`

	BasePrompt              pgtype.Text                   `db:"base_prompt"`
	VarietyLevel            pgtype.Int8                   `db:"variety_level"`
	PrimaryGoal             pgtype.Int8                   `db:"primary_goal"`
	SecondaryGoals          pgtype.FlatArray[pgtype.Text] `db:"secondary_goals"`
	ExperienceLevel         pgtype.Int8                   `db:"experience_level"`
	DaysPerWeek             pgtype.Int8                   `db:"days_per_week"`
	SessionDurationMinutes  pgtype.Int8                   `db:"session_duration_minutes"`
	Injuries                pgtype.Text                   `db:"injuries"`
	PriorityMuscleGroupsIDs pgtype.FlatArray[pgtype.UUID] `db:"priority_muscle_groups"`
	WorkoutPlanType         pgtype.Int8                   `db:"workout_plan_type"`
	Hash                    pgtype.Text                   `db:"hash"`
}

func (e llmSettings) toDomain() domain.GenerationSettings {
	return domain.GenerationSettings{
		Model: domain.Model{
			ID:        domain.ID(e.ID.Bytes),
			CreatedAt: e.CreatedAt.Time,
			UpdatedAt: e.UpdatedAt.Time,
		},
		UserID: domain.ID(e.UserID.Bytes),

		BasePrompt:              nullableStringFromPgtype(e.BasePrompt),
		VarietyLevel:            nullableIntFromPgtype(e.VarietyLevel),
		PrimaryGoal:             domain.Goal(e.PrimaryGoal.Int64),
		SecondaryGoals:          arrayToSlice(e.SecondaryGoals, func(t pgtype.Text) string { return t.String }),
		ExperienceLevel:         domain.ExperienceLevel(e.ExperienceLevel.Int64),
		DaysPerWeek:             nullableIntFromPgtype(e.DaysPerWeek),
		SessionDurationMinutes:  nullableIntFromPgtype(e.SessionDurationMinutes),
		Injuries:                nullableStringFromPgtype(e.Injuries),
		PriorityMuscleGroupsIDs: arrayToSlice(e.PriorityMuscleGroupsIDs, func(u pgtype.UUID) domain.ID { return domain.ID(u.Bytes) }),
		WorkoutPlanType:         domain.WorkoutPlanType(e.WorkoutPlanType.Int64),
		Hash:                    e.Hash.String,
	}
}

func llmSettingsFromDomain(settings domain.GenerationSettings) llmSettings {
	return llmSettings{
		ID:        uuidToPgtype(settings.ID),
		CreatedAt: timeToPgtype(settings.CreatedAt),
		UpdatedAt: timeToPgtype(settings.UpdatedAt),

		UserID: uuidToPgtype(settings.UserID),

		BasePrompt:              nullableStringToPgtype(settings.BasePrompt),
		VarietyLevel:            nullableIntToPgtype(settings.VarietyLevel),
		PrimaryGoal:             pgtype.Int8{Int64: int64(settings.PrimaryGoal), Valid: true},
		SecondaryGoals:          sliceToArray(settings.SecondaryGoals, func(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }),
		ExperienceLevel:         pgtype.Int8{Int64: int64(settings.ExperienceLevel), Valid: true},
		DaysPerWeek:             nullableIntToPgtype(settings.DaysPerWeek),
		SessionDurationMinutes:  nullableIntToPgtype(settings.SessionDurationMinutes),
		Injuries:                nullableStringToPgtype(settings.Injuries),
		PriorityMuscleGroupsIDs: sliceToArray(settings.PriorityMuscleGroupsIDs, func(id domain.ID) pgtype.UUID { return uuidToPgtype(id) }),
		WorkoutPlanType:         pgtype.Int8{Int64: int64(settings.WorkoutPlanType), Valid: true},
		Hash:                    pgtype.Text{String: settings.Hash, Valid: true},
	}
}

func llmSettingsToDomainSlice(settings []llmSettings) []domain.GenerationSettings {
	result := make([]domain.GenerationSettings, 0, len(settings))
	for _, s := range settings {
		result = append(result, s.toDomain())
	}
	return result
}

func (r *PGXRepository) GetGenerationSettings(ctx context.Context, userID domain.ID) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetGenerationSettings")
	defer span.Finish()

	const query = `
		SELECT 
			id, 
			user_id, 
			created_at, 
			updated_at, 
			base_prompt, 
			variety_level, 
			primary_goal, 
			secondary_goals, 
			experience_level, 
			days_per_week, 
			session_duration_minutes, 
			injuries, 
			priority_muscle_groups, 
			workout_plan_type,
			hash
		FROM llm_settings
		WHERE user_id = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var settings llmSettings
	if err := pgxscan.Get(ctx, engine, &settings, query, uuidToPgtype(userID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GenerationSettings{}, domain.ErrNotFound
		}
		return domain.GenerationSettings{}, fmt.Errorf("failed to get generation settings: %w", err)
	}

	return settings.toDomain(), nil
}

func (r *PGXRepository) CreateOrUpdateGenerationSettings(ctx context.Context, settings domain.GenerationSettings) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.SaveGenerationSettings")
	defer span.Finish()

	const query = `
		INSERT INTO 
			llm_settings (
				id, 
				user_id, 
				created_at, 
				updated_at,
				base_prompt, 
				variety_level, 
				primary_goal,
				secondary_goals,
				experience_level,
				days_per_week,
				session_duration_minutes,
				injuries,
				priority_muscle_groups,
				workout_plan_type,
				hash
			)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (user_id) DO UPDATE
		SET 
			base_prompt = excluded.base_prompt,
			primary_goal = excluded.primary_goal,
			secondary_goals = excluded.secondary_goals,
			experience_level = excluded.experience_level,
			days_per_week = excluded.days_per_week,
			session_duration_minutes = excluded.session_duration_minutes,
			injuries = excluded.injuries,
			priority_muscle_groups = excluded.priority_muscle_groups,
			workout_plan_type = excluded.workout_plan_type,
			hash = excluded.hash,
			updated_at = excluded.updated_at
		RETURNING 
			id, 
			user_id, 
			created_at, 
			updated_at,
			base_prompt, 
			variety_level, 
			primary_goal,
			secondary_goals,
			experience_level,
			days_per_week,
			session_duration_minutes,
			injuries,
			priority_muscle_groups,
			workout_plan_type,
			hash
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	settingsEntity := llmSettingsFromDomain(settings)

	if err := pgxscan.Get(ctx, engine, &settingsEntity, query,
		settingsEntity.ID,
		settingsEntity.UserID,
		settingsEntity.CreatedAt,
		settingsEntity.UpdatedAt,
		settingsEntity.BasePrompt,
		settingsEntity.VarietyLevel,
		settingsEntity.PrimaryGoal,
		settingsEntity.SecondaryGoals,
		settingsEntity.ExperienceLevel,
		settingsEntity.DaysPerWeek,
		settingsEntity.SessionDurationMinutes,
		settingsEntity.Injuries,
		settingsEntity.PriorityMuscleGroupsIDs,
		settingsEntity.WorkoutPlanType,
		settingsEntity.Hash,
	); err != nil {
		return domain.GenerationSettings{}, fmt.Errorf("failed to create or update generation settings: %w", err)
	}

	return settingsEntity.toDomain(), nil
}

func (r *PGXRepository) ListGenerationSettingsToProcess(
	ctx context.Context,
	debounceTime time.Duration,
) ([]domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListGenerationSettingsToProcess")
	defer span.Finish()

	const query = `
		SELECT 
			us.id,
			us.created_at,
			us.updated_at,
			us.user_id,
			us.base_prompt,
			us.variety_level,
			us.primary_goal,
			us.secondary_goals,
			us.experience_level,
			us.days_per_week,
			us.session_duration_minutes,
			us.injuries,
			us.priority_muscle_groups,
			us.workout_plan_type,
			us.hash
		FROM llm_settings us
		LEFT JOIN LATERAL (
			SELECT p.settings_hash, p.created_at
			FROM prompts p
			WHERE p.user_id = us.user_id
			ORDER BY p.created_at DESC
			LIMIT 1
		) last_prompt ON TRUE
	WHERE 
		(
			last_prompt.settings_hash IS NULL
			OR last_prompt.settings_hash != us.hash
			OR last_prompt.created_at < us.updated_at
		)
		AND us.updated_at < NOW() - $1::interval
	ORDER BY us.updated_at ASC
`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var settings []llmSettings
	if err := pgxscan.Select(
		ctx,
		engine,
		&settings,
		query,
		intervalToPgtype(debounceTime),
	); err != nil {
		return nil, fmt.Errorf("failed to list generation settings to process: %w", err)
	}

	return llmSettingsToDomainSlice(settings), nil
}
