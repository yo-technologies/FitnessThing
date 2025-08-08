"use client";

import React, { useState, useEffect } from "react";
import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
} from "@nextui-org/modal";
import { Button } from "@nextui-org/button";
import { Progress } from "@nextui-org/progress";
import { Divider } from "@nextui-org/divider";
import { toast } from "react-toastify";
import { useAutoAnimate } from "@formkit/auto-animate/react";

import {
  ONBOARDING_STEPS,
  DEFAULT_FORM_DATA,
  GOAL_OPTIONS,
  EXPERIENCE_LEVEL_OPTIONS,
  WORKOUT_PLAN_TYPE_OPTIONS,
  formatDuration,
  getVarietyLabel,
  ChipSelector,
  SliderWithMarks,
  TextField,
  MultiChipSelector,
  EditableChipField,
  OnboardingFormData,
} from "./forms/generation-settings";

import { authApi } from "@/api/api";
import { WorkoutMuscleGroup } from "@/api/api.generated";
import { translateMuscleGroup } from "@/config/muscle-groups";

interface OnboardingModalProps {
  isOpen: boolean;
  onClose: () => void;
  onComplete: () => void;
}

export function OnboardingModal({
  isOpen,
  onClose,
  onComplete,
}: OnboardingModalProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formData, setFormData] =
    useState<OnboardingFormData>(DEFAULT_FORM_DATA);
  const [muscleGroups, setMuscleGroups] = useState<WorkoutMuscleGroup[]>([]);
  const [animateRef] = useAutoAnimate({
    duration: 300,
    easing: "ease-in-out",
  });
  const [titleAnimateRef] = useAutoAnimate({
    duration: 200,
    easing: "ease-in-out",
  });

  // Загрузка групп мышц при открытии модального окна
  useEffect(() => {
    if (isOpen) {
      fetchMuscleGroups();
    }
  }, [isOpen]);

  async function fetchMuscleGroups() {
    try {
      const response = await authApi.v1.exerciseServiceGetMuscleGroups();

      setMuscleGroups(response.data.muscleGroups || []);
    } catch (error) {
      console.log(error);
      toast.error("Не удалось загрузить группы мышц");
    }
  }

  const updateFormData = (updates: Partial<OnboardingFormData>) => {
    setFormData((prev) => ({ ...prev, ...updates }));
  };

  const handleNext = () => {
    if (currentStep < ONBOARDING_STEPS.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleSubmit = async () => {
    setIsSubmitting(true);
    try {
      await authApi.v1.userServiceUpdateWorkoutGenerationSettings(formData);
      toast.success("Добро пожаловать! Настройки сохранены.");
      onComplete();
      onClose();
    } catch (error) {
      console.log(error);
      toast.error("Не удалось сохранить настройки");
    } finally {
      setIsSubmitting(false);
    }
  };

  const progress = ((currentStep + 1) / ONBOARDING_STEPS.length) * 100;
  const currentStepData = ONBOARDING_STEPS[currentStep];

  const renderStepContent = () => {
    const content = (() => {
      switch (currentStepData.id) {
        case "welcome":
          return (
            <div className="flex flex-col gap-4 items-center justify-center">
              <div className="flex flex-col gap-6">
                <div className="flex justify-center items-center gap-4">
                  <div className="text-5xl">🏋️‍♂️</div>
                  <h2 className="text-2xl font-bold text-primary text-left">
                    Добро пожаловать в FitnessThing!
                  </h2>
                </div>
                <div className="flex flex-col gap-2">
                  <p className="text-md text-default-500">
                    Ответьте на несколько вопросов, чтобы наш ИИ мог создавать
                    индивидуальные тренировки специально для вас!
                  </p>
                </div>
              </div>
            </div>
          );

        case "basics":
          return (
            <div className="flex flex-col gap-4">
              <ChipSelector
                color="primary"
                options={GOAL_OPTIONS}
                title="Основная цель:"
                tooltip="Выберите основную цель ваших тренировок."
                value={formData.primaryGoal}
                onChange={(goal) => updateFormData({ primaryGoal: goal })}
              />
              <EditableChipField
                color="primary"
                placeholder="Например: Улучшить осанку, увеличить гибкость, развить координацию..."
                title="Дополнительные цели:"
                tooltip="Укажите дополнительные цели, которые важны для вас помимо основной."
                value={formData.secondaryGoals}
                onChange={(goals) => updateFormData({ secondaryGoals: goals })}
              />
              <ChipSelector
                color="success"
                options={EXPERIENCE_LEVEL_OPTIONS}
                title="Уровень опыта:"
                tooltip="Укажите ваш уровень опыта в тренировках."
                value={formData.experienceLevel}
                onChange={(level) => updateFormData({ experienceLevel: level })}
              />
              <ChipSelector
                color="secondary"
                options={WORKOUT_PLAN_TYPE_OPTIONS}
                title="Тип тренировочного плана:"
                tooltip="Выберите предпочитаемый тип тренировочного плана."
                value={formData.workoutPlanType}
                onChange={(planType) =>
                  updateFormData({ workoutPlanType: planType })
                }
              />
            </div>
          );

        case "schedule":
          return (
            <div className="flex flex-col gap-4">
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
                value={formData.daysPerWeek}
                onChange={(value) =>
                  updateFormData({ daysPerWeek: value as number })
                }
              />
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
                value={formData.sessionDurationMinutes}
                onChange={(value) =>
                  updateFormData({ sessionDurationMinutes: value as number })
                }
              />
            </div>
          );

        case "limitations":
          return (
            <div className="flex flex-col gap-4">
              <TextField
                maxRows={4}
                minRows={2}
                placeholder="Например: Болит спина, нельзя делать становую тягу. Травма колена - избегать приседаний."
                title="Травмы и ограничения:"
                tooltip="Опишите имеющиеся травмы или физические ограничения."
                value={formData.injuries}
                onChange={(value) => updateFormData({ injuries: value })}
              />
              <MultiChipSelector
                color="success"
                options={muscleGroups.map((group) => ({
                  id: group.id!,
                  name: translateMuscleGroup(group.name!),
                }))}
                selectedIds={formData.priorityMuscleGroupsIds}
                title="Приоритетные группы мышц:"
                tooltip="Выберите группы мышц, на которые хотите сделать акцент."
                onChange={(selectedIds) =>
                  updateFormData({ priorityMuscleGroupsIds: selectedIds })
                }
              />
            </div>
          );

        case "personalization":
          return (
            <div className="flex flex-col gap-4">
              <TextField
                maxRows={6}
                minRows={3}
                placeholder="Например: Я предпочитаю чередовать задействованые группы мышц"
                title="Дополнительные пожелания:"
                tooltip="Здесь вы можете указать дополнительные пожелания в свободной форме"
                value={formData.basePrompt}
                onChange={(value) => updateFormData({ basePrompt: value })}
              />
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
                value={formData.varietyLevel}
                onChange={(value) =>
                  updateFormData({ varietyLevel: value as number })
                }
              />
            </div>
          );

        default:
          return null;
      }
    })();

    return (
      <div key={currentStepData.id} className="w-full">
        {content}
      </div>
    );
  };

  return (
    <Modal
      isKeyboardDismissDisabled
      backdrop="blur"
      classNames={{
        base: "max-h-[95vh] mb-0 h-full",
        body: "max-h-[80vh] overflow-y-auto",
        closeButton: "absolute top-3 right-2",
      }}
      isDismissable={false}
      isOpen={isOpen}
      placement="bottom"
      scrollBehavior="inside"
      size="2xl"
      onOpenChange={onClose}
    >
      <ModalContent>
        {() => (
          <>
            <ModalHeader className="flex flex-col gap-2 bg-background p-4">
              <h2 className="text-lg font-bold">
                Настройка вашего фитнес-плана
              </h2>
              <div className="flex items-center justify-between gap-2">
                <Progress
                  className="w-full"
                  color="primary"
                  size="sm"
                  value={progress}
                />
                <span className="text-sm text-default-500 whitespace-nowrap">
                  {currentStep + 1} из {ONBOARDING_STEPS.length}
                </span>
              </div>
            </ModalHeader>

            <Divider />

            <ModalBody className="p-4 max-h-[75vh] overflow-y-auto bg-background">
              <div ref={titleAnimateRef}>
                {currentStep !== 0 && (
                  <div key={`title-${currentStep}`} className="flex flex-col">
                    <h3 className="text-lg font-medium">
                      {currentStepData.title}:{" "}
                    </h3>
                    <p className="text-xs text-default-400">
                      {currentStepData.description}
                    </p>
                  </div>
                )}
              </div>
              <div ref={animateRef} className="min-h-[400px] flex flex-col">
                {renderStepContent()}
              </div>
            </ModalBody>

            <Divider />

            <ModalFooter className="p-4 pb-8">
              <div className="flex justify-between w-full">
                {currentStep === 0 ? (
                  // Welcome страница - только кнопка "Начать"
                  <div className="flex justify-center w-full">
                    <Button color="primary" onPress={handleNext}>
                      Начать настройку
                    </Button>
                  </div>
                ) : (
                  // Остальные страницы - обычная навигация
                  <>
                    <Button
                      color="default"
                      isDisabled={currentStep === 0}
                      variant="bordered"
                      onPress={handlePrev}
                    >
                      Назад
                    </Button>

                    {currentStep === ONBOARDING_STEPS.length - 1 ? (
                      <Button
                        color="primary"
                        isLoading={isSubmitting}
                        onPress={handleSubmit}
                      >
                        Завершить
                      </Button>
                    ) : (
                      <Button color="primary" onPress={handleNext}>
                        Далее
                      </Button>
                    )}
                  </>
                )}
              </div>
            </ModalFooter>
          </>
        )}
      </ModalContent>
    </Modal>
  );
}
