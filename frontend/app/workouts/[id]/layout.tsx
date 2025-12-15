"use client";

import React, { ReactNode, useCallback, useEffect, use } from "react";
import { useDisclosure } from "@heroui/modal";
import { Button } from "@heroui/button";
import { useSearchParams } from "next/navigation";

import { WorkoutChatPanel } from "@/components/workout-chat";
import { ChatBubbleIcon } from "@/config/icons";

export default function WorkoutLayout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ id: string }>;
}) {
  const resolvedParams = use(params);
  const workoutId = resolvedParams.id;
  const searchParams = useSearchParams();

  const {
    isOpen: isChatOpen,
    onOpen: onChatOpen,
    onClose: onChatClose,
  } = useDisclosure();

  // Открываем чат автоматически и предзаполняем инпут, если переданы параметры
  const prefill = searchParams.get("prefill") ?? undefined;
  const shouldOpenChat = searchParams.get("openChat") === "1";

  // Автоматическое открытие чата при загрузке страницы
  useEffect(() => {
    if (shouldOpenChat) {
      onChatOpen();
      try {
        const url = new URL(window.location.href);

        if (!url.searchParams.has("openChat")) return;

        url.searchParams.delete("openChat");

        const newUrl = `${url.pathname}${url.search ? `?${url.searchParams.toString()}` : ""}${url.hash}`;

        window.history.replaceState(null, "", newUrl);
      } catch {
        // no-op
      }
    }
  }, [shouldOpenChat, onChatOpen]);

  const handleChatClose = useCallback(() => {
    onChatClose();
  }, [onChatClose]);

  const handleToolSuccess = useCallback(() => {
    // Можем отправить событие для обновления данных на странице
    window.dispatchEvent(new CustomEvent("workout-data-updated"));
  }, []);

  return (
    <>
      {children}
      <Button
        isIconOnly
        aria-label="Открыть чат с тренером"
        className="fixed bottom-20 right-6 z-40 shadow-lg w-12 h-12"
        color="secondary"
        radius="full"
        size="sm"
        variant="shadow"
        onPress={onChatOpen}
      >
        <ChatBubbleIcon className="h-6 w-6" />
      </Button>
      <WorkoutChatPanel
        isOpen={isChatOpen}
        prefill={prefill}
        workoutId={workoutId}
        onClose={handleChatClose}
        onToolComplete={handleToolSuccess}
      />
    </>
  );
}
