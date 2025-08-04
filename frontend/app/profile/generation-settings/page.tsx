"use client";

import { useEffect, useState, useCallback } from "react";
import {
  Card,
  CardBody,
  Popover,
  PopoverContent,
  PopoverTrigger,
  Slider,
  Textarea,
  Chip,
  Divider,
  Button,
} from "@nextui-org/react";
import { toast } from "react-toastify";
import { debounce } from "lodash";

import { PageHeader } from "@/components/page-header";
import { Loading } from "@/components/loading";
import { authApi } from "@/api/api";
import {
  WorkoutWorkoutGenerationSettings,
  WorkoutGoal,
  WorkoutExperienceLevel,
  WorkoutWorkoutPlanType,
  WorkoutMuscleGroup,
} from "@/api/api.generated";
import { CircleQuestionIcon } from "@/config/icons";

// Типы для опций
interface OptionItem<T> {
  key: T;
  label: string;
}

// Константы для селектов
const GOAL_OPTIONS: OptionItem<WorkoutGoal>[] = [
  { key: WorkoutGoal.GOAL_MUSCLE_GAIN, label: "Набор мышечной массы" },
  { key: WorkoutGoal.GOAL_WEIGHT_LOSS, label: "Снижение веса" },
  { key: WorkoutGoal.GOAL_STRENGTH, label: "Увеличение силы" },
  { key: WorkoutGoal.GOAL_ENDURANCE, label: "Выносливость" },
  { key: WorkoutGoal.GOAL_FLEXIBILITY, label: "Гибкость" },
];

