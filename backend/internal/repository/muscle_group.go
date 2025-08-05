package repository

import (
	"context"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/logger"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type muscleGroupEntity struct {
	ID   pgtype.UUID
	Name string
}

func (m muscleGroupEntity) toDTO() dto.MuscleGroupDTO {
	return dto.MuscleGroupDTO{
		ID:   domain.ID(m.ID.Bytes),
		Name: m.Name,
	}
}

func (r *PGXRepository) GetMuscleGroups(ctx context.Context) ([]dto.MuscleGroupDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetMuscleGroups")
	defer span.Finish()

	query := `
		SELECT * FROM muscle_groups
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var muscleGroups []muscleGroupEntity
	if err := pgxscan.Select(ctx, engine, &muscleGroups, query); err != nil {
		logger.Errorf("failed to get muscle groups: %v", err)
		return nil, err
	}

	var muscleGroupsDTO []dto.MuscleGroupDTO
	for _, muscleGroup := range muscleGroups {
		muscleGroupsDTO = append(muscleGroupsDTO, muscleGroup.toDTO())
	}

	return muscleGroupsDTO, nil
}

func (r *PGXRepository) GetMuscleGroupByName(ctx context.Context, name string) (dto.MuscleGroupDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetMuscleGroupByName")
	defer span.Finish()

	query := `
		SELECT * FROM muscle_groups
		WHERE name = $1
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var muscleGroup muscleGroupEntity
	if err := pgxscan.Get(ctx, engine, &muscleGroup, query, name); err != nil {
		logger.Errorf("failed to get muscle group by name: %v", err)
		return dto.MuscleGroupDTO{}, err
	}

	return muscleGroup.toDTO(), nil
}

func (r *PGXRepository) GetMuscleGroupsByIDs(ctx context.Context, ids []domain.ID) ([]dto.MuscleGroupDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetMuscleGroupsByIDs")
	defer span.Finish()

	if len(ids) == 0 {
		return []dto.MuscleGroupDTO{}, nil
	}

	query := `
		SELECT * FROM muscle_groups
		WHERE id = ANY($1)
	`

	engine := r.contextManager.GetEngineFromContext(ctx)

	var muscleGroups []muscleGroupEntity
	if err := pgxscan.Select(ctx, engine, &muscleGroups, query, pgtype.FlatArray[pgtype.UUID](uuidsToPgtype(ids))); err != nil {
		logger.Errorf("failed to get muscle groups by IDs: %v", err)
			return nil, err
		}

	var muscleGroupsDTO []dto.MuscleGroupDTO
	for _, muscleGroup := range muscleGroups {
		muscleGroupsDTO = append(muscleGroupsDTO, muscleGroup.toDTO())
	}

	return muscleGroupsDTO, nil
}