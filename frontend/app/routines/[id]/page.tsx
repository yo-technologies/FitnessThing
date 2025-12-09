"use client";

import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { DropdownItem } from "@heroui/dropdown";
import { Form } from "@heroui/form";
import { Input, Textarea } from "@heroui/input";
import {
  Modal,
  ModalContent,
  ModalHeader,
  useDisclosure,
} from "@heroui/modal";
import { useRouter } from "next/navigation";
import { use, useEffect, useState } from "react";
import { addToast } from "@heroui/toast";
import {
  DndContext,
  DragEndEvent,
  KeyboardSensor,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { hapticFeedback } from "@tma.js/sdk-react";

import { ChevronRightIcon, GripVerticalIcon, PlusIcon } from "@/config/icons";
import { ModalSelectExercise } from "@/components/pick-exercises-modal";
import { PageHeader } from "@/components/page-header";
import { Loading } from "@/components/loading";
import { translateMuscleGroups } from "@/config/muscle-groups";
import {
  WorkoutExerciseInstanceDetails,
  WorkoutRoutineDetailResponse,
} from "@/api/api.generated";
import { authApi } from "@/api/api";

export default function RoutineDetailsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const [isLoading, setIsLoading] = useState(true);

  const [routineDetails, setRoutineDetails] =
    useState<WorkoutRoutineDetailResponse>({});

  const { id } = use(params);

  const { isOpen, onOpen, onClose } = useDisclosure();

  const {
    isOpen: isRenameOpen,
    onOpen: onRenameOpen,
    onOpenChange: onRenameOpenChange,
  } = useDisclosure();

  const sensors = useSensors(
    useSensor(TouchSensor, {
      activationConstraint: {
        delay: 250,
        tolerance: 25,
      },
    }),
    useSensor(MouseSensor, {
      activationConstraint: {
        distance: 10,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  const router = useRouter();

  async function fetchRoutineDetails() {
    console.log(id);
    await authApi.v1
      .routineServiceGetRoutineDetail(id)
      .then((response) => {
        console.log(response.data);
        setRoutineDetails(response.data);
      })
      .catch((error) => {
        console.log(error);
        addToast({ title: "Ошибка при загрузке данных", color: "danger" });
      });
  }

  async function addExerciseToRoutine(exerciseId: string) {
    await authApi.v1
      .routineServiceAddExerciseToRoutine(id, { exerciseId: exerciseId })
      .then((response) => {
        console.log(response.data);
      })
      .catch((error) => {
        console.log(error);
        addToast({ title: "Ошибка при добавлении упражнения", color: "danger" });
      });
  }

  async function submitPickExercise(exerciseIds: string[]) {
    console.log(exerciseIds);
    try {
      for (const exerciseId of exerciseIds) {
        await addExerciseToRoutine(exerciseId);
      }
      await fetchRoutineDetails();
    } catch (error) {
      console.log(error);
      addToast({ title: "Ошибка при добавлении упражнения", color: "danger" });
    }
  }

  async function updateExerciseOreder(ids: string[]) {
    await authApi.v1
      .routineServiceSetExerciseOrder(id, { exerciseInstanceIds: ids })
      .then((response) => {
        console.log(response.data);
      })
      .catch((error) => {
        console.log(error);
        addToast({ title: "Ошибка при обновлении порядка упражнений", color: "danger" });
      });
  }

  function onDragStart() {
    if ("vibrate" in navigator) {
      navigator.vibrate(100);
    }

    if (hapticFeedback.impactOccurred.isAvailable()) {
      hapticFeedback.impactOccurred("medium");
    }
  }

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;

    if (!over || !active) {
      return;
    }

    if (active.id === over!.id) {
      return;
    }

    const activeIndex = routineDetails.exerciseInstances!.findIndex(
      (ei) => ei.exerciseInstance!.id === active.id,
    );

    const overIndex = routineDetails.exerciseInstances!.findIndex(
      (ei) => ei.exerciseInstance!.id === over!.id,
    );

    const newExerciseInstances = [...routineDetails.exerciseInstances!];

    newExerciseInstances.splice(activeIndex, 1);
    newExerciseInstances.splice(
      overIndex,
      0,
      routineDetails.exerciseInstances![activeIndex],
    );

    updateExerciseOreder(
      newExerciseInstances.map((ei) => ei.exerciseInstance!.id!),
    );

    setRoutineDetails((prev) => ({
      ...prev,
      exerciseInstances: newExerciseInstances,
    }));
  }

  async function fetchData() {
    setIsLoading(true);
    await fetchRoutineDetails();
    setIsLoading(false);
  }

  useEffect(() => {
    fetchData();
  }, []);

  if (isLoading) {
    return <Loading />;
  }

  function ExercieInstanceCard({
    exerciseInstanceDetails,
  }: {
    exerciseInstanceDetails: WorkoutExerciseInstanceDetails;
  }) {
    const setsCount = exerciseInstanceDetails.sets?.length || 0;

    const { attributes, listeners, setNodeRef, transform, transition } =
      useSortable({ id: exerciseInstanceDetails.exerciseInstance!.id! });

    const style = {
      transform: CSS.Transform.toString(transform),
      transition,
    };

    return (
      <div
        ref={setNodeRef}
        className="select-none touch-manipulation"
        style={style}
      >
        <Card
          fullWidth
          className="flex flex-row flex-grow p-3 gap-3 items-center"
          shadow="sm"
        >
          <div
            className="rounded-md bg-default-100 cursor-grab flex items-center justify-center p-1"
            {...attributes}
            {...listeners}
          >
            <GripVerticalIcon className="w-4 h-4" />
          </div>
          <button
            className="flex items-center justify-between max-w-full w-full"
            onClick={() =>
              router.push(
                `/routines/${id}/exerciseInstance/${exerciseInstanceDetails.exerciseInstance!.id}`,
              )
            }
          >
            <div className="flex flex-col items-start justify-between">
              <CardHeader className="p-0">
                <p className="text-md font-bold text-start">
                  {exerciseInstanceDetails.exercise!.name}
                </p>
              </CardHeader>
              <CardBody className="p-0">
                <div className="flex flex-row gap-1">
                  <p className="text-xs text-default-400">
                    {setsCount}{" "}
                    {"подход" +
                      (setsCount % 10 === 1
                        ? ""
                        : setsCount % 10 >= 2 && setsCount % 10 <= 4
                          ? "а"
                          : "ов")}
                    {" • "}
                    {translateMuscleGroups(
                      exerciseInstanceDetails.exercise!.targetMuscleGroups!,
                    ).join(", ")}
                  </p>
                </div>
              </CardBody>
            </div>
            <div className="flex flex-col items-center justify-center">
              <ChevronRightIcon className="w-4 h-4" fill="currentColor" />
            </div>
          </button>
        </Card>
      </div>
    );
  }

  function ModalRenameRoutine() {
    const [routineName, setRoutineName] = useState(
      routineDetails.routine?.name,
    );
    const [routineDescription, setRoutineDescription] = useState(
      routineDetails.routine?.description,
    );
    const [isButtonLoading, setIsButtonLoading] = useState(false);

    async function updateRoutine() {
      setIsButtonLoading(true);
      try {
        await authApi.v1.routineServiceUpdateRoutine(id, {
          name: routineName,
          description: routineDescription,
        });
        await fetchRoutineDetails();
      } catch (error) {
        console.log(error);
        addToast({ title: "Ошибка при обновлении рутины", color: "danger" });
      } finally {
        setIsButtonLoading(false);
      }
    }

    return (
      <Modal
        className="overflow-y-auto p-2 w-full"
        isOpen={isRenameOpen}
        placement="center"
        scrollBehavior="inside"
        size="xs"
        onClose={onRenameOpenChange}
      >
        <ModalContent>
          {(onClose) => (
            <>
              <ModalHeader className="p-2">Переименовать рутину</ModalHeader>

              <Form
                className="inline-block text-center justify-center w-full max-w-lg p-0"
                onSubmit={async (e) => {
                  e.preventDefault();
                  await updateRoutine();
                  onClose();
                }}
              >
                <div className="grid grid-cols-1 gap-4 p-2">
                  <Input
                    autoFocus
                    placeholder="Название"
                    value={routineName}
                    onChange={(e) => setRoutineName(e.target.value)}
                  />
                  <Textarea
                    placeholder="Описание"
                    value={routineDescription ? routineDescription : ""}
                    onChange={(e) => setRoutineDescription(e.target.value)}
                  />
                  <Button
                    color="primary"
                    isLoading={isButtonLoading}
                    type="submit"
                  >
                    Сохранить
                  </Button>
                </div>
              </Form>
            </>
          )}
        </ModalContent>
      </Modal>
    );
  }

  async function onDelete() {
    await authApi.v1
      .routineServiceDeleteRoutine(id)
      .catch((error) => {
        console.log(error);
        addToast({ title: "Ошибка при удалении рутины", color: "danger" });
      })
      .then(() => {
        router.back();
      });
  }

  return (
    <>
      <div className="flex flex-col py-4 gap-4">
        <PageHeader enableBackButton title={routineDetails.routine?.name!}>
          <DropdownItem key="rename" color="primary" onPress={onRenameOpen}>
            Переименовать
          </DropdownItem>
          <DropdownItem
            key="delete"
            className="text-danger"
            color="danger"
            onPress={onDelete}
          >
            Удалить
          </DropdownItem>
        </PageHeader>
        {routineDetails.routine?.description ? (
          <div className="px-4">
            <p className="text-sm text-gray-500">
              {routineDetails.routine?.description}
            </p>
          </div>
        ) : null}
        <div className="flex flex-col gap-4 px-4">
          <DndContext
            sensors={sensors}
            onDragEnd={handleDragEnd}
            onDragStart={onDragStart}
          >
            <SortableContext
              items={routineDetails.exerciseInstances!.map(
                (ei) => ei.exerciseInstance!.id!,
              )}
              strategy={verticalListSortingStrategy}
            >
              {routineDetails.exerciseInstances?.map(
                (exerciseInstanceDetails: WorkoutExerciseInstanceDetails) => (
                  <ExercieInstanceCard
                    key={exerciseInstanceDetails.exerciseInstance!.id}
                    exerciseInstanceDetails={exerciseInstanceDetails}
                  />
                ),
              )}
              <Card className="p-2">
                <Button onPress={onOpen}>
                  <PlusIcon className="w-4 h-4" />
                  Добавить упражнение
                </Button>
              </Card>
            </SortableContext>
          </DndContext>
        </div>
      </div>
      <ModalSelectExercise
        excludeExerciseIds={routineDetails.exerciseInstances!.map(
          (ei) => ei.exercise!.id!,
        )}
        isOpen={isOpen}
        onClose={onClose}
        onSubmit={submitPickExercise}
      />

      <ModalRenameRoutine />
    </>
  );
}
