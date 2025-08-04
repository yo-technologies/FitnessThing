"use client";

import { useEffect, useState, useCallback } from "react";
import { Divider } from "@nextui-org/react";
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
import {
  ChipSelector,
  SliderWithMarks,
  TextField,
  MultiChipSelector,
  GOAL_OPTIONS,
  EXPERIENCE_LEVEL_OPTIONS,
  WORKOUT_PLAN_TYPE_OPTIONS,
  formatDuration,
  getVarietyLabel,
} from "@/components/forms/generation-settings";

export default function GenerationSettingsPage() {
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
