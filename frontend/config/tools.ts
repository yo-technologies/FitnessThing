export type ToolMeta = {
  id: string;
  label: string;
  description?: string;
};

// Словарь инструментов: ключ — техническое имя инструмента из backend, значение — человекочитаемые метки
export const TOOL_DICTIONARY: Record<string, ToolMeta> = {
  list_muscle_groups: {
    id: "list_muscle_groups",
    label: "Список групп мышц",
    description: "Возвращает доступные группы мышц",
  },
  list_exercises: {
    id: "list_exercises",
    label: "Список упражнений",
    description: "Подбор упражнений по критериям",
  },
  get_workout_history: {
    id: "get_workout_history",
    label: "История тренировок",
    description: "История ваших тренировок",
  },
  get_exercise_history: {
    id: "get_exercise_history",
    label: "История упражнения",
    description: "История выполнения выбранного упражнения",
  },
  get_workout_plan: {
    id: "get_workout_plan",
    label: "План тренировки",
    description: "Получение актуального плана тренировки",
  },
  add_exercises_to_workout: {
    id: "add_exercises_to_workout",
    label: "Добавление упражнений",
    description: "Добавление упражнений в тренировку",
  },
  remove_exercise_logs_from_workout: {
    id: "remove_exercise_logs_from_workout",
    label: "Удаление упражнений",
    description: "Удаление записей упражнений из тренировки",
  },
  replace_expected_sets: {
    id: "replace_expected_sets",
    label: "Установка подходов",
    description: "Полная замена ожидаемых подходов",
  },
};

export function getToolMeta(id?: string | null): ToolMeta | undefined {
  if (!id) return undefined;

  return TOOL_DICTIONARY[id] ?? undefined;
}

export function formatToolLabel(id?: string | null): string {
  if (!id || id.trim().length === 0) return "Инструмент";
  const meta = getToolMeta(id);

  return meta?.label ?? id;
}
