"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Card, ScrollShadow } from "@heroui/react";

import { PageHeader } from "@/components/page-header";
import { Loading } from "@/components/loading";
import { authApi } from "@/api/api";
import { GetWorkoutsResponseWorkoutDetails } from "@/api/api.generated";
import InfiniteScroll, {
  useInfiniteScroll,
} from "@/components/infinite-scroll";

const limit = 10;

function WorkoutHistoryCard({
  workoutDetails,
  onClick,
}: {
  workoutDetails: GetWorkoutsResponseWorkoutDetails;
  onClick: () => void;
}) {
  function getWorkoutDuration(
    workoutDetails: GetWorkoutsResponseWorkoutDetails,
  ) {
    const startDate = new Date(workoutDetails?.workout?.createdAt!);
    const workoutDuration = new Date(
      new Date(workoutDetails?.workout?.finishedAt!).getTime() -
        startDate.getTime(),
    );

    const hours = workoutDuration.getUTCHours();
    const minutes = workoutDuration.getUTCMinutes();

    if (hours === 0) {
      return `${minutes} мин.`;
    }

    return `${hours} ч. ${minutes} мин.`;
  }

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
    <Card isPressable className="flex flex-col p-3 gap-2" onPress={onClick}>
      <div className="flex justify-between items-center w-full">
        <p className="text-md font-bold">{"Тренировка"}</p>
        <p className="text-sm font-semibold text-default-600">
          {/* <p className="text-md font-bold"> */}
          {formatDate(workoutDetails.workout?.createdAt!)}
        </p>
      </div>
      <div className="flex justify-between items-center gap-2 w-full">
        <div className="flex flex-row items-center gap-2">
          <p className="text-xs font-semibold text-default-500">
            {"Упражнения: "}
          </p>
          <p className="text-xs text-default-500">
            {workoutDetails.exerciseLogs?.length}
          </p>
        </div>
        <div className="flex flex-row items-center gap-2">
          <p className="text-xs font-semibold text-default-500">{"Время: "}</p>
          <p className="text-xs text-default-500">
            {getWorkoutDuration(workoutDetails)}
          </p>
        </div>
      </div>
    </Card>
  );
}

function WorkoutHistoryContainer({
  workouts,
  fetchMore,
  hasMore,
}: {
  workouts: GetWorkoutsResponseWorkoutDetails[];
  fetchMore: () => Promise<void>;
  hasMore: boolean;
}) {
  const router = useRouter();

  return (
    <ScrollShadow>
      <InfiniteScroll
        showLoading
        className="flex flex-col gap-4 p-4"
        fetchMore={fetchMore}
        hasMore={hasMore}
      >
        {workouts.map((workout) => (
          <WorkoutHistoryCard
            key={workout.workout?.id}
            workoutDetails={workout}
            onClick={() => {
              router.push(`/profile/workouts/${workout.workout?.id}`);
            }}
          />
        ))}
      </InfiniteScroll>
    </ScrollShadow>
  );
}

export default function WorkoutsHistoryPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);
  const [offset, setOffset] = useState(0);

  const { hasMore, setHasMore } = useInfiniteScroll();

  const [exerciseLogHistory, setExerciseLogHistory] = useState<
    GetWorkoutsResponseWorkoutDetails[]
  >([]);

  async function fetchExerciseLogHistory() {
    await authApi.v1
      .workoutServiceGetWorkouts({ offset, limit })
      .then((response) => {
        console.log(response.data);
        setExerciseLogHistory((prev) => [...prev, ...response.data.workouts!]);
        setHasMore(response.data.workouts!.length === limit);
        setOffset(offset + limit);
      })
      .catch((error) => {
        console.log(error);
        throw error;
      });
  }

  async function fetchData() {
    setIsLoading(true);
    try {
      await fetchExerciseLogHistory();
      setIsError(false);
    } catch {
      setIsError(true);
    }
    setIsLoading(false);
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
    <div className="py-4 flex flex-col h-full">
      <PageHeader enableBackButton title="История" />
      <WorkoutHistoryContainer
        fetchMore={fetchExerciseLogHistory}
        hasMore={hasMore}
        workouts={exerciseLogHistory}
      />
    </div>
  );
}
