package mappers

import (
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/utils"
)

func nullableStringToOptionalProto(value utils.Nullable[string]) *string {
	if value.IsValid {
		return &value.V
	}
	return nil
}

func nullableIntToOptionalProto(value utils.Nullable[int]) *int32 {
	if value.IsValid {
		v := int32(value.V)
		return &v
	}
	return nil
}

func idsToStrings(ids []domain.ID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}
