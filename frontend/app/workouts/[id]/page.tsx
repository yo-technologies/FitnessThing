"use client";

import { Button } from "@nextui-org/button";
import { Card, CardBody, CardHeader } from "@nextui-org/card";
import { Link } from "@nextui-org/link";
import {
  Modal,
  ModalContent,
  ModalFooter,
  ModalHeader,
  useDisclosure,
} from "@nextui-org/modal";
import { Accordion, AccordionItem } from "@nextui-org/accordion";
import { DropdownItem } from "@nextui-org/dropdown";
import { useRouter } from "next/navigation";
import { use, useEffect, useState } from "react";
import { toast } from "react-toastify";
import ReactMarkdown from "react-markdown";

import { BoltIcon, ChevronRightIcon, PlusIcon } from "@/config/icons";
import { ModalSelectExercise } from "@/components/pick-exercises-modal";
import { PageHeader } from "@/components/page-header";
import { Loading } from "@/components/loading";
import {
  WorkoutExerciseLogDetails,
  WorkoutGetWorkoutResponse,
  WorkoutWeightUnit,
} from "@/api/api.generated";
import { authApi } from "@/api/api";

function WorkoutTimer({ startTime }: { startTime: string | undefined }) {
  const [elapsedTime, setElapsedTime] = useState(() => {
    // Вычисляем начальное время сразу при создании компонента
    if (!startTime) return 0;
    const now = new Date().getTime();
    const start = new Date(startTime).getTime();

    return Math.floor((now - start) / 1000);
  });

  useEffect(() => {
    // Добавим отладочную информацию
    console.log("WorkoutTimer startTime:", startTime);

    if (!startTime) {
      console.warn("StartTime is not provided");

      return;
    }

    // Устанавливаем текущее время сразу
    const updateElapsedTime = () => {
      const now = new Date().getTime();
      const start = new Date(startTime).getTime();
      const elapsed = Math.floor((now - start) / 1000);

      setElapsedTime(elapsed);
    };

    // Обновляем время сразу
    updateElapsedTime();

    const interval = setInterval(updateElapsedTime, 1000);

    return () => clearInterval(interval);
  }, [startTime]);

  const formatTime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
    }

    return `${minutes}:${secs.toString().padStart(2, "0")}`;
  };

  // Добавим fallback если startTime не предоставлен
  if (!startTime) {
    return <p className="text-lg font-extralight">(--:--)</p>;
  }

  return (
    <p className="text-lg font-extralight">
      {"(" + formatTime(elapsedTime) + ")"}
    </p>
  );
}

function ExerciseLogCard({
  exerciseLogDetails,
  workoutId,
}: {
  exerciseLogDetails: WorkoutExerciseLogDetails;
  workoutId: string;
}) {
  return (
    <Card
      fullWidth
      as={Link}
      className="flex flex-row items-center justify-between p-3 cursor-pointer gap-3"
      href={`/workouts/${workoutId}/exerciseLogs/${exerciseLogDetails.exerciseLog?.id}`}
      shadow="sm"
    >
      <div className="flex flex-col items-start justify-between w-full gap-3">
        <CardHeader className="flex flex-col p-0 items-start">
          <div className="flex flex-row justify-between w-full gap-4">
            <p className="text-md font-bold text-start">
              {exerciseLogDetails.exercise?.name}
            </p>
            {exerciseLogDetails.exerciseLog?.powerRating! !== 0 && (
              <div className="flex flex-row items-center h-full">
                <BoltIcon className="w-3 h-3" />
                <p className="text-sm/6 font-semibold">
                  {exerciseLogDetails.exerciseLog?.powerRating}/10
                </p>
              </div>
            )}
          </div>
          {exerciseLogDetails.expectedSets!.length > 0 && (
            <div className="text-xs text-default-400">
              {exerciseLogDetails.expectedSets!.length} подходов x{" "}
              {(exerciseLogDetails.expectedSets!.reduce(
                (acc, set) => acc + set.reps!,
                0,
              )! /
                exerciseLogDetails.expectedSets!.length!) |
                0}{" "}
              раз
            </div>
          )}
        </CardHeader>
        {exerciseLogDetails.setLogs!.length > 0 && (
          <CardBody className="flex flex-col w-full gap-1 p-0">
            {exerciseLogDetails.setLogs?.map((setLog, setNum) => (
              <div
                key={setLog.id}
                className="grid grid-cols-[1.5rem_3rem_1rem_auto] gap-2 w-full"
              >
                <div className="text-sm font-semibold text-center w-4">
                  {setNum + 1}
                </div>
                <div className="text-sm font-semibold justify-self-start">
                  {setLog.weight! > 0
                    ? `${setLog.weight} ${exerciseLogDetails.exerciseLog?.weightUnit === WorkoutWeightUnit.WEIGHT_UNIT_LB ? "lb" : "кг"}`
                    : ""}
                </div>
                <div className="text-sm font-semibold justify-self-center">
                  x
                </div>
                <div className="text-sm font-semibold justify-self-start">
                  {setLog?.reps} раз
                </div>
              </div>
            ))}
          </CardBody>
        )}
      </div>
      <div className="flex flex-col items-center justify-between">
        <ChevronRightIcon className="w-4 h-4" fill="currentColor" />
      </div>
    </Card>
  );
}

