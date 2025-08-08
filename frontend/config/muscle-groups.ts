export const MUSCLE_GROUP_TRANSLATIONS: Record<string, string> = {
  chest: "грудь",
  lats: "широчайшие",
  shoulders: "плечи",
  biceps: "бицепсы",
  triceps: "трицепсы",
  forearms: "предплечья",
  abs: "пресс",
  quads: "квадрицепсы",
  hamstrings: "задняя поверхность бедра",
  calves: "икры",
  glutes: "ягодицы",
  "lower-back": "поясница",
  traps: "трапеции",
  abductors: "отводящие мышцы бедра",
};

export function translateMuscleGroup(key?: string | null): string {
  if (!key) return "";

  return MUSCLE_GROUP_TRANSLATIONS[key] || key;
}

export function translateMuscleGroups(
  keys?: (string | null | undefined)[],
): string[] {
  if (!keys) return [];

  return keys.map((k) => translateMuscleGroup(k || undefined));
}
