/* eslint-disable react/no-unknown-property */
"use client";
import { Button } from "@nextui-org/button";
import { Card, CardFooter, CardHeader } from "@nextui-org/card";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";

import { WorkoutWorkout } from "@/api/api.generated";
import { AnimationProcessor } from "@/components/animated-background";
import { BoltIcon, ChevronRightIcon, PlayIcon } from "@/config/icons";
import { Loading } from "@/components/loading";
import { OnboardingModal } from "@/components/OnboardingModal";
import { authApi } from "@/api/api";

export default function Home() {
  const [user, setUser] = useState<any>({});
  const [routines, setRoutines] = useState<any[]>([]);
  const [activeWorkouts, setActiveWorkouts] = useState<WorkoutWorkout[]>([]);

  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);
  const [isWorkoutGenerating, setIsWorkoutGenerating] = useState(false);

  // Состояние для онбординга
  const [showOnboarding, setShowOnboarding] = useState(false);
  const [hasCheckedOnboarding, setHasCheckedOnboarding] = useState(false);

  const router = useRouter();

  async function fetchUser() {
    await authApi.v1
      .userServiceGetMe()
      .then((response) => {
        console.log(response.data);
        setUser(response.data.user!);
      })
      .catch((error) => {
        console.log(error);
        throw error;
      });
  }

  async function fetRoutines() {
    await authApi.v1
      .routineServiceGetRoutines()
      .then((response) => {
        console.log(response.data);
        setRoutines(response.data.routines!);
      })
      .catch((error) => {
        console.log(error);
        throw error;
      });
  }

  async function fetchActiveWorkouts() {
    await authApi.v1
      .workoutServiceGetActiveWorkouts()
      .then((response) => {
        console.log(response.data);
        setActiveWorkouts(response.data.workouts!);
      })
      .catch((error) => {
        console.log(error);
        throw error;
      });
  }

  async function fetchData() {
    setIsLoading(true);
    try {
      await Promise.all([fetchUser(), fetRoutines(), fetchActiveWorkouts()]);
      setIsError(false);

      // Проверяем, нужно ли показать онбординг
      checkOnboardingStatus();
    } catch (error) {
      console.log(error);
      setIsError(true);
    } finally {
      setIsLoading(false);
    }
  }

  // Проверяем статус онбординга
  async function checkOnboardingStatus() {
    if (hasCheckedOnboarding) return;

    try {
      const response =
        await authApi.v1.userServiceGetWorkoutGenerationSettings();
      const settings = response.data.settings;

      // Если настройки пустые или основные поля не заполнены, показываем онбординг
      const isOnboardingNeeded =
        !settings ||
        !settings.primaryGoal ||
        !settings.experienceLevel ||
        !settings.workoutPlanType;

      if (isOnboardingNeeded) {
        setShowOnboarding(true);
      }

      setHasCheckedOnboarding(true);
    } catch (error: any) {
      console.log("Error checking onboarding status:", error);

      // Если получили 404, значит настройки еще не созданы - нужен онбординг
      if (error.response?.status === 404) {
        setShowOnboarding(true);
      }

      setHasCheckedOnboarding(true);
    }
  }

  const handleOnboardingComplete = () => {
    setShowOnboarding(false);
    toast.success("Добро пожаловать! Настройки сохранены.");
  };

  const handleOnboardingClose = () => {
    setShowOnboarding(false);
  };

  async function startWorkout(
    routineId: string | undefined,
    generate: boolean = false,
  ) {
    if (activeWorkouts.length > 0) {
      toast.error("Сначала завершите активную тренировку");

      return;
    }
    setIsWorkoutGenerating(true);
    await authApi.v1
      .workoutServiceStartWorkout({
        routineId: routineId,
        generateWorkout: generate,
      })
      .then((response) => {
        console.log(response.data);
        router.push(`/workouts/${response.data.workout?.id}`);
      })
      .catch((error) => {
        console.log(error);
        setIsWorkoutGenerating(false);

        if (error.response?.status === 429) {
          toast.error("Превышен лимит генераций на сегодня");

          return;
        }

        toast.error("Ошибка при начале тренировки");
        throw error;
      });
  }

  useEffect(() => {
    fetchData();
  }, []);

  useEffect(() => {
    const canvas = document.getElementById("home-bg") as HTMLCanvasElement;

    if (!canvas) return;

    const animationProcessor = new AnimationProcessor(canvas, 400, 1);

    const updateAnimationDimensions = () => {
      const container = canvas.parentElement?.parentElement;

      if (!container) return;

      const width = container.clientWidth;
      const height = container.clientHeight;

      // Update center to be in the middle of the container
      animationProcessor.updateCenter(width / 2, height / 2);

      // Update radius to be 60% of the smallest dimension
      const radius = Math.min(width, height) * 0.6;

      animationProcessor.updateRadius(radius);
    };

    // Initial setup
    updateAnimationDimensions();
    animationProcessor.start();

    // Add resize listener
    window.addEventListener("resize", updateAnimationDimensions);

    return () => {
      animationProcessor.stop();
      window.removeEventListener("resize", updateAnimationDimensions);
    };
  }, [isLoading]);

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

  function RoutineCard({ routine }: { routine: any }) {
    return (
      <Card key={routine.id} className="w-52 h-52 p-3 gap-1">
        <Link
          className="flex flex-col h-full gap-1"
          href={`/routines/${routine.id}`}
        >
          <h3 className="text-lg font-bold">{routine.name}</h3>

          <div className="flex flex-col h-full overflow-hidden">
            <p className="text-xs font-light line-clamp-[7]">
              {routine.description}
            </p>
          </div>
        </Link>
        <Button
          className="flex items-center p-2 w-full"
          color="primary"
          size="sm"
          onPress={async () => {
            await startWorkout(routine.id);
          }}
        >
          <PlayIcon className="w-3 h-3" fill="currentColor" />
          <span className="text-sm font-bold">Начать</span>
        </Button>
      </Card>
    );
  }

  function RoutinesSection() {
    return (
      <section className="flex flex-col flex-grow max-w-full h-fit relative max-w-full">
        {/* Список шаблонов. Горизонтальный скролл с квадратными карточками */}
        <Link className="flex items-center px-4 gap-1" href="/routines">
          <h4 className="text-xl font-bold">Шаблоны</h4>
          <ChevronRightIcon className="w-4 h-4" fill="currentColor" />
        </Link>
        <div className="flex flex-col p-4 max-w-full overflow-x-auto">
          <div className="flex flex-row gap-4 justify-start w-fit">
            {routines.map((routine) => (
              <RoutineCard key={routine.id} routine={routine} />
            ))}
            {routines.length === 0 && (
              <p className="p-0 text-xs">Ты еще не создал ни одного шаблона</p>
            )}
          </div>
        </div>
      </section>
    );
  }

  return (
    <div className="flex flex-col flex-grow max-w-full justify-start">
      <section
        className="flex flex-col flex-grow max-w-full h-[65vh] relative"
        id="home"
      >
        <div className="flex flex-col flex-grow absolute w-full h-[75vh] z-0 fade-bottom opacity-80 contrast-[1.15]">
          <canvas id="home-bg" />
        </div>
        <div className="flex-grow flex flex-col justify-center items-center relative z-10 drop-shadow-lg">
          <h1 className="text-2xl font-bold p-4 absolute top-0 left-0">
            Привет{user.firstName ? `, ${user.firstName}` : ""}!
          </h1>
          <Button
            disableRipple
            className="flex items-center bg-transparent h-fit"
            size="lg"
            onPress={async () => {
              await startWorkout(undefined, true);
            }}
          >
            <BoltIcon className="w-7 h-7" fill="currentColor" />
            <span className="text-2xl font-bold">Стать лучше</span>
          </Button>
          <Button
            className="flex items-center text-white-500 bg-transparent underline p-0"
            size="sm"
            onPress={async () => {
              await startWorkout(undefined, false);
            }}
          >
            Пустая тренировка
          </Button>
        </div>
      </section>
      {activeWorkouts.length > 0 && (
        <section className="flex flex-col flex-grow max-w-full h-fit relative">
          <h4 className="text-xl font-bold px-4">Активные тренировки</h4>
          <div className="flex flex-col p-4 max-w-full gap-4">
            {activeWorkouts.map((workout) => (
              <Card key={workout.id} className="w-full p-3 gap-4">
                <CardHeader className="p-0">
                  <h3 className="text-lg font-bold">
                    {"Тренировка "}
                    {new Date(workout.createdAt!).toLocaleString("ru-RU", {
                      weekday: "long",
                      day: "numeric",
                      month: "long",
                    })}
                  </h3>
                </CardHeader>
                <CardFooter className="p-0 rounded-sm">
                  <Button
                    className="flex items-center px-2 w-full"
                    color="primary"
                    size="sm"
                    onPress={async () => {
                      router.push(`/workouts/${workout.id}`);
                    }}
                  >
                    <PlayIcon className="w-3 h-3" fill="currentColor" />
                    <span className="text-sm font-bold">Продолжить</span>
                  </Button>
                </CardFooter>
              </Card>
            ))}
          </div>
        </section>
      )}
      <RoutinesSection />
      <style jsx>{`
        .fade-bottom::after {
          content: "";
          position: absolute;
          bottom: 0;
          left: 0;
          width: 100%;
          height: 50px;
          background: linear-gradient(
            to top,
            hsl(var(--nextui-background)),
            transparent 100%
          );
          pointer-events: none;
        }
      `}</style>

      {isWorkoutGenerating && (
        <div className="h-full w-full fixed top-0 left-0 z-10 backdrop-blur-sm">
          <div className="flex flex-col items-center justify-center gap-4 h-full bg-background bg-opacity-50">
            <div className="flex flex-col items-center">
              <h2 className="text-xl font-bold">Генерация тренировки...</h2>
              <p className="text-xs font-light text-default-500">
                Это может занять некоторое время
              </p>
            </div>
            <div>
              <Loading showText={false} size="lg" />
            </div>
          </div>
        </div>
      )}

      {/* Модальное окно онбординга */}
      <OnboardingModal
        isOpen={showOnboarding}
        onClose={handleOnboardingClose}
        onComplete={handleOnboardingComplete}
      />
    </div>
  );
}
