import {
  WorkoutGoal,
  WorkoutExperienceLevel,
  WorkoutWorkoutPlanType,
} from "@/api/api.generated";

// Типы для опций
export interface OptionItem<T> {
  key: T;
  label: string;
}

// Интерфейсы для компонентов
export interface SectionHeaderProps {
  title: string;
  tooltip: string;
  badge?: React.ReactNode;
}

export interface ChipSelectorProps<T> {
  title: string;
  tooltip: string;
  options: OptionItem<T>[];
  value: T;
  color: "primary" | "secondary" | "success" | "warning" | "danger" | "default";
  onChange: (value: T) => void;
}

export interface SliderWithMarksProps {
  title: string;
  tooltip: string;
  value: number;
  minValue: number;
  maxValue: number;
  step: number;
  marks: Array<{ value: number; label: string }>;
  color: "primary" | "secondary" | "success" | "warning" | "danger" | "default";
  formatValue: (value: number) => string;
  onChange: (value: number | number[]) => void;
}

export interface TextFieldProps {
  title: string;
  tooltip: string;
  value: string;
  placeholder: string;
  minRows?: number;
  maxRows?: number;
  badge?: React.ReactNode;
  onChange: (value: string) => void;
  onBlur?: () => void;
}

export interface MultiChipSelectorProps {
  title: string;
  tooltip: string;
  options: Array<{ id: string; name: string }>;
  selectedIds: string[];
  color: "primary" | "secondary" | "success" | "warning" | "danger" | "default";
  onChange: (selectedIds: string[]) => void;
  onClear: () => void;
}

// Типы для шагов онбординга
export interface OnboardingStep {
  id: string;
  title: string;
  description: string;
}

export interface OnboardingFormData {
  primaryGoal: WorkoutGoal;
  experienceLevel: WorkoutExperienceLevel;
  workoutPlanType: WorkoutWorkoutPlanType;
  daysPerWeek: number;
  sessionDurationMinutes: number;
  injuries: string;
  priorityMuscleGroupsIds: string[];
  basePrompt: string;
  varietyLevel: number;
}
