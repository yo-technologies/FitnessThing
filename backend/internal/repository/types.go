package repository

import (
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/utils"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func timeToPgtype(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

func floatToPgtype(f float32) pgtype.Float4 {
	return pgtype.Float4{Float32: f, Valid: f != 0}
}

func uuidToPgtype(id domain.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.UUID(id), Valid: id != domain.ID{}}
}

func nullableIDToPgtype(id utils.Nullable[domain.ID]) pgtype.UUID {
	if !id.IsValid {
		return pgtype.UUID{Valid: false}
	}

	return pgtype.UUID{Bytes: uuid.UUID(id.V), Valid: true}
}

func durationFromPgtype(d pgtype.Interval) time.Duration {
	return time.Duration(d.Microseconds) * time.Microsecond
}

func intervalToPgtype(d time.Duration) pgtype.Interval {
	return pgtype.Interval{Microseconds: int64(d / time.Microsecond), Valid: d != 0}
}

func timeFromPgtype(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}

	return t.Time
}

func uuidsToPgtype(ids []domain.ID) []pgtype.UUID {
	result := make([]pgtype.UUID, 0, len(ids))
	for _, id := range ids {
		result = append(result, uuidToPgtype(id))
	}

	return result
}

func nullableStringToPgtype(s utils.Nullable[string]) pgtype.Text {
	if !s.IsValid {
		return pgtype.Text{Valid: false}
	}

	return pgtype.Text{String: s.V, Valid: true}
}

func nullableFloatToPgtype(f utils.Nullable[float32]) pgtype.Float4 {
	if !f.IsValid {
		return pgtype.Float4{Valid: false}
	}

	return pgtype.Float4{Float32: f.V, Valid: true}
}

func nullableIntToPgtype(i utils.Nullable[int]) pgtype.Int8 {
	if !i.IsValid {
		return pgtype.Int8{Valid: false}
	}

	return pgtype.Int8{Int64: int64(i.V), Valid: true}
}

func nullableStringFromPgtype(t pgtype.Text) utils.Nullable[string] {
	if !t.Valid {
		return utils.NewNullable("", false)
	}

	return utils.NewNullable(t.String, true)
}

func nullableIntFromPgtype(t pgtype.Int8) utils.Nullable[int] {
	if !t.Valid {
		return utils.NewNullable(0, false)
	}

	return utils.NewNullable(int(t.Int64), true)
}

func arrayToSlice[T any, U any](arr pgtype.FlatArray[T], f func(T) U) []U {
	result := make([]U, 0, len(arr.Dimensions()))
	for _, elem := range arr {
		result = append(result, f(elem))
	}

	return result
}

func sliceToArray[T any, U any](slice []T, f func(T) U) pgtype.FlatArray[U] {
	result := make(pgtype.FlatArray[U], 0, len(slice))
	for _, elem := range slice {
		result = append(result, f(elem))
	}

	return result
}
