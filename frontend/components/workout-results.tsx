"use client";

import { Modal, ModalContent, useDisclosure } from "@nextui-org/modal";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { Divider } from "@nextui-org/divider";
import { Card } from "@nextui-org/card";
import { Button } from "@nextui-org/button";
import { Input, Textarea } from "@nextui-org/input";
import clsx from "clsx";
import { Form } from "@nextui-org/form";

import {
  HollowStarIcon,
  PersonIcon,
  StarIcon,
  WeightIcon,
} from "@/config/icons";
import {
  WorkoutGetWorkoutResponse,
  WorkoutWeightUnit,
} from "@/api/api.generated";
import { authApi } from "@/api/api";
import { Loading } from "@/components/loading";
import { translateMuscleGroups } from "@/config/muscle-groups";
import { convertWeight, weightUnitLabel } from "@/utils/units";

export function WorkoutResults({
  id,
  className,
}: {
  id: string;
  className?: string;
}) {
  const [workoutDetails, setWorkoutDetails] =
    useState<WorkoutGetWorkoutResponse>();

  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  const { isOpen, onOpen, onClose } = useDisclosure();

  async function fetchWorkoutDetails() {
    await authApi.v1
      .workoutServiceGetWorkout(id)
      .then((response) => {
        console.log(response.data);
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

  async function saveWorkoutAsRoutine(name: string) {
    console.log("Save workout as routine");
    console.log(name);
    await authApi.v1
      .routineServiceCreateRoutine({
        name: name,
        workoutId: id,
      })
      .then((response) => {
        console.log(response.data);
        toast.success("Шаблон сохранен");
      })
      .catch((error) => {
        console.log(error);
        toast.error("Ошибка сохранения шаблона");
      });
  }

  async function submitComment(comment: string) {
    console.log("Submit comment");
    console.log(comment);
    await authApi.v1
      .workoutServiceAddCommentToWorkout(id, { comment: comment })
      .then((response) => {
        console.log(response.data);
      })
      .catch((error) => {
        console.log(error);
        toast.error("Ошибка добавления комментария");
      });
  }

  async function submitRating(rating: number) {
    console.log("Submit rating");
    console.log(rating);
    await authApi.v1
      .workoutServiceRateWorkout(id, { rating: rating })
      .then((response) => {
        console.log(response.data);
      })
      .catch((error) => {
        console.log(error);
        toast.error("Ошибка добавления оценки");
      });
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

  function WorkoutResultsDate() {
    const startDate = new Date(workoutDetails?.workout?.createdAt!);
    const workoutDuration = new Date(
      new Date(workoutDetails?.workout?.finishedAt!).getTime() -
        startDate.getTime(),
    );

    return (
      <section className="flex flex-row justify-between items-center">
        <p className="text-md font-semibold">
          {startDate.toLocaleDateString("ru-RU", {
            day: "numeric",
            month: "long",
            hour: "numeric",
            minute: "numeric",
          })}
        </p>
        <p className="text-sm">
          {(() => {
            if (workoutDuration.getUTCHours() === 0) {
              return `${workoutDuration.getUTCMinutes()} минут`;
            }

            return `${workoutDuration.getUTCHours()} часа ${workoutDuration.getUTCMinutes()} минут`;
          })()}
        </p>
      </section>
    );
  }

  function WorkoutResultsExercises() {
    return (
      <section className="flex flex-col gap-2">
        <div className="flex flex-row items-center justify-between">
          <p className="text-sm font-bold">Упражнения</p>
          <p className="text-sm font-bold">Лучший сет</p>
        </div>
        <div className="flex flex-col">
          {workoutDetails?.exerciseLogs?.map((workoutExercise, index) => (
            <div
              key={index}
              className="flex flex-row justify-between w-full items-center gap-2"
            >
              {/* {num_of_sets} x {exercise_name} - {best_set} */}
              <p className="text-sm font-light w-fit">
                {workoutExercise?.setLogs?.length}
                {" x "}
                {workoutExercise?.exercise?.name}
              </p>
              <p className="text-sm font-light min-w-fit">
                {workoutExercise?.setLogs?.[0]?.weight || 0}{" "}
                {weightUnitLabel(workoutExercise?.exerciseLog?.weightUnit)} x{" "}
                {workoutExercise?.setLogs?.[0]?.reps || 0} раз
              </p>
            </div>
          ))}
        </div>
      </section>
    );
  }

  function WorkoutResultsAdditionalInfo() {
    const totalWeightKg = (workoutDetails?.exerciseLogs || []).reduce(
      (total, exerciseLog) => {
        const sum = (exerciseLog.setLogs || []).reduce((acc, setLog) => {
          const weightKg = convertWeight(
            setLog.weight || 0,
            exerciseLog.exerciseLog?.weightUnit,
            WorkoutWeightUnit.WEIGHT_UNIT_KG,
          );

          return acc + weightKg * (setLog.reps || 0);
        }, 0);

        return total + sum;
      },
      0,
    );

    const distinctMuscleGroups = [
      ...new Set(
        workoutDetails?.exerciseLogs?.reduce((muscleGroups, exerciseLog) => {
          return [
            ...muscleGroups,
            ...exerciseLog.exercise?.targetMuscleGroups!,
          ];
        }, [] as string[]),
      ),
    ];

    return (
      <section className="flex flex-col gap-2">
        {/*  список групп мышц */}
        <div className="flex flex-row gap-2 items-center">
          <div className="flex flex-row items-center justify-center w-3">
            <PersonIcon className="w-4 h-4" />
          </div>
          <p className="text-sm">
            {translateMuscleGroups(distinctMuscleGroups).join(", ")}
          </p>
        </div>
        {/* общий вес упражнений */}
        <div className="flex flex-row gap-2 items-center">
          <div className="flex flex-row items-center justify-center w-3">
            <WeightIcon className="w-3 h-3" />
          </div>
          <p className="text-sm">
            {Math.round((totalWeightKg || 0) * 10) / 10}
            {" кг"}
          </p>
        </div>
      </section>
    );
  }

  function WorkoutResultsCard() {
    return (
      <Card className="flex flex-col gap-4 w-full p-4">
        <WorkoutResultsDate />
        <Divider />
        <WorkoutResultsExercises />
        <Divider />
        <WorkoutResultsAdditionalInfo />
      </Card>
    );
  }

  function RateWorkoutCard() {
    const [comment, setComment] = useState("");
    const [rating, setRating] = useState(0);

    return (
      <Card className="flex flex-col gap-2 p-4">
        {/* starts and text area for coment and button to save comment*/}
        <p className="text-md font-semibold">Оценить тренировку</p>
        {/* stars */}
        <div className="flex flex-row gap-2 items-center ">
          <p className="text-sm font-bold">Оценка: </p>
          <div className="flex flex-row gap-2">
            {[1, 2, 3, 4, 5].map((btnRating) => (
              <Button
                key={btnRating}
                isIconOnly
                className={clsx(
                  "p-0 min-w-fit h-fit w-fit rounded-full bg-transparent",
                )}
                onPress={() => {
                  let newRating = btnRating === rating ? 0 : btnRating;

                  setRating(newRating);
                  submitRating(newRating);
                }}
              >
                {btnRating <= rating ? (
                  <StarIcon className="w-5 h-5" />
                ) : (
                  <HollowStarIcon className="w-5 h-5" />
                )}
              </Button>
            ))}
          </div>
        </div>
        {/* text area */}
        <Textarea
          placeholder="Оставьте комментарий"
          value={comment}
          onChange={(e) => setComment(e.target.value)}
        />
        <Button
          className="w-full"
          color="primary"
          size="sm"
          onPress={() => {
            submitComment(comment);
          }}
        >
          <span>Сохранить</span>
        </Button>
      </Card>
    );
  }

  function SaveWorkoutAsRoutineCard() {
    return (
      <Card className="flex flex-col gap-2 p-4">
        <p className="text-md font-semibold">
          Сохранить тренировку как шаблон?
        </p>
        <Button
          className="w-full"
          color="primary"
          size="sm"
          onPress={() => {
            console.log("Save workout as template");
            onOpen();
          }}
        >
          <span>Создать шаблон</span>
        </Button>
      </Card>
    );
  }

  function SaveWorkoutAsRoutineModal() {
    const [name, setName] = useState<string>();
    const [isLoading, setIsLoading] = useState(false);

    return (
      <Modal
        className="w-full"
        isOpen={isOpen}
        placement="center"
        size="xs"
        onClose={onClose}
      >
        <ModalContent>
          {(onClose) => (
            <div className="flex flex-col gap-4 p-4">
              <p className="text-md font-semibold">
                Сохранить тренировку как шаблон?
              </p>
              <Form
                className="flex flex-col gap-4"
                validationBehavior="native"
                onSubmit={(e) => {
                  e.preventDefault();
                  setIsLoading(true);
                  saveWorkoutAsRoutine(name!);
                  onClose();
                }}
              >
                <Input
                  autoFocus
                  isRequired
                  placeholder="Название шаблона"
                  validate={(value) => {
                    if (value.length < 3) {
                      return "Название должно быть длиннее 3 символов";
                    }
                  }}
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
                <Button
                  className="w-full"
                  color="primary"
                  isLoading={isLoading}
                  type="submit"
                >
                  <span>Сохранить</span>
                </Button>
              </Form>
            </div>
          )}
        </ModalContent>
      </Modal>
    );
  }

  return (
    <>
      <div className={clsx("flex flex-col gap-4", className)}>
        <WorkoutResultsCard />
        {(workoutDetails?.workout?.routineId === undefined ||
          workoutDetails?.workout?.routineId === "") && (
          <SaveWorkoutAsRoutineCard />
        )}
        <RateWorkoutCard />
      </div>
      <SaveWorkoutAsRoutineModal />
    </>
  );
}
