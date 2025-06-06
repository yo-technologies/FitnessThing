"use client";
import { Button } from "@nextui-org/button";
import { Card, CardBody, CardHeader } from "@nextui-org/card";
import {
  Modal,
  ModalContent,
  ModalHeader,
  useDisclosure,
} from "@nextui-org/modal";
import { Form } from "@nextui-org/form";
import { Input } from "@nextui-org/input";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { ChevronRightIcon, PlusIcon } from "@/config/icons";
import { PageHeader } from "@/components/page-header";
import { Loading } from "@/components/loading";
import { authApi } from "@/api/api";
import { WorkoutRoutine } from "@/api/api.generated";

export default function RoutinesPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  const { isOpen, onOpen, onOpenChange } = useDisclosure();

  const [routines, setRoutines] = useState<WorkoutRoutine[]>([]);

  const router = useRouter();

  async function fetchData() {
    setIsLoading(true);
    authApi.v1
      .routineServiceGetRoutines()
      .then((response) => {
        console.log(response.data);
        setIsError(false);
        setRoutines(response.data.routines!);
      })
      .catch((error) => {
        console.log(error);
        setIsError(true);
      })
      .finally(() => {
        setIsLoading(false);
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

  function declension(
    exerciseCount: number | undefined,
    forms: string[],
  ): import("react").ReactNode {
    if (exerciseCount === undefined) return null;

    const form =
      exerciseCount % 10 === 1 && exerciseCount % 100 !== 11
        ? forms[0]
        : exerciseCount % 10 >= 2 &&
            exerciseCount % 10 <= 4 &&
            (exerciseCount % 100 < 12 || exerciseCount % 100 > 14)
          ? forms[1]
          : forms[2];

    return <span>{form}</span>;
  }

  return (
    <>
      <div className="py-4 flex-grow">
        <PageHeader enableBackButton title="Шаблоны" />
        <div className="grid grid-cols-1 gap-4 p-4">
          {routines.map((routine) => (
            <Card
              key={routine.id}
              fullWidth
              isPressable
              className="flex flex-row items-center justify-between p-3 cursor-pointer gap-3"
              shadow="sm"
              onPress={() => router.push(`/routines/${routine.id}`)}
            >
              <div className="flex flex-col items-start justify-between w-full">
                <CardHeader className="p-0">
                  <p className="text-md font-bold text-start">{routine.name}</p>
                </CardHeader>
                <CardBody className="p-0">
                  <p className="text-xs text-default-400">
                    {routine.exerciseCount}{" "}
                    {declension(routine.exerciseCount, [
                      "упражнение",
                      "упражнения",
                      "упражнений",
                    ])}
                  </p>
                </CardBody>
              </div>
              <div className="flex flex-col items-center justify-between">
                <ChevronRightIcon className="w-4 h-4" fill="currentColor" />
              </div>
            </Card>
          ))}
          <Card className="p-2">
            <Button onPress={onOpen}>
              <PlusIcon className="w-4 h-4" />
              Добавить рутину
            </Button>
          </Card>
        </div>
      </div>
      <Modal
        isOpen={isOpen}
        placement="center"
        size="xs"
        onClose={onOpenChange}
      >
        <ModalContent>
          {() => (
            <>
              <ModalHeader>
                <h2>Добавить рутину</h2>
              </ModalHeader>
              <Form
                className="inline-block text-center justify-center w-full max-w-lg p-4"
                validationBehavior="native"
                onSubmit={async (e: React.FormEvent<HTMLFormElement>) => {
                  e.preventDefault();
                  const data = Object.fromEntries(
                    new FormData(e.currentTarget),
                  );

                  console.log(data);
                  await authApi.v1
                    .routineServiceCreateRoutine({
                      name: data.name.toString(),
                      description: data.description.toString(),
                    })
                    .then((response) => {
                      console.log(response.data);
                      router.push(`/routines/${response.data.routine?.id}`);
                    })
                    .catch((error) => {
                      console.log(error);

                      return error;
                    });
                }}
              >
                <div className="flex flex-col items-center justify-center gap-4 py-4">
                  <Input
                    autoFocus
                    isRequired
                    label="Название"
                    labelPlacement="outside"
                    name="name"
                    placeholder="Название"
                    type="text"
                  />
                  <Input
                    label="Описание"
                    labelPlacement="outside"
                    name="description"
                    placeholder="Описание"
                    type="text"
                  />
                  <Button color="primary" type="submit">
                    Добавить
                  </Button>
                </div>
              </Form>
            </>
          )}
        </ModalContent>
      </Modal>
    </>
  );
}
