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
import { DropdownItem } from "@nextui-org/dropdown";
import { useRouter, useSearchParams } from "next/navigation";
import { use, useEffect, useState } from "react";
import { useAutoAnimate } from "@formkit/auto-animate/react";
import { toast } from "react-toastify";

import {
  BoltIcon,
  ChatBubbleIcon,
  ChevronRightIcon,
  PlusIcon,
} from "@/config/icons";
import { ModalSelectExercise } from "@/components/pick-exercises-modal";
import { PageHeader } from "@/components/page-header";
import { Loading } from "@/components/loading";
import { WorkoutChatPanel } from "@/components/workout-chat";
import {
  WorkoutExerciseLogDetails,
  WorkoutGetWorkoutResponse,
} from "@/api/api.generated";
import { authApi } from "@/api/api";
import { weightUnitLabel } from "@/utils/units";

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
                    ? `${setLog.weight} ${weightUnitLabel(exerciseLogDetails.exerciseLog?.weightUnit)}`
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
  const searchParams = useSearchParams();

  const { isOpen, onOpen, onClose } = useDisclosure();

  const {
    isOpen: isChatOpen,
    onOpen: onChatOpen,
    onClose: onChatClose,
  } = useDisclosure();

  // Открываем чат автоматически и предзаполняем инпут, если переданы параметры
  const prefill = searchParams.get("prefill") ?? undefined;
  const shouldOpenChat = searchParams.get("openChat") === "1";

  useEffect(() => {
    if (shouldOpenChat) {
      onChatOpen();
    }
    // Открывать только при первом маунте
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // При закрытии панели чата обновляем данные тренировки (вдруг были изменения через инструменты)
  const handleChatClose = () => {
    onChatClose();
    // Лёгкий рефреш без изменения глобального isLoading, чтобы не мигал весь экран
    fetchWorkoutDetails().catch((e) => {
      console.warn("Failed to refresh workout after chat close", e);
    });
  };

  async function fetchWorkoutDetails() {
    try {
      const response = await authApi.v1.workoutServiceGetWorkout(id);

      setWorkoutDetails(response.data!);
    } catch (error) {
      console.log(error);
      throw error;
    }
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

  function WorkoutSections({
    workoutDetails,
    id,
    onOpen,
    onFinishModalOpen,
  }: {
    workoutDetails: WorkoutGetWorkoutResponse;
    id: string;
    onOpen: () => void;
    onFinishModalOpen: () => void;
  }) {
    const [exerciseListParent] = useAutoAnimate({ duration: 250 });

    return (
      <>
        <section className="flex flex-col gap-4 flex-grow">
          <div
            ref={exerciseListParent}
            className="flex flex-col gap-4 px-4 flex-grow"
          >
            {workoutDetails.exerciseLogs?.map((exerciseLogDetails, index) => (
              <ExerciseLogCard
                key={index}
                exerciseLogDetails={exerciseLogDetails}
                workoutId={id}
              />
            ))}
            <Card className="p-2">
              <Button
                fullWidth
                onPress={() => {
                  onOpen();
                }}
              >
                <PlusIcon className="w-4 h-4" />
                <span>Добавить упражнение</span>
              </Button>
            </Card>
          </div>
          <div className="px-4">
            <Button fullWidth color="primary" onPress={onFinishModalOpen}>
              Завершить тренировку
            </Button>
          </div>
        </section>
      </>
    );
  }

  function MainContent() {
    const [switchParent] = useAutoAnimate({
      duration: 220,
      easing: "ease-in-out",
    });

    return (
      <>
        <div className="py-4 flex flex-col h-full flex-grow max-w-full basis-full gap-4">
          <PageHeader
            enableBackButton
            inner={
              <div className="flex h-full w-full items-center justify-start">
                <WorkoutTimer startTime={workoutDetails.workout?.createdAt} />
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
          <div
            ref={switchParent}
            className="min-h-[160px] flex flex-col flex-grow"
          >
            <WorkoutSections
              id={id}
              workoutDetails={workoutDetails}
              onFinishModalOpen={onFinishModalOpen}
              onOpen={onOpen}
            />
          </div>
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
        <Button
          isIconOnly
          aria-label="Открыть чат с тренером"
          className="fixed bottom-20 right-6 z-40 shadow-lg w-14 h-14"
          color="secondary"
          radius="full"
          onPress={onChatOpen}
        >
          <ChatBubbleIcon className="h-7 w-7" />
        </Button>
      </>
    );
  }

  return (
    <>
      <MainContent />
      <WorkoutChatPanel
        isOpen={isChatOpen}
        prefill={prefill}
        workoutId={id}
        onClose={handleChatClose}
      />
    </>
  );
}
