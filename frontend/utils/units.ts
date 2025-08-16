import { WorkoutWeightUnit } from "@/api/api.generated";

export const LB_TO_KG = 0.45359237;
export const KG_TO_LB = 1 / LB_TO_KG;

export function weightUnitLabel(unit?: WorkoutWeightUnit): string {
  switch (unit) {
    case WorkoutWeightUnit.WEIGHT_UNIT_LB:
      return "lb";
    case WorkoutWeightUnit.WEIGHT_UNIT_KG:
    default:
      return "кг";
  }
}

export function unitKey(unit?: WorkoutWeightUnit): "kg" | "lb" {
  return unit === WorkoutWeightUnit.WEIGHT_UNIT_LB ? "lb" : "kg";
}

export function unitFromKey(key: string): WorkoutWeightUnit {
  return key === "lb"
    ? WorkoutWeightUnit.WEIGHT_UNIT_LB
    : WorkoutWeightUnit.WEIGHT_UNIT_KG;
}

export function convertWeight(
  value: number,
  from?: WorkoutWeightUnit,
  to?: WorkoutWeightUnit,
): number {
  const v = Number(value) || 0;
  const fromUnit = from ?? WorkoutWeightUnit.WEIGHT_UNIT_KG;
  const toUnit = to ?? WorkoutWeightUnit.WEIGHT_UNIT_KG;

  if (fromUnit === toUnit) return v;

  if (
    fromUnit === WorkoutWeightUnit.WEIGHT_UNIT_LB &&
    toUnit === WorkoutWeightUnit.WEIGHT_UNIT_KG
  ) {
    return v * LB_TO_KG;
  }
  if (
    fromUnit === WorkoutWeightUnit.WEIGHT_UNIT_KG &&
    toUnit === WorkoutWeightUnit.WEIGHT_UNIT_LB
  ) {
    return v * KG_TO_LB;
  }

  return v;
}
