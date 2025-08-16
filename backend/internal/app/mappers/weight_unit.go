package mappers

import (
    "fitness-trainer/internal/domain"
    desc "fitness-trainer/pkg/workouts"
)

func weightUnitToProto(u domain.WeightUnit) desc.WeightUnit {
    switch u {
    case domain.WeightUnitKG:
        return desc.WeightUnit_WEIGHT_UNIT_KG
    case domain.WeightUnitLB:
        return desc.WeightUnit_WEIGHT_UNIT_LB
    default:
        return desc.WeightUnit_WEIGHT_UNIT_UNSPECIFIED
    }
}

func WeightUnitFromProto(u desc.WeightUnit) domain.WeightUnit {
    switch u {
    case desc.WeightUnit_WEIGHT_UNIT_KG:
        return domain.WeightUnitKG
    case desc.WeightUnit_WEIGHT_UNIT_LB:
        return domain.WeightUnitLB
    default:
        return domain.WeightUnitUnknown
    }
}
