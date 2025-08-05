package repository

import (
	"context"
	"errors"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type userEntity struct {
	ID pgtype.UUID

	TelegramID int64       `db:"telegram_id"`
	Username   pgtype.Text `db:"username"`

	FirstName pgtype.Text `db:"first_name"`
	LastName  pgtype.Text `db:"last_name"`

	PictureProfileURL pgtype.Text `db:"picture_profile_url"`

	DateOfBirth pgtype.Timestamptz `db:"date_of_birth"`

	Weight pgtype.Float4 `db:"weight"`
	Height pgtype.Float4 `db:"height"`

	HasCompletedOnboarding pgtype.Bool `db:"has_completed_onboarding"`

	CreatedAt pgtype.Timestamptz `db:"created_at"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at"`
}

func (u userEntity) toDomain() domain.User {
	return domain.User{
		Model: domain.Model{
			ID:        domain.ID(u.ID.Bytes),
			CreatedAt: timeFromPgtype(u.CreatedAt),
			UpdatedAt: timeFromPgtype(u.UpdatedAt),
		},
		TelegramID:             u.TelegramID,
		TelegramUsername:       nullableStringFromPgtype(u.Username),
		FirstName:              nullableStringFromPgtype(u.FirstName),
		LastName:               nullableStringFromPgtype(u.LastName),
		DateOfBirth:            timeFromPgtype(u.DateOfBirth),
		Weight:                 utils.NewNullable(u.Weight.Float32, u.Weight.Valid),
		Height:                 utils.NewNullable(u.Height.Float32, u.Height.Valid),
		ProfilePicURL:          utils.NewNullable(u.PictureProfileURL.String, u.PictureProfileURL.Valid),
		HasCompletedOnboarding: u.HasCompletedOnboarding.Bool,
	}
}

func userFromDomain(user domain.User) userEntity {
	return userEntity{
		ID:                     uuidToPgtype(user.ID),
		TelegramID:             user.TelegramID,
		Username:               nullableStringToPgtype(user.TelegramUsername),
		FirstName:              nullableStringToPgtype(user.FirstName),
		LastName:               nullableStringToPgtype(user.LastName),
		DateOfBirth:            timeToPgtype(user.DateOfBirth),
		Weight:                 nullableFloatToPgtype(user.Weight),
		Height:                 nullableFloatToPgtype(user.Height),
		CreatedAt:              timeToPgtype(user.CreatedAt),
		UpdatedAt:              timeToPgtype(user.UpdatedAt),
		PictureProfileURL:      nullableStringToPgtype(user.ProfilePicURL),
		HasCompletedOnboarding: pgtype.Bool{Bool: user.HasCompletedOnboarding, Valid: true},
	}
}

func (r *PGXRepository) GetUserByID(ctx context.Context, id domain.ID) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetUserByID")
	defer span.Finish()

	const query = `
		select id, telegram_id, username, first_name, last_name, date_of_birth, height, weight, created_at, updated_at, picture_profile_url, has_completed_onboarding
		from users
		where id = $1;
	`

	var user userEntity

	engine := r.contextManager.GetEngineFromContext(ctx)

	err := pgxscan.Get(ctx, engine, &user, query, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, domain.ErrNotFound
		}
		logger.Errorf("error getting user by id: %v", err)
	}

	return user.toDomain(), nil
}

func (r *PGXRepository) GetOrCreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetOrCreateUser")
	defer span.Finish()

	const query = `
		insert into users (id, telegram_id, username, first_name, last_name, date_of_birth, height, weight, created_at, updated_at, picture_profile_url, has_completed_onboarding)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		on conflict (telegram_id) do update
		set 
			username = excluded.username,
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			updated_at = excluded.updated_at,
			picture_profile_url = excluded.picture_profile_url
		returning id, telegram_id, username, first_name, last_name, date_of_birth, height, weight, created_at, updated_at, picture_profile_url, has_completed_onboarding;
	`

	userEntity := userFromDomain(user)

	engine := r.contextManager.GetEngineFromContext(ctx)

	err := pgxscan.Get(ctx, engine, &userEntity, query,
		userEntity.ID,
		userEntity.TelegramID,
		userEntity.Username,
		userEntity.FirstName,
		userEntity.LastName,
		userEntity.DateOfBirth,
		userEntity.Height,
		userEntity.Weight,
		userEntity.CreatedAt,
		userEntity.UpdatedAt,
		userEntity.PictureProfileURL,
		userEntity.HasCompletedOnboarding,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return domain.User{}, domain.ErrAlreadyExists
			}
		}
		logger.Errorf("error creating user: %v", err)
		return domain.User{}, err
	}

	return userEntity.toDomain(), nil
}

func (r *PGXRepository) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateUser")
	defer span.Finish()

	const query = `
		update users
		set 
			username = $3,
			first_name = $4,
			last_name = $5,
			date_of_birth = $6,
			height = $7,
			weight = $8,
			updated_at = $9,
			picture_profile_url = $10,
			has_completed_onboarding = $11
		where id = $1 and telegram_id = $2
		returning id, telegram_id, username, first_name, last_name, date_of_birth, height, weight, created_at, updated_at, picture_profile_url, has_completed_onboarding;
	`

	userEntity := userFromDomain(user)

	engine := r.contextManager.GetEngineFromContext(ctx)

	err := pgxscan.Get(
		ctx, engine, &userEntity, query,
		userEntity.ID,
		userEntity.TelegramID,
		userEntity.Username,
		userEntity.FirstName,
		userEntity.LastName,
		userEntity.DateOfBirth,
		userEntity.Height,
		userEntity.Weight,
		userEntity.UpdatedAt,
		userEntity.PictureProfileURL,
		userEntity.HasCompletedOnboarding,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}

		logger.Errorf("error updating user: %v", err)
		return domain.User{}, err
	}

	return userEntity.toDomain(), nil
}
