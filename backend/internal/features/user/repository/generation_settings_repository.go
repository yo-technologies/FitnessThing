package repository

import (
	"context"
	"errors"
	"fitness-trainer/internal/shared/domain"
	"fmt"

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

type GenerationSettingsRepository struct {
	contextManager ContextManager
}

func NewGenerationSettingsRepository(contextManager ContextManager) *GenerationSettingsRepository {
	return &GenerationSettingsRepository{
		contextManager: contextManager,
	}
}

func (r *GenerationSettingsRepository) GetGenerationSettings(ctx context.Context, userID domain.ID) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GenerationSettingsRepository.GetGenerationSettings")
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

func (r *GenerationSettingsRepository) CreateOrUpdateGenerationSettings(ctx context.Context, settings domain.GenerationSettings) (domain.GenerationSettings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GenerationSettingsRepository.CreateOrUpdateGenerationSettings")
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
			variety_level = excluded.variety_level,
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