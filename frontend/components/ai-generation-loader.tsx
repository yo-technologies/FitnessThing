import { useEffect, useState } from "react";
import { Spinner } from "@nextui-org/react";

const messages = [
  "Собираем данные...",
  "Анализируем прогресс...",
  "Подсчитываем нагрузку...",
  "Придумываем упражнения...",
  "Генерируем тренировку...",
];

export const AIGenerationLoader = () => {
  const [currentMessageIndex, setCurrentMessageIndex] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      const timeout = setTimeout(() => {
        setCurrentMessageIndex((prev) => (prev + 1) % messages.length);
      }, 280);

      return () => clearTimeout(timeout);
    }, 2400);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="p-4 flex flex-col flex-grow items-center justify-center gap-4 select-none">
      <Spinner size="lg" />
      <p className="text-sm text-center h-5 leading-5">
        <span className="inline-block font-semibold bg-gradient-to-r from-neutral-400 via-neutral-900 to-neutral-400 bg-[length:200%_100%] animate-ai-shimmer bg-clip-text text-transparent will-change-transform">
          {messages[currentMessageIndex]}
        </span>
      </p>
    </div>
  );
};
