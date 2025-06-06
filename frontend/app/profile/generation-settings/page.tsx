"use client";

import { useEffect, useRef, useState } from "react";
import {
  Card,
  CardBody,
  Popover,
  PopoverContent,
  PopoverTrigger,
  Slider,
  Textarea,
} from "@nextui-org/react";
import { toast } from "react-toastify";

import { PageHeader } from "@/components/page-header";
import { Loading } from "@/components/loading";
import { authApi } from "@/api/api";
import { WorkoutWorkoutGenerationSettings } from "@/api/api.generated";
import { CircleQuestionIcon } from "@/config/icons";

export default function RecordsPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  const [settings, setSettings] = useState<WorkoutWorkoutGenerationSettings>(
    {},
  );

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
      await fetchSettings();
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
    function VarietyLevelSlider() {
      const [varietyLevel, setVarietyLevel] = useState<number>(
        settings.varietyLevel!,
      );
      const timeoutRef = useRef<NodeJS.Timeout | null>(null);
      const valueRef = useRef(varietyLevel);

      useEffect(() => {
        valueRef.current = varietyLevel;
      }, [varietyLevel]);

      async function handleVarietyLevelChange(value: number) {
        try {
          await authApi.v1.userServiceUpdateWorkoutGenerationSettings({
            varietyLevel: value,
          });
        } catch (error) {
          console.log(error);
          toast.error("Не удалось обновить настройки");
        }
      }

      const handleChange = (value: number | number[]) => {
        value = value as number;

        setVarietyLevel(value);

        if (timeoutRef.current) {
          clearTimeout(timeoutRef.current);
        }

        timeoutRef.current = setTimeout(() => {
          handleVarietyLevelChange(valueRef.current);
        }, 300);
      };

      const popoverContent = (
        <PopoverContent>
          <p className="text-xs font-light text-default-600 p-1">
            Чем выше уровень, тем сильнее новые тренировки будут отличаться от
            предыдущих.
          </p>
        </PopoverContent>
      );

      return (
        <Card>
          <CardBody>
            <div className="flex flex-col gap-4">
              <div className="flex flex-row gap-2 items-center justify-between">
                <div className="flex flex-row gap-2 items-center">
                  <p>Уровень разнообразия:</p>
                  <p>{varietyLevel ?? "не известно"}</p>
                </div>
                <Popover
                  backdrop="opaque"
                  className="relative w-1/2 transform translate-x-[90%]"
                >
                  <PopoverTrigger>
                    <CircleQuestionIcon className="w-4 h-4 text-default-500" />
                  </PopoverTrigger>
                  {popoverContent}
                </Popover>
              </div>
              <Slider
                aria-label="Variety level"
                className="w-full"
                color="primary"
                maxValue={3}
                minValue={1}
                size="sm"
                step={1}
                value={varietyLevel}
                onChange={handleChange}
              />
            </div>
          </CardBody>
        </Card>
      );
    }

    function BasePromptField() {
      const [basePrompt, setBasePrompt] = useState<string>(
        settings.basePrompt || "",
      );

      const popoverContent = (
        <PopoverContent>
          <p className="text-xs font-light text-default-600 p-1">
            Здесь вы можете указать свои предпочтения или противопоказания,
            которые будут учтены при генерации тренировок.
          </p>
        </PopoverContent>
      );

      async function handleBasePromptChange() {
        try {
          await authApi.v1.userServiceUpdateWorkoutGenerationSettings({
            basePrompt,
          });
        } catch (error) {
          console.log(error);
          toast.error("Не удалось обновить настройки");
        }
      }

      return (
        <Card>
          <CardBody className="flex flex-col gap-4">
            <div className="flex flex-row gap-2 items-center justify-between">
              <p>Базовый запрос:</p>
              <Popover
                backdrop="opaque"
                className="relative w-1/2 transform translate-x-[90%]"
              >
                <PopoverTrigger>
                  <CircleQuestionIcon className="w-4 h-4 text-default-500" />
                </PopoverTrigger>
                {popoverContent}
              </Popover>
            </div>
            <Textarea
              fullWidth
              classNames={{
                input: "text-xs font-light placeholder:text-default-400",
              }}
              placeholder="Моя текущая цель — набрать мышечную массу. Мне нельзя делать упражнения с осевыми нагрузками."
              value={basePrompt}
              onBlur={handleBasePromptChange}
              onValueChange={(value) => setBasePrompt(value)}
            />
          </CardBody>
        </Card>
      );
    }

    return (
      <div className="flex flex-col gap-4 px-4">
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