const EXPERIENCE_LEVEL_OPTIONS: OptionItem<WorkoutExperienceLevel>[] = [
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

const WORKOUT_PLAN_TYPE_OPTIONS: OptionItem<WorkoutWorkoutPlanType>[] = [
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

// Утилиты
const formatDuration = (minutes: number): string => {
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;

  if (hours > 0) {
    return `${hours}ч ${mins}м`;
  }

  return `${minutes}м`;
};

const getVarietyLabel = (level: number): string => {
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

// Компонент для заголовка секции с подсказкой
interface SectionHeaderProps {
  title: string;
  tooltip: string;
  badge?: React.ReactNode;
}

function SectionHeader({ title, tooltip, badge }: SectionHeaderProps) {
  return (
    <div className="flex flex-row gap-2 items-center justify-between">
      <div className="flex items-center gap-2">
        <p className="text-medium font-medium">{title}</p>
        {badge}
      </div>
      <Popover backdrop="opaque" size="sm">
        <PopoverTrigger>
          <CircleQuestionIcon className="w-4 h-4 text-default-500 cursor-pointer" />
        </PopoverTrigger>
        <PopoverContent>
          <p className="text-xs font-light text-default-600 p-1">{tooltip}</p>
        </PopoverContent>
      </Popover>
    </div>
  );
}

// Компонент для выбора чипами
interface ChipSelectorProps<T> {
  title: string;
  tooltip: string;
  options: OptionItem<T>[];
  value: T;
  color: "primary" | "secondary" | "success" | "warning" | "danger" | "default";
  onChange: (value: T) => void;
}

function ChipSelector<T>({
  title,
  tooltip,
  options,
  value,
  color,
  onChange,
}: ChipSelectorProps<T>) {
  return (
    <Card>
      <CardBody className="flex flex-col gap-4">
        <SectionHeader title={title} tooltip={tooltip} />
        <div className="flex flex-wrap gap-2">
          {options.map((option) => (
            <Chip
              key={String(option.key)}
              className="cursor-pointer transition-all"
              color={value === option.key ? color : "default"}
              variant={value === option.key ? "solid" : "bordered"}
              onClick={() => onChange(option.key)}
            >
              {option.label}
            </Chip>
          ))}
        </div>
      </CardBody>
    </Card>
  );
}

// Компонент для слайдера с метками
interface SliderWithMarksProps {
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

function SliderWithMarks({
  title,
  tooltip,
  value,
  minValue,
  maxValue,
  step,
  marks,
  color,
  formatValue,
  onChange,
}: SliderWithMarksProps) {
  // Преобразуем цвет для Chip и Slider
  const chipColor = color === "default" ? "primary" : color;
  const sliderColor = color === "default" ? "foreground" : color;

  return (
    <Card>
      <CardBody className="flex flex-col gap-4">
        <SectionHeader
          badge={
            <Chip color={chipColor} size="sm" variant="flat">
              {formatValue(value)}
            </Chip>
          }
          title={title}
          tooltip={tooltip}
        />
        <Slider
          className="w-full"
          color={sliderColor}
          marks={marks}
          maxValue={maxValue}
          minValue={minValue}
          step={step}
          value={value}
          onChange={onChange}
        />
      </CardBody>
    </Card>
  );
}

// Компонент для текстового поля
interface TextFieldProps {
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

function TextField({
  title,
  tooltip,
  value,
  placeholder,
  minRows = 2,
  maxRows = 4,
  badge,
  onChange,
  onBlur,
}: TextFieldProps) {
  return (
    <Card>
      <CardBody className="flex flex-col gap-4">
        <SectionHeader badge={badge} title={title} tooltip={tooltip} />
        <Textarea
          classNames={{
            input:
              "text-sm placeholder:text-default-400 placeholder:text-xs placeholder:font-light",
          }}
          maxRows={maxRows}
          minRows={minRows}
          placeholder={placeholder}
          value={value}
          variant="bordered"
          onBlur={onBlur}
          onValueChange={onChange}
        />
      </CardBody>
    </Card>
  );
}

// Компонент для множественного выбора чипами
interface MultiChipSelectorProps {
  title: string;
  tooltip: string;
  options: Array<{ id: string; name: string }>;
  selectedIds: string[];
  color: "primary" | "secondary" | "success" | "warning" | "danger" | "default";
  onChange: (selectedIds: string[]) => void;
  onClear: () => void;
}

function MultiChipSelector({
  title,
  tooltip,
  options,
  selectedIds,
  color,
  onChange,
  onClear,
}: MultiChipSelectorProps) {
  const handleToggle = (id: string) => {
    const newSelection = selectedIds.includes(id)
      ? selectedIds.filter((selectedId) => selectedId !== id)
      : [...selectedIds, id];

    onChange(newSelection);
  };

  return (
    <Card>
      <CardBody className="flex flex-col gap-2">
        <SectionHeader title={title} tooltip={tooltip} />

        {selectedIds.length > 0 ? (
          <div className="flex justify-between items-center">
            <p className="text-xs text-default-500">
              Выбрано: {selectedIds.length}
            </p>
            <Button
              className="p-0"
              color="danger"
              size="sm"
              variant="light"
              onPress={onClear}
            >
              Очистить
            </Button>
          </div>
        ) : (
          <div />
        )}

        <div className="flex flex-wrap gap-2">
          {options.map((option) => (
            <Chip
              key={option.id}
              className="cursor-pointer transition-all hover:scale-105"
              color={selectedIds.includes(option.id) ? color : "default"}
              variant={selectedIds.includes(option.id) ? "solid" : "bordered"}
              onClick={() => handleToggle(option.id)}
            >
              {option.name}
            </Chip>
          ))}
        </div>

        {options.length === 0 && (
          <div className="text-center py-4 text-default-500">
            <p>Загрузка опций...</p>
          </div>
        )}
      </CardBody>
    </Card>
  );
}

export default function RecordsPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);
  const [muscleGroups, setMuscleGroups] = useState<WorkoutMuscleGroup[]>([]);

  const [settings, setSettings] = useState<WorkoutWorkoutGenerationSettings>(
    {},
  );

  async function fetchMuscleGroups() {
    try {
      const response = await authApi.v1.exerciseServiceGetMuscleGroups();

      setMuscleGroups(response.data.muscleGroups || []);
    } catch (error) {
      console.log(error);
      toast.error("Не удалось загрузить группы мышц");
    }
  }

  async function fetchSettings() {
    await authApi.v1
      .userServiceGetWorkoutGenerationSettings()
      .then((response) => {
        console.log(response.data);
        setSettings(response.data.settings!);
      })
      .catch((error) => {
        console.log(error);
        toast.error("Не удалось загрузить настройки");
      });
  }

  async function fetchData() {
    setIsLoading(true);
    try {
      await Promise.all([fetchSettings(), fetchMuscleGroups()]);
    } catch (error) {
      console.log(error);
      setIsError(true);
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    fetchData();
  }, []);

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return (
      <div className="p-4">
        <h2 className="text-lg text-red-500">Ошибка при загрузке данных</h2>
        <p>Проверьте соединение с сервером или обновите страницу.</p>
      </div>
    );
  }

  function SettingsFrom() {
    async function updateSettings(
      updates: Partial<WorkoutWorkoutGenerationSettings>,
    ) {
      try {
        await authApi.v1.userServiceUpdateWorkoutGenerationSettings(updates);
        setSettings({ ...settings, ...updates });
      } catch (error) {
        console.log(error);
        toast.error("Не удалось обновить настройки");
      }
    }

    function PrimaryGoalSelect() {
      const [primaryGoal, setPrimaryGoal] = useState<WorkoutGoal>(
        settings.primaryGoal || WorkoutGoal.GOAL_UNSPECIFIED,
      );

      const handleChange = (goal: WorkoutGoal) => {
        setPrimaryGoal(goal);
        updateSettings({ primaryGoal: goal });
      };

      return (
        <ChipSelector
          color="primary"
          options={GOAL_OPTIONS}
          title="Основная цель:"
          tooltip="Выберите основную цель ваших тренировок."
          value={primaryGoal}
          onChange={handleChange}
        />
      );
    }

    function ExperienceLevelSelect() {
      const [experienceLevel, setExperienceLevel] =
        useState<WorkoutExperienceLevel>(
          settings.experienceLevel ||
            WorkoutExperienceLevel.EXPERIENCE_LEVEL_UNSPECIFIED,
        );

      const handleChange = (level: WorkoutExperienceLevel) => {
        setExperienceLevel(level);
        updateSettings({ experienceLevel: level });
      };

      return (
        <ChipSelector
          color="success"
          options={EXPERIENCE_LEVEL_OPTIONS}
          title="Уровень опыта:"
          tooltip="Укажите ваш уровень опыта в тренировках."
          value={experienceLevel}
          onChange={handleChange}
        />
      );
    }

    function WorkoutPlanTypeSelect() {
      const [workoutPlanType, setWorkoutPlanType] =
        useState<WorkoutWorkoutPlanType>(
          settings.workoutPlanType ||
            WorkoutWorkoutPlanType.WORKOUT_PLAN_TYPE_UNSPECIFIED,
        );

      const handleChange = (planType: WorkoutWorkoutPlanType) => {
        setWorkoutPlanType(planType);
        updateSettings({ workoutPlanType: planType });
      };

      return (
        <ChipSelector
          color="secondary"
          options={WORKOUT_PLAN_TYPE_OPTIONS}
          title="Тип тренировочного плана:"
          tooltip="Выберите предпочитаемый тип тренировочного плана."
          value={workoutPlanType}
          onChange={handleChange}
        />
      );
    }

    function DaysPerWeekSlider() {
      const [daysPerWeek, setDaysPerWeek] = useState<number>(
        settings.daysPerWeek || 3,
      );

      // Создаем debounced функцию для обновления настроек
      const debouncedUpdate = useCallback(
        debounce((days: number) => {
          updateSettings({ daysPerWeek: days });
        }, 500),
        [updateSettings],
      );

      const handleChange = (value: number | number[]) => {
        const days = value as number;

        setDaysPerWeek(days);
        debouncedUpdate(days);
      };

      return (
        <SliderWithMarks
          color="primary"
          formatValue={(value) => String(value)}
          marks={[
            { value: 1, label: "1" },
            { value: 2, label: "2" },
            { value: 3, label: "3" },
            { value: 4, label: "4" },
            { value: 5, label: "5" },
            { value: 6, label: "6" },
            { value: 7, label: "7" },
          ]}
          maxValue={7}
          minValue={1}
          step={1}
          title="Дней в неделю:"
          tooltip="Количество тренировочных дней в неделю."
          value={daysPerWeek}
          onChange={handleChange}
        />
      );
    }

    function SessionDurationSlider() {
      const [sessionDuration, setSessionDuration] = useState<number>(
        settings.sessionDurationMinutes || 60,
      );

      // Создаем debounced функцию для обновления настроек
      const debouncedUpdate = useCallback(
        debounce((duration: number) => {
          updateSettings({ sessionDurationMinutes: duration });
        }, 500),
        [updateSettings],
      );

      const handleChange = (value: number | number[]) => {
        const duration = value as number;

        setSessionDuration(duration);
        debouncedUpdate(duration);
      };

      return (
        <SliderWithMarks
          color="secondary"
          formatValue={formatDuration}
          marks={[
            { value: 30, label: "30м" },
            { value: 60, label: "1ч" },
            { value: 90, label: "1.5ч" },
            { value: 120, label: "2ч" },
          ]}
          maxValue={120}
          minValue={30}
          step={15}
          title="Продолжительность:"
          tooltip="Желаемая продолжительность тренировки."
          value={sessionDuration}
          onChange={handleChange}
        />
      );
    }

    function InjuriesField() {
      const [injuries, setInjuries] = useState<string>(settings.injuries || "");

      // Создаем debounced функцию для обновления настроек
      const debouncedUpdate = useCallback(
        debounce((injuriesText: string) => {
          updateSettings({ injuries: injuriesText });
        }, 1000),
        [updateSettings],
      );

      const handleChange = (value: string) => {
        setInjuries(value);
        debouncedUpdate(value);
      };

      return (
        <TextField
          maxRows={4}
          minRows={2}
          placeholder="Например: Болит спина, нельзя делать становую тягу. Травма колена - избегать приседаний."
          title="Травмы и ограничения:"
          tooltip="Опишите имеющиеся травмы или физические ограничения."
          value={injuries}
          onChange={handleChange}
        />
      );
    }

    function PriorityMuscleGroupsField() {
      const [priorityMuscleGroups, setPriorityMuscleGroups] = useState<
        string[]
      >(settings.priorityMuscleGroupsIds || []);

      const handleChange = (selectedIds: string[]) => {
        setPriorityMuscleGroups(selectedIds);
        updateSettings({ priorityMuscleGroupsIds: selectedIds });
      };

      const handleClear = () => {
        setPriorityMuscleGroups([]);
        updateSettings({ priorityMuscleGroupsIds: [] });
      };

      return (
        <MultiChipSelector
          color="success"
          options={muscleGroups.map((group) => ({
            id: group.id!,
            name: group.name!,
          }))}
          selectedIds={priorityMuscleGroups}
          title="Приоритетные группы мышц:"
          tooltip="Выберите группы мышц, на которые хотите сделать акцент."
          onChange={handleChange}
          onClear={handleClear}
        />
      );
    }

    function VarietyLevelSlider() {
      const [varietyLevel, setVarietyLevel] = useState<number>(
        settings.varietyLevel || 2,
      );

      // Создаем debounced функцию для обновления настроек
      const debouncedUpdate = useCallback(
        debounce((level: number) => {
          updateSettings({ varietyLevel: level });
        }, 500),
        [updateSettings],
      );

      const handleChange = (value: number | number[]) => {
        const level = value as number;

        setVarietyLevel(level);
        debouncedUpdate(level);
      };

      return (
        <SliderWithMarks
          color="secondary"
          formatValue={getVarietyLabel}
          marks={[
            { value: 1, label: "Мин" },
            { value: 2, label: "Сред" },
            { value: 3, label: "Макс" },
          ]}
          maxValue={3}
          minValue={1}
          step={1}
          title="Уровень разнообразия:"
          tooltip="Чем выше уровень, тем сильнее новые тренировки будут отличаться от предыдущих."
          value={varietyLevel}
          onChange={handleChange}
        />
      );
    }

    function BasePromptField() {
      const [basePrompt, setBasePrompt] = useState<string>(
        settings.basePrompt || "",
      );

      // Создаем debounced функцию для обновления настроек
      const debouncedUpdate = useCallback(
        debounce((promptText: string) => {
          updateSettings({ basePrompt: promptText });
        }, 1000),
        [updateSettings],
      );

      const handleChange = (value: string) => {
        setBasePrompt(value);
        debouncedUpdate(value);
      };

      return (
        <TextField
          maxRows={6}
          minRows={3}
          placeholder="Например: Моя цель — набрать мышечную массу. Нельзя делать упражнения с осевыми нагрузками на позвоночник."
          title="Дополнительные пожелания:"
          tooltip="Здесь вы можете указать свои предпочтения или противопоказания, которые будут учтены при генерации тренировок."
          value={basePrompt}
          onChange={handleChange}
        />
      );
    }

    return (
      <div className="flex flex-col gap-4 px-4 pb-4">
        <PrimaryGoalSelect />
        <ExperienceLevelSelect />
        <WorkoutPlanTypeSelect />
        <Divider />
        <DaysPerWeekSlider />
        <SessionDurationSlider />
        <Divider />
        <InjuriesField />
        <PriorityMuscleGroupsField />
        <Divider />
        <BasePromptField />
        <VarietyLevelSlider />
      </div>
    );
  }

  return (
    <div className="py-4 flex flex-col h-full gap-4">
      <PageHeader enableBackButton title="Настройки генерации" />
      <SettingsFrom />
    </div>
  );
}
