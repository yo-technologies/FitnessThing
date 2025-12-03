"use client";

import { useRouter } from "next/navigation";
import { use, useEffect, useRef, useState } from "react";
import { toast } from "react-toastify";
import { Button } from "@heroui/button";
import { Card, CardBody } from "@heroui/card";
import { DropdownItem } from "@heroui/dropdown";
import { Spinner } from "@heroui/spinner";

import { Loading } from "@/components/loading";
import { authApi } from "@/api/api";
import {
  WorkoutExerciseInstanceDetails,
  WorkoutSet,
  WorkoutSetType,
} from "@/api/api.generated";
import { PageHeader } from "@/components/page-header";
import { PlusIcon, RepeatIcon, TrashCanIcon } from "@/config/icons";
import { InputWithIncrement } from "@/components/input-with-increments";

export default function ExerciseInstancePage({
  params,
}: {
  params: Promise<{ id: string; exerciseInstanceId: string }>;
}) {
  const { exerciseInstanceId, id } = use(params);

  const [exerciseInstanceDetails, setExerciseInstanceDetails] =
    useState<WorkoutExerciseInstanceDetails>({});

  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  const router = useRouter();

  async function fetchExerciseInstanceDetails() {
    await authApi.v1
      .routineServiceGetExerciseInstanceDetails(id, exerciseInstanceId)
      .then((response) => {
        console.log(response.data);
        setExerciseInstanceDetails(response.data.exerciseInstanceDetails!);
      })
      .catch((error) => {
        console.log(error);
        throw error;
      });
  }

  async function fetchData() {
    setIsLoading(true);
    try {
      await fetchExerciseInstanceDetails();
      setIsError(false);
    } catch (error) {
      console.log(error);
      toast.error("Failed to fetch workout details");
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

  function SetCard({ num, set }: { num: number; set: WorkoutSet }) {
    const [isLoading, setIsLoading] = useState(false);
    const [value, setValue] = useState(set.reps!);
    const timeoutRef = useRef<NodeJS.Timeout>();
    const valueRef = useRef(value); // Храним актуальное значение

    useEffect(() => {
      valueRef.current = value;
    }, [value]);

    useEffect(() => {
      return () => {
        if (timeoutRef.current) clearTimeout(timeoutRef.current);
      };
    }, []);

    async function updateSet() {
      try {
        await authApi.v1
          .routineServiceUpdateSetInExerciseInstance(
            id,
            exerciseInstanceId,
            set.id!,
            {
              reps: valueRef.current, // Используем актуальное значение из ref
              setType: WorkoutSetType.SET_TYPE_REPS,
            },
          )
          .then((response) => {
            console.log(response.data);
            fetchExerciseInstanceDetails();
          });
      } catch (error) {
        console.log(error);
        toast.error("Не удалось обновить подход");
      }
    }

    async function onDelete() {
      setIsLoading(true);
      await authApi.v1
        .routineServiceRemoveSetFromExerciseInstance(
          id,
          exerciseInstanceId,
          set.id!,
        )
        .then((response) => {
          console.log(response.data);
          fetchExerciseInstanceDetails();
        })
        .catch((error) => {
          console.log(error);
          toast.error("Не удалось удалить подход");
        })
        .finally(() => {
          setIsLoading(false);
        });
    }

    function onChange(newValue: number) {
      newValue = Math.max(0, newValue);

      setValue(newValue);

      if (timeoutRef.current) clearTimeout(timeoutRef.current);

      timeoutRef.current = setTimeout(() => {
        updateSet();
      }, 300);
    }

    return (
      <Card className="flex flex-row items-start gap-3 p-3 w-full">
        <CardBody className="flex flex-col gap-3 p-0">
          <div className="flex flex-row justify-between items-center">
            <div className="flex flex-row gap-1 items-center">
              <div className="h-7 w-7 p-2 flex flex-row items-center justify-center bg-default-100 rounded-small">
                <h2 className="text-sm font-bold text-center">{num}</h2>
              </div>
              <p className="text-sm font-bold">подход</p>
            </div>
            <Button
              isIconOnly
              className="min-w-fit h-fit py-2 px-2 w-7 h-7"
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
              onPress={onDelete}
            >
              <TrashCanIcon className="w-[0.6rem] h-[0.6rem]" />
            </Button>
          </div>
          <div className="flex flex-row justify-between items-center w-full">
            <div className="flex flex-row gap-2 items-center">
              <div className="bg-default-100 rounded-small p-2 h-7 items-center flex flex-row">
                {(() => {
                  switch (set.setType) {
                    case WorkoutSetType.SET_TYPE_REPS:
                      return (
                        <div className="flex flex-row gap-1 items-center">
                          <RepeatIcon className="w-3 h-3" />
                          <p className="text-xs font-regular">Повторения</p>
                        </div>
                      );
                    default:
                      return <div className="w-3 h-3" />;
                  }
                })()}
              </div>
            </div>
            {(() => {
              switch (set.setType) {
                case WorkoutSetType.SET_TYPE_REPS:
                  return (
                    <div className="flex flex-row items-center gap-2">
                      <div className="flex flex-row items-center h-7">
                        <InputWithIncrement
                          allowFloat={false}
                          className="w-14 h-7"
                          classNames={{ incrementButton: "h-7 w-7" }}
                          placeholder="10"
                          setValue={onChange}
                          size="sm"
                          step={1}
                          value={value}
                        />
                      </div>
                    </div>
                  );
                default:
                  return null;
              }
            })()}
          </div>
        </CardBody>
      </Card>
    );
  }

  function SetsList({ sets }: { sets: WorkoutSet[] }) {
    async function onAddSet() {
      await authApi.v1
        .routineServiceAddSetToExerciseInstance(id, exerciseInstanceId, {
          reps:
            exerciseInstanceDetails.sets![
              exerciseInstanceDetails.sets!.length - 1
            ]?.reps! || 8,
          setType: WorkoutSetType.SET_TYPE_REPS,
        })
        .then((response) => {
          console.log(response.data);
          fetchExerciseInstanceDetails();
        })
        .catch((error) => {
          console.log(error);
          toast.error("Не удалось добавить подход");
        });
    }

    return (
      <div className="flex flex-col gap-4 p-4">
        {sets?.map((set, index) => (
          <SetCard key={set.id} num={index + 1} set={set} />
        ))}

        <Card className="p-2">
          <Button fullWidth onPress={onAddSet}>
            <PlusIcon className="w-4 h-4" />
            Добавить подход
          </Button>
        </Card>
      </div>
    );
  }

  function ContentCard() {
    return (
      <div className="pb-4 flex flex-col gap-4">
        {/* <AddSetBlock /> */}
        <SetsList sets={exerciseInstanceDetails.sets!} />
      </div>
    );
  }

  async function onDelete() {
    try {
      await authApi.v1
        .routineServiceRemoveExerciseInstanceFromRoutine(id, exerciseInstanceId)
        .then((response) => {
          console.log(response.data);
          router.back();
        })
        .catch((error) => {
          console.log(error);
          throw error;
        });
    } catch (error) {
      console.log(error);
      toast.error("Не удалось удалить инстанс упражнения");
    }
  }

  return (
    <div className="pt-4 flex flex-col h-full">
      <PageHeader
        enableBackButton
        title={exerciseInstanceDetails.exercise?.name!}
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
      <ContentCard />
    </div>
  );
}
