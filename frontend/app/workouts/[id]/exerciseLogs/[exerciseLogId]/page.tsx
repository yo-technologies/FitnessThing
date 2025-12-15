"use client";

import { Button } from "@heroui/button";
import { Card, CardBody } from "@heroui/card";
import { Divider } from "@heroui/divider";
import {
  Dropdown,
  DropdownItem,
  DropdownMenu,
  DropdownTrigger,
} from "@heroui/dropdown";
import { Form } from "@heroui/form";
import { Textarea } from "@heroui/input";
import {
  Modal,
  ModalBody,
  ModalContent,
  ModalFooter,
  ModalHeader,
  useDisclosure,
} from "@heroui/modal";
import { Tabs, Tab } from "@heroui/tabs";
import { Slider } from "@heroui/slider";
import { Tooltip } from "@heroui/tooltip";
import { useRouter } from "next/navigation";
import { use, useEffect, useRef, useState } from "react";
import { addToast } from "@heroui/toast";
import { Popover, PopoverContent, PopoverTrigger, Spinner } from "@heroui/react";

import { PageHeader } from "@/components/page-header";
import { BoltIcon, CircleQuestionIcon, TrashCanIcon } from "@/config/icons";
import { Loading } from "@/components/loading";
import {
  WorkoutExerciseLogDetails,
  WorkoutExpectedSet,
  WorkoutSetLog,
  WorkoutWeightUnit,
} from "@/api/api.generated";
import { authApi } from "@/api/api";
import { InputWithIncrement } from "@/components/input-with-increments";
import InfiniteScroll from "@/components/infinite-scroll";
import { unitFromKey, unitKey, weightUnitLabel } from "@/utils/units";

