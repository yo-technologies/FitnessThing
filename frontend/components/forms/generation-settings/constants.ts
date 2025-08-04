import { OptionItem, OnboardingStep } from "./types";

import {
  WorkoutGoal,
  WorkoutExperienceLevel,
  WorkoutWorkoutPlanType,
} from "@/api/api.generated";

// Константы для селектов
export const GOAL_OPTIONS: OptionItem<WorkoutGoal>[] = [
  { key: WorkoutGoal.GOAL_UNSPECIFIED, label: "Не указано" },
  { key: WorkoutGoal.GOAL_MUSCLE_GAIN, label: "Набор мышечной массы" },
  { key: WorkoutGoal.GOAL_WEIGHT_LOSS, label: "Снижение веса" },
  { key: WorkoutGoal.GOAL_STRENGTH, label: "Увеличение силы" },
  { key: WorkoutGoal.GOAL_ENDURANCE, label: "Выносливость" },
  { key: WorkoutGoal.GOAL_FLEXIBILITY, label: "Гибкость" },
];

export const EXPERIENCE_LEVEL_OPTIONS: OptionItem<WorkoutExperienceLevel>[] = [
  {
    key: WorkoutExperienceLevel.EXPERIENCE_LEVEL_UNSPECIFIED,
    label: "Не указано",
  },
  {
    key: WorkoutExperienceLevel.EXPERIENCE_LEVEL_BEGINNER,
    label: "Новичок",
  },
  {
    key: WorkoutExperienceLevel.EXPERIENCE_LEVEL_INTERMEDIATE,
    label: "Средний уровень",
  },
  {
    key: WorkoutExperienceLevel.EXPERIENCE_LEVEL_ADVANCED,
    label: "Продвинутый",
  },
];

export const WORKOUT_PLAN_TYPE_OPTIONS: OptionItem<WorkoutWorkoutPlanType>[] = [
  {
    key: WorkoutWorkoutPlanType.WORKOUT_PLAN_TYPE_UNSPECIFIED,
    label: "Не указано",
  },
  {
    key: WorkoutWorkoutPlanType.WORKOUT_PLAN_TYPE_FULL_BODY,
    label: "Фулбоди",
  },
  {
    key: WorkoutWorkoutPlanType.WORKOUT_PLAN_TYPE_SPLIT,
    label: "Сплит",
  },
  {
    key: WorkoutWorkoutPlanType.WORKOUT_PLAN_TYPE_UPPER_LOWER,
    label: "Верх/Низ",
  },
  {
    key: WorkoutWorkoutPlanType.WORKOUT_PLAN_TYPE_PUSH_PULL_LEGS,
    label: "Толкай/Тяни/Ноги",
  },
];

// Шаги онбординга
export const ONBOARDING_STEPS: OnboardingStep[] = [
  {
    id: "welcome",
    title: "Добро пожаловать",
    description: "Настроим ваш персональный фитнес-план",
  },
  {
    id: "basics",
    title: "Основы",
    description: "Расскажите о ваших целях и опыте",
  },
  {
    id: "schedule",
    title: "Расписание",
    description: "Планирование тренировок",
  },
  {
    id: "limitations",
    title: "Ограничения",
    description: "Травмы и предпочтения",
  },
  {
    id: "personalization",
    title: "Персонализация",
    description: "Финальные настройки",
  },
];

// Утилиты
export const formatDuration = (minutes: number): string => {
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;

  if (hours > 0) {
    return `${hours}ч ${mins}м`;
  }

  return `${minutes}м`;
};

export const getVarietyLabel = (level: number): string => {
  switch (level) {
    case 1:
      return "Минимальное";
    case 2:
      return "Умеренное";
    case 3:
      return "Максимальное";
    default:
      return "Не указано";
  }
};

// Значения по умолчанию
export const DEFAULT_FORM_DATA = {
  primaryGoal: WorkoutGoal.GOAL_UNSPECIFIED,
  experienceLevel: WorkoutExperienceLevel.EXPERIENCE_LEVEL_UNSPECIFIED,
  workoutPlanType: WorkoutWorkoutPlanType.WORKOUT_PLAN_TYPE_UNSPECIFIED,
  daysPerWeek: 3,
  sessionDurationMinutes: 60,
  injuries: "",
  priorityMuscleGroupsIds: [],
  basePrompt: "",
  varietyLevel: 2,
};