export default function WorkoutDetailsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  const [workoutDetails, setWorkoutDetails] =
    useState<WorkoutGetWorkoutResponse>({});

  const { id } = use(params);

  const router = useRouter();

  const { isOpen, onOpen, onClose } = useDisclosure();

  async function fetchWorkoutDetails() {
    await authApi.v1
      .workoutServiceGetWorkout(id)
      .then((response) => {
        console.log("Workout details response:", response.data);
        console.log("Workout createdAt:", response.data?.workout?.createdAt);
        setWorkoutDetails(response.data!);
      })
      .catch((error) => {
        console.log(error);
        throw error;
      });
  }

  async function fetchData() {
    setIsLoading(true);
    try {
      await fetchWorkoutDetails();
    } catch (error) {
      console.log(error);
      toast.error("Failed to fetch workout details");
      setIsError(true);
    } finally {
      setIsLoading(false);
    }
  }

  async function finishWorkout() {
    try {
      await authApi.v1
        .workoutServiceCompleteWorkout(id, {})
        .then((response) => {
          console.log(response.data);
          router.push(`/workouts/${id}/results`);
        })
        .catch((error) => {
          console.log(error);
          throw error;
        });
    } catch (error) {
      console.log(error);
      toast.error("Не удалось завершить тренировку");
    } finally {
      setIsLoading(false);
    }
  }

  async function addExercisesToWorkout(exerciseIds: string[]) {
    try {
      for (const exerciseId of exerciseIds) {
        await authApi.v1
          .workoutServiceLogExercise(id, {
            exerciseId,
          })
          .then((response) => {
            console.log(response.data);
          })
          .catch((error) => {
            console.log(error);
            throw error;
          });
      }
      onClose();
      await fetchWorkoutDetails();
    } catch (error) {
      console.log(error);
      toast.error("Failed to add exercises to workout");
    } finally {
      setIsLoading(false);
    }
  }

  async function onDelete() {
    try {
      await authApi.v1
        .workoutServiceDeleteWorkout(id)
        .then((response) => {
          console.log(response.data);
          router.push("/");
        })
        .catch((error) => {
          console.log(error);
          throw error;
        });
    } catch (error) {
      console.log(error);
      toast.error("Не удалось удалить тренировку");
    }
  }

  const {
    isOpen: isFinishModalOpen,
    onOpen: onFinishModalOpen,
    onClose: onFinishModalClose,
  } = useDisclosure();

  function FinishWorkoutModal() {
    const [loading, setLoading] = useState(false);

    return (
      <Modal
        isOpen={isFinishModalOpen}
        placement="center"
        size="xs"
        title="Завершить тренировку"
        onClose={onFinishModalClose}
      >
        <ModalContent>
          {(onClose) => (
            <div className="flex flex-col py-4 gap-4">
              <ModalHeader className="p-0 px-4">
                Завершить тренировку?
              </ModalHeader>
              <ModalFooter className="flex flex-row gap-2 w-full px-2 py-0">
                <Button className="w-full" color="danger" onPress={onClose}>
                  Отмена
                </Button>
                <Button
                  className="w-full"
                  color="success"
                  isLoading={loading}
                  onPress={async () => {
                    setLoading(true);
                    await finishWorkout();
                    setLoading(false);
                  }}
                >
                  Завершить
                </Button>
              </ModalFooter>
            </div>
          )}
        </ModalContent>
        <p>Вы уверены, что хотите завершить тренировку?</p>
      </Modal>
    );
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

  return (
    <>
      <div className="py-4 flex flex-col h-full flex-grow max-w-full basis-full gap-4">
        <PageHeader
          enableBackButton={true}
          inner={
            <div className="flex flex-row items-end justify-start w-full h-full">
              <WorkoutTimer startTime={workoutDetails.workout!.createdAt} />
            </div>
          }
          title={"Тренировка"}
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
        <section className="flex flex-col gap-4 ">
          {workoutDetails.workout?.isAiGenerated && (
            <Accordion className="px-4" title="Сгенерировано ИИ">
              <AccordionItem
                className="flex flex-col p-0"
                classNames={{
                  trigger: "p-0",
                  content: "p-0 mt-2",
                  title: "text-xs font-semibold text-default-400",
                }}
                title="Записка от ИИ"
              >
                <ReactMarkdown className="text-xs/4 text-default-400">
                  {workoutDetails.workout?.reasoning}
                </ReactMarkdown>
              </AccordionItem>
            </Accordion>
          )}
          <div className="flex flex-col gap-4 px-4">
            {workoutDetails.exerciseLogs?.map((exerciseLogDetails, index) => (
              <ExerciseLogCard
                key={index}
                exerciseLogDetails={exerciseLogDetails}
                workoutId={id}
              />
            ))}
            <Card className="p-2">
              <Button
                className="w-full"
                onPress={() => {
                  onOpen();
                }}
              >
                <PlusIcon className="w-4 h-4" />
                <span>Добавить упражнение</span>
              </Button>
            </Card>
          </div>
        </section>
        <section className="w-full px-4">
          <Button
            className="w-full"
            color="primary"
            onPress={onFinishModalOpen}
          >
            Завершить тренировку
          </Button>
        </section>
      </div>
      <ModalSelectExercise
        excludeExerciseIds={workoutDetails.exerciseLogs!.map(
          (exerciseLog) => exerciseLog.exerciseLog!.exerciseId!,
        )}
        isOpen={isOpen}
        onClose={onClose}
        onSubmit={addExercisesToWorkout}
      />
      <FinishWorkoutModal />
    </>
  );
}