export default function RoutineDetailsPage({
  params,
}: {
  params: Promise<{ id: string; exerciseLogId: string }>;
}) {
  const { exerciseLogId, id } = use(params);

  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  const [exerciseLogDetails, setExerciseLogDetails] =
    useState<WorkoutExerciseLogDetails>({});

  const currentUnit: WorkoutWeightUnit =
    exerciseLogDetails.exerciseLog?.weightUnit ||
    WorkoutWeightUnit.WEIGHT_UNIT_KG;
  const unitLabel = weightUnitLabel(currentUnit);

  const [exerciseLogHistory, setExerciseLogHistory] = useState<
    WorkoutExerciseLogDetails[]
  >([]);
  const [exerciseLogForUpdate, setExerciseLogForUpdate] =
    useState<WorkoutSetLog>({});

  const limit = 10;
  const [offset, setOffset] = useState<number>(0);
  const [hasMore, setHasMore] = useState<boolean>(true);

  const { isOpen, onOpen, onClose } = useDisclosure();

  const router = useRouter();

  async function fetchExerciseLogDetails() {
    await authApi.v1
      .workoutServiceGetExerciseLogDetails(id, exerciseLogId)
      .then((response) => {
        console.log(response.data);
        setExerciseLogDetails(response.data.exerciseLogDetails!);
      })
      .catch((error) => {
        console.log(error);
        throw error;
      });
  }

  async function fetchMore() {
    try {
      await authApi.v1
        .exerciseServiceGetExerciseHistory(exerciseLogDetails.exercise?.id!, {
          offset: offset,
          limit: limit,
        })
        .then((response) => {
          console.log(response.data);
          setExerciseLogHistory((prev) => [
            ...prev,
            ...response.data.exerciseLogs!,
          ]);
          setHasMore(response.data.exerciseLogs!.length === limit);
          setOffset(offset + limit);
        });
    } catch (error) {
      console.log(error);
      addToast({ title: "Failed to fetch more exercise logs", color: "danger" });
    }
  }

  async function fetchData() {
    setIsLoading(true);
    try {
      await fetchExerciseLogDetails();
    } catch (error) {
      console.log(error);
      addToast({ title: "Failed to fetch workout details", color: "danger" });
      setIsError(true);
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    fetchData();

    // Слушаем событие обновления данных из layout (когда чат применяет инструменты)
    const handleDataUpdate = () => {
      fetchExerciseLogDetails().catch((e) => {
        console.warn("Failed to refresh exercise log after data update", e);
      });
    };

    window.addEventListener("workout-data-updated", handleDataUpdate);

    return () => {
      window.removeEventListener("workout-data-updated", handleDataUpdate);
    };
  }, []);

  useEffect(() => {
    if (exerciseLogDetails.exercise?.id) {
      fetchMore();
    }
  }, [exerciseLogDetails]);

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

  // Обработчик смены единиц измерения вынесен отдельно от разметки
  async function handleWeightUnitSelectionChange(keys: any) {
    const key = Array.from(keys)[0] as string;
    const nextUnit = unitFromKey(key);

    if (nextUnit === currentUnit) return;
    const prevUnit = currentUnit;

    // оптимистичное обновление локального состояния (моментально меняем юниты в UI)
    setExerciseLogDetails((prev) => ({
      ...prev,
      exerciseLog: {
        ...prev.exerciseLog,
        weightUnit: nextUnit,
      },
    }));

    try {
      const res = await authApi.v1.workoutServiceUpdateExerciseLogWeightUnit(
        id,
        exerciseLogId,
        { weightUnit: nextUnit },
      );

      // сохраняем данные из ответа (пересчитанные веса сетов)
      const { exerciseLogDetails: updated } = res.data;

      if (updated) {
        setExerciseLogDetails(updated);
      } else {
        setExerciseLogDetails((prev) => ({
          ...prev,
          exerciseLog: (res.data as any)?.exerciseLog ?? prev.exerciseLog,
          setLogs: (res.data as any)?.setLogs ?? prev.setLogs,
          expectedSets: (res.data as any)?.expectedSets ?? prev.expectedSets,
        }));
      }
    } catch (error) {
      console.log(error);
      // откатываем оптимистичное изменение
      setExerciseLogDetails((prev) => ({
        ...prev,
        exerciseLog: {
          ...prev.exerciseLog,
          weightUnit: prevUnit,
        },
      }));
      addToast({ title: "Не удалось сменить единицы измерения", color: "danger" });
    }
  }

  function WeightUnitSelectorLabel() {
    return (
      <Dropdown>
        <DropdownTrigger>
          <p className="cursor-pointer text-md font-light">Вес {unitLabel} ▾</p>
        </DropdownTrigger>
        <DropdownMenu
          disallowEmptySelection
          aria-label="Выбор единиц"
          selectedKeys={new Set([unitKey(currentUnit)])}
          selectionMode="single"
          onSelectionChange={handleWeightUnitSelectionChange}
        >
          <DropdownItem key="kg">кг</DropdownItem>
          <DropdownItem key="lb">lb</DropdownItem>
        </DropdownMenu>
      </Dropdown>
    );
  }
  function SetLogCard({
    setLog,
    setNum,
    unitLabel,
    enableDelete,
    isDisabled,
    onDelete,
    onPress,
  }: {
    setLog: WorkoutSetLog | WorkoutExpectedSet;
    setNum: number;
    unitLabel: string;
    enableDelete?: boolean;
    isDisabled?: boolean;
    onDelete?: () => Promise<void>;
    onPress?: () => void;
  }) {
    const [isLoading, setIsLoading] = useState(false);

    return (
      <Card
        className="flex flex-row items-center justify-between p-2 w-full h-[2.5rem]"
        isDisabled={isDisabled}
      >
        <button onClick={onPress}>
          <div className="grid grid-cols-[1.5rem_3rem_1rem_auto] gap-2 px-2 w-full">
            <div className="text-sm font-semibold text-center w-4">
              {setNum + 1}
            </div>
            <div className="text-sm font-semibold justify-self-start">
              {setLog.weight! > 0 ? `${setLog.weight} ${unitLabel}` : ""}
            </div>
            <div className="text-sm font-semibold justify-self-center">x</div>
            <div className="text-sm font-semibold justify-self-start">
              {setLog?.reps} раз
            </div>
          </div>
        </button>
        {enableDelete && (
          <div className="flex flex-col">
            <Button
              isIconOnly
              className="h-fit w-fit min-w-fit p-2"
              color="danger"
              isLoading={isLoading}
              size="sm"
              spinner={
                <Spinner
                  classNames={{ wrapper: "w-3 h-3" }}
                  color="white"
                  size="sm"
                />
              }
              onPress={async () => {
                setIsLoading(true);
                await onDelete!();
                setIsLoading(false);
              }}
            >
              <TrashCanIcon className="w-3 h-3" />
            </Button>
          </div>
        )}
      </Card>
    );
  }

  async function onDelete() {
    try {
      await authApi.v1.workoutServiceDeleteExerciseLog(id, exerciseLogId);
      router.back();
    } catch (error) {
      console.log(error);
      addToast({ title: "Failed to delete exercise log", color: "danger" });
    }
  }

  function UpdateSetLogModal({
    isOpen,
    onClose,
    setLog,
  }: {
    isOpen: boolean;
    onClose: () => void;
    setLog: WorkoutSetLog;
  }) {
    const [weight, setWeight] = useState<number>(setLog.weight!);
    const [reps, setReps] = useState<number>(setLog.reps!);

    async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
      event.preventDefault();
      if (weight < 0) {
        addToast({ title: "Вес не может быть отрицательным", color: "danger" });

        return;
      }

      if (reps <= 0) {
        addToast({ title: "Повторы должны быть больше 0", color: "danger" });

        return;
      }

      try {
        await authApi.v1.workoutServiceUpdateSetLog(
          id,
          exerciseLogId,
          setLog.id!,
          {
            weight: weight!,
            reps: reps!,
          },
        );
        fetchExerciseLogDetails();
        onClose();
      } catch (error) {
        console.log(error);
        addToast({ title: "Ошибка при изменении сета", color: "danger" });
      } finally {
        setIsLoading(false);
      }
    }

    return (
      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalContent>
          {(onClose) => (
            <div className="flex flex-col py-4 mb-4">
              <ModalHeader className="p-0 px-4">Изменить сет</ModalHeader>
              <Form
                className="flex flex-col p-0 px-2"
                validationBehavior="native"
                onSubmit={handleSubmit}
              >
                <ModalBody className="flex flex-row gap-2 px-2 w-full">
                  <div className="flex flex-col gap-1 w-1/2">
                    <InputWithIncrement
                      className="h-10"
                      classNames={{ incrementButton: "w-12" }}
                      label={`Вес ${unitLabel}`}
                      min={0}
                      placeholder="10"
                      setValue={setWeight}
                      value={weight}
                    />
                  </div>
                  <div className="flex flex-col gap-1 w-1/2">
                    <InputWithIncrement
                      allowFloat={false}
                      className="h-10"
                      classNames={{ incrementButton: "w-12" }}
                      label="Повторы"
                      min={0}
                      placeholder="10"
                      setValue={setReps}
                      step={1}
                      value={reps}
                    />
                  </div>
                </ModalBody>
                <ModalFooter className="flex flex-col gap-2 w-full justify-around px-2 py-0">
                  <Button
                    className="w-full"
                    color="success"
                    size="sm"
                    type="submit"
                  >
                    Изменить
                  </Button>
                  <Button
                    className="w-full"
                    color="danger"
                    size="sm"
                    onPress={onClose}
                  >
                    Отмена
                  </Button>
                </ModalFooter>
              </Form>
            </div>
          )}
        </ModalContent>
      </Modal>
    );
  }

  function TodayContent() {
    function SetLogsCard() {
      const [weight, setWeight] = useState<number>(0);
      const [reps, setReps] = useState<number>(0);
      const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

      useEffect(() => {
        const setIndex = exerciseLogDetails.setLogs?.length ?? 0;

        // 1. Ожидаемый вес из expectedSets
        const expectedSet = exerciseLogDetails.expectedSets?.[setIndex];

        if (expectedSet && expectedSet.weight && expectedSet.weight > 0) {
          setWeight(expectedSet.weight);
        }
        // 2. Если нет ожидаемого, берём вес последнего сета текущей тренировки
        else if (exerciseLogDetails.setLogs?.length) {
          const lastIndex = exerciseLogDetails.setLogs.length - 1;

          setWeight(exerciseLogDetails.setLogs[lastIndex]?.weight!);
        }
        // 3. Если нет текущих сетов, берём из истории
        else if (
          exerciseLogHistory.length &&
          exerciseLogHistory[0].setLogs?.length
        ) {
          const lastIndex = exerciseLogHistory[0]!.setLogs!.length - 1;

          setWeight(exerciseLogHistory[0]!.setLogs![lastIndex]?.weight!);
        }

        // Повторы: 1. Ожидаемые повторы из expectedSets
        if (expectedSet && expectedSet.reps && expectedSet.reps > 0) {
          setReps(expectedSet.reps);
        }
        // 2. Если нет ожидаемых, берём из последнего сета текущей тренировки
        else if (exerciseLogDetails.setLogs?.length) {
          const lastIndex = exerciseLogDetails.setLogs.length - 1;

          setReps(exerciseLogDetails.setLogs[lastIndex]?.reps!);
        }
        // 3. Если нет текущих сетов, берём из истории
        else if (
          exerciseLogHistory.length &&
          exerciseLogHistory[0].setLogs?.length
        ) {
          const lastIndex = exerciseLogHistory[0]!.setLogs!.length - 1;

          setReps(exerciseLogHistory[0]!.setLogs![lastIndex]?.reps!);
        }
      }, [exerciseLogDetails, exerciseLogHistory]);

      async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
        event.preventDefault();
        if (isSubmitting) return;
        if (weight < 0) {
          addToast({ title: "Вес не может быть отрицательным", color: "danger" });

          return;
        }

        if (reps <= 0) {
          addToast({ title: "Повторы должны быть больше 0", color: "danger" });

          return;
        }

        try {
          setIsSubmitting(true);
          await authApi.v1.workoutServiceLogSet(id, exerciseLogId, {
            weight: weight!,
            reps: reps!,
          });
          await fetchExerciseLogDetails();
        } catch (error) {
          console.log(error);
          addToast({ title: "Ошибка при добавлении сета", color: "danger" });
        } finally {
          setIsSubmitting(false);
        }
      }

      async function onDeleteSet(setId: string) {
        try {
          await authApi.v1.workoutServiceDeleteSetLog(id, exerciseLogId, setId);
          await fetchExerciseLogDetails();
        } catch (error) {
          console.log(error);
          addToast({ title: "Failed to delete set", color: "danger" });
        }
      }

      return (
        <Card>
          <CardBody>
            <div className="flex flex-col gap-4">
              <Form
                className="flex flex-col gap-4"
                validationBehavior="native"
                onSubmit={handleSubmit}
              >
                <div className="flex flex-row justify-around gap-4">
                  <div className="flex flex-col gap-1 w-1/2 h-16">
                    <InputWithIncrement
                      className="h-10"
                      classNames={{ incrementButton: "w-12" }}
                      labelNode={<WeightUnitSelectorLabel />}
                      min={0}
                      placeholder="10"
                      setValue={setWeight}
                      size="md"
                      value={weight}
                    />
                  </div>
                  <div className="flex flex-col gap-1 w-1/2">
                    <InputWithIncrement
                      allowFloat={false}
                      className="h-10"
                      classNames={{ incrementButton: "w-12" }}
                      label="Повторы"
                      min={0}
                      placeholder="10"
                      setValue={setReps}
                      step={1}
                      value={reps}
                    />
                  </div>
                </div>
                <Button
                  className="w-full"
                  color="primary"
                  isDisabled={isSubmitting}
                  isLoading={isSubmitting}
                  size="sm"
                  spinner={
                    <Spinner
                      classNames={{ wrapper: "w-3 h-3" }}
                      color="white"
                      size="sm"
                    />
                  }
                  type="submit"
                >
                  Добавить
                </Button>
              </Form>
              <Divider />
              <div className="flex flex-col gap-2">
                <>
                  {exerciseLogDetails.setLogs?.map((setLog, index) => (
                    <SetLogCard
                      key={index}
                      enableDelete
                      setLog={setLog}
                      setNum={index}
                      unitLabel={unitLabel}
                      onDelete={() => onDeleteSet(setLog.id!)}
                      onPress={() => {
                        setExerciseLogForUpdate(setLog);
                        onOpen();
                      }}
                    />
                  ))}
                  {exerciseLogDetails.expectedSets?.length! >
                    exerciseLogDetails.setLogs?.length! &&
                    exerciseLogDetails.expectedSets
                      ?.slice(exerciseLogDetails.setLogs?.length!)
                      .map((set, index) => (
                        <SetLogCard
                          key={index}
                          isDisabled
                          setLog={set}
                          setNum={index + exerciseLogDetails.setLogs?.length!}
                          unitLabel={unitLabel}
                        />
                      ))}
                </>
              </div>
            </div>
          </CardBody>
        </Card>
      );
    }

    function PowerRatingCard() {
      const [powerRating, setPowerRating] = useState<number>(
        exerciseLogDetails.exerciseLog?.powerRating!,
      );
      const timeoutRef = useRef<NodeJS.Timeout>();
      const powerRatingRef = useRef(powerRating);

      useEffect(() => {
        powerRatingRef.current = powerRating;
      }, [powerRating]);

      async function handlePowerRatingChange() {
        try {
          await authApi.v1.workoutServiceAddPowerRatingToExerciseLog(
            id,
            exerciseLogId,
            { powerRating: powerRatingRef.current },
          );
        } catch (error) {
          console.log(error);
          addToast({ title: "Failed to update power rating", color: "danger" });
        }
      }

      useEffect(() => {
        return () => {
          if (timeoutRef.current) clearTimeout(timeoutRef.current);
        };
      }, []);

      const handleChange = (value: number | number[]) => {
        value = value as number;

        setPowerRating(value);

        if (timeoutRef.current) clearTimeout(timeoutRef.current);

        timeoutRef.current = setTimeout(() => {
          handlePowerRatingChange();
        }, 500);
      };

      return (
        <Card>
          <CardBody>
            <div className="flex flex-col gap-4">
              <div className="flex flex-row justify-between items-center">
                <div className="flex flex-row gap-2 items-center">
                  <p>
                    Оценка усилия:
                  </p>
                  <p>{powerRating}</p>
                </div>
                <Popover
                  backdrop="opaque" 
                  size="sm" 
                  placement="top"
                  classNames={{base:"flex flex-row justify-end items-end px-2"}}
                >
                  <PopoverTrigger>
                    <CircleQuestionIcon className="w-4 h-4 text-default-500 cursor-pointer" />
                  </PopoverTrigger>
                  <PopoverContent className="w-[80%] items-start">
                    <p className="text-xs font-light text-default-600 mb-1">
                      Оцените нагрузку (Power Rating) от 0 до 10
                    </p>
                    <p className="text-xs font-light text-default-600">
                      6–7 — умеренно, можно увеличить нагрузку.
                    </p>
                    <p className="text-xs font-light text-default-600">
                      8–9 — оптимум, повышать не нужно.
                    </p>
                    <p className="text-xs font-light text-default-600 mb-1">
                      9–10 — слишком тяжело, прогрессии быть не должно.
                    </p>
                    <p className="text-xs font-light text-default-600">
                      Цель — удерживать нагрузку в зоне 8–9.
                    </p>
                  </PopoverContent>
                </Popover>
              </div>
              <Slider
                aria-label="Power rating"
                className="w-full"
                classNames={{
                  track: "bg-[linear-gradient(to_right,#3b82f6_0%,#3b82f6_60%,#eab308_60%,#eab308_80%,#22c55e_80%,#22c55e_90%,#ef4444_90%,#ef4444_100%)] border-none",
                  filler: "bg-transparent",
                }}
                marks={[
                  { value: 0, label: "0" },
                  { value: 6, label: "6" },
                  { value: 8, label: "8" },
                  { value: 9, label: "9" },
                  { value: 10, label: "10" },
                ]}
                maxValue={10}
                minValue={0}
                size="sm"
                step={1}
                value={powerRating}
                onChange={handleChange}
              />
            </div>
          </CardBody>
        </Card>
      );
    }

    function NotesCard() {
      const [notes, setNotes] = useState<string>(
        exerciseLogDetails.exerciseLog?.notes!,
      );

      async function submitNotesChange() {
        try {
          await authApi.v1.workoutServiceAddNotesToExerciseLog(
            id,
            exerciseLogId,
            {
              notes: notes,
            },
          );
        } catch (error) {
          console.log(error);
          addToast({ title: "Failed to update exercise log", color: "danger" });
        }
      }

      return (
        <Card>
          <CardBody className="flex flex-col gap-4">
            <p>Заметки:</p>
            <Textarea
              className="w-full"
              placeholder="Заметки"
              value={notes}
              onValueChange={(value) => setNotes(value)}
            />
            <Button
              className="w-full"
              color="primary"
              size="sm"
              onPress={submitNotesChange}
            >
              Сохранить
            </Button>
          </CardBody>
        </Card>
      );
    }

    return (
      <div className="flex flex-col gap-4">
        <SetLogsCard />
        <PowerRatingCard />
        <NotesCard />
      </div>
    );
  }

  function HistoryContent() {
    function ExerciseHistoryCard({
      exerciseLog,
    }: {
      exerciseLog: WorkoutExerciseLogDetails;
    }) {
      const historyUnitLabel = weightUnitLabel(
        exerciseLog.exerciseLog?.weightUnit || WorkoutWeightUnit.WEIGHT_UNIT_KG,
      );

      function formatDate(date: string) {
        const dateObj = new Date(date);

        let formatOpts = {};

        if (Date.now() - dateObj.getTime() < 60 * 60 * 24 * 14 * 1000) {
          formatOpts = {
            weekday: "long",
            day: "numeric",
            month: "long",
          };
        } else {
          formatOpts = {
            day: "numeric",
            month: "long",
            year: "numeric",
          };
        }

        return dateObj.toLocaleDateString("ru-RU", formatOpts);
      }

      return (
        <Card>
          <CardBody className="flex flex-col gap-2">
            <div className="flex flex-row justify-between items-center">
              <p className="text-md font-bold">
                {formatDate(exerciseLog.exerciseLog?.createdAt!)}
              </p>

              {exerciseLog.exerciseLog?.powerRating! !== 0 && (
                <div className="flex flex-row items-center">
                  <BoltIcon className="w-3 h-3" />
                  <p className="text-sm font-semibold">
                    {exerciseLog.exerciseLog?.powerRating}/10
                  </p>
                </div>
              )}
            </div>
            <div className="flex flex-col gap-1">
              {exerciseLog.setLogs?.map((setLog, index) => (
                <SetLogCard
                  key={index}
                  setLog={setLog}
                  setNum={index}
                  unitLabel={historyUnitLabel}
                />
              ))}
            </div>
            {exerciseLog.exerciseLog?.notes && (
              <Textarea
                isReadOnly
                className="w-full"
                classNames={{
                  input: "text-xs font-light",
                }}
                maxRows={4}
                value={exerciseLog.exerciseLog?.notes}
              />
            )}
          </CardBody>
        </Card>
      );
    }

    return (
      <InfiniteScroll
        className="flex flex-col gap-4"
        fetchMore={fetchMore}
        hasMore={hasMore}
      >
        {exerciseLogHistory.map(
          (exerciseLog, index) =>
            exerciseLog.setLogs!.length > 0 &&
            exerciseLog.exerciseLog?.workoutId != id && (
              <ExerciseHistoryCard key={index} exerciseLog={exerciseLog} />
            ),
        )}
        {exerciseLogHistory.length === 0 && (
          <div className="flex flex-col gap-2">
            <p>Пока тут пусто(</p>
            <p className="text-sm font-light">
              Сначала потренируйся, а потом мы покажем историю
            </p>
          </div>
        )}
      </InfiniteScroll>
    );
  }

  return (
    <>
      <div className="py-4 flex-grow max-w-full">
        <div className="h-full max-h-full overflow-y-auto gap-4 flex flex-col">
          <PageHeader
            enableBackButton
            title={exerciseLogDetails.exercise?.name!}
          >
            <DropdownItem
              key="delete"
              className="text-danger"
              color="danger"
              onPress={onDelete}
            >
              Удалить
            </DropdownItem>
          </PageHeader>
          <section className="flex flex-col flex-grow gap-4 px-4 justify-start">
            <Tabs aria-label="Options" classNames={{ panel: "p-0" }}>
              <Tab key="today" title="Сегодня">
                <TodayContent />
              </Tab>
              <Tab key="history" title="История">
                <HistoryContent />
              </Tab>
            </Tabs>
          </section>
        </div>
      </div>
      <UpdateSetLogModal
        isOpen={isOpen}
        setLog={exerciseLogForUpdate}
        onClose={onClose}
      />
    </>
  );
}
