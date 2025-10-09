"use client";

import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
} from "react";
import { isAxiosError } from "axios";
import { Button } from "@nextui-org/button";
import { Textarea } from "@nextui-org/input";
import { Spinner } from "@nextui-org/spinner";
import { Divider } from "@nextui-org/divider";
import { Drawer, DrawerContent } from "@nextui-org/drawer";

import { AssistantMarkdown } from "./assistant-markdown";

import {
  WorkoutChat,
  WorkoutChatMessage,
  WorkoutChatMessageRole,
} from "@/api/api.generated";
import { authApi } from "@/api/api";
import {
  WorkoutChatSendRequest,
  WorkoutChatStreamCallbacks,
  sendWorkoutChatMessageStream,
} from "@/api/chat-stream";
import { GearIcon, RightArrowIcon } from "@/config/icons";

type WorkoutChatPanelProps = {
  workoutId: string;
  isOpen: boolean;
  onClose: () => void;
};

type StreamState = {
  status?: string;
  usageTokens?: number;
};

type ChatSessionHandle = ReturnType<typeof sendWorkoutChatMessageStream> | null;

function roleLabel(role?: WorkoutChatMessageRole | null) {
  switch (role) {
    case WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_USER:
      return "Вы";
    case WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_ASSISTANT:
      return "Тренер";
    case WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_TOOL:
      return "Система";
    case WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_SYSTEM:
      return "Система";
    default:
      return "";
  }
}

function MessageBubble({
  message,
  isStreaming,
}: {
  message: WorkoutChatMessage;
  isStreaming?: boolean;
}) {
  const isUser = message.role === WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_USER;
  const isTool = message.role === WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_TOOL;
  const isSystem =
    message.role === WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_SYSTEM;
  const alignment = isUser ? "items-end" : "items-start";
  const bubbleClasses = isUser
    ? "bg-primary text-primary-foreground"
    : "bg-default-200 text-default-800";

  // Отображение system сообщения
  if (isSystem) {
    return null;
  }

  // Отображение tool сообщения
  if (isTool) {
    return (
      <div className="flex w-full flex-col gap-1 items-start">
        <div className="flex items-center gap-2 rounded-medium border border-default-200 bg-default-100 px-3 py-2 text-sm text-default-600">
          <GearIcon className="inline h-3 w-3" />
          <span className="text-xs">{message.toolName}</span>
          {isStreaming && (
            <span className="flex items-center gap-1 text-warning text-xs ml-2">
              <Spinner color="warning" size="sm" />
              <span>выполняется…</span>
            </span>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className={`flex w-full flex-col gap-1 ${alignment}`}>
      <span className="text-xs font-semibold text-default-400">
        {roleLabel(message.role)}
      </span>
      {message.content && (
        <div
          className={`max-w-[90%] rounded-large px-3 py-2 text-sm shadow-sm ${bubbleClasses}`}
        >
          <AssistantMarkdown content={message.content} />
          {isStreaming && <span className="ml-1 animate-pulse">▍</span>}
        </div>
      )}
      {message.error && (
        <span className="text-xs text-danger truncate max-w-[90%]">
          {message.error}
        </span>
      )}
    </div>
  );
}

export function WorkoutChatPanel({
  workoutId,
  isOpen,
  onClose,
}: WorkoutChatPanelProps) {
  const [chat, setChat] = useState<WorkoutChat | undefined>();
  const [messages, setMessages] = useState<WorkoutChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [streamState, setStreamState] = useState<StreamState>({});
  const [streamingAssistantMessage, setStreamingAssistantMessage] =
    useState<string>("");
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamingToolMessage, setStreamingToolMessage] =
    useState<WorkoutChatMessage | null>(null);

  // Ref на скроллируемый контейнер (overflow-y-auto)
  const scrollRef = useRef<HTMLDivElement>(null);
  const sessionRef = useRef<ChatSessionHandle>(null);

  const hasMessages =
    messages.length > 0 || streamingAssistantMessage.length > 0;

  const loadChat = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await authApi.v1.chatServiceGetChat({
        workoutId: workoutId,
      });
      const data = response.data;

      setChat(data.chat ?? undefined);
      setMessages(data.messages ?? []);
    } catch (err) {
      if (isAxiosError(err) && err.response?.status === 404) {
        setChat(undefined);
        setMessages([]);
        setError(null);
      } else {
        console.error("Failed to load workout chat", err);
        setError("Не удалось загрузить чат");
      }
    } finally {
      setLoading(false);
    }
  }, [workoutId]);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    void loadChat();
  }, [isOpen, loadChat]);

  useEffect(() => {
    if (!isOpen && sessionRef.current) {
      sessionRef.current.close();
      sessionRef.current = null;
    }
  }, [isOpen]);

  // Автоскролл сообщений (основной список)
  useEffect(() => {
    const el = scrollRef.current;

    if (!el) return;

    // Используем requestAnimationFrame для гарантированного рендера DOM
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
      });
    });
  }, [messages]);

  // Автоскролл для потоковой дельты
  useEffect(() => {
    if (!streamingAssistantMessage) return;
    const el = scrollRef.current;

    if (!el) return;
    el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
  }, [streamingAssistantMessage]);

  // Автоскролл для streaming tool сообщений
  useEffect(() => {
    if (!streamingToolMessage) return;
    const el = scrollRef.current;

    if (!el) return;
    el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
  }, [streamingToolMessage]);

  const sendMessage = useCallback(
    async (content: string) => {
      const trimmed = content.trim();

      if (!trimmed || isStreaming) {
        return;
      }

      const userMessage: WorkoutChatMessage = {
        id: `temp-${Date.now()}`,
        chatId: chat?.id,
        role: WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_USER,
        content: trimmed,
        createdAt: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, userMessage]);
      setInputValue("");
      setStreamingAssistantMessage("");
      setStreamState({ status: "assistant_thinking" });
      setIsStreaming(true);
      setError(null);

      const request: WorkoutChatSendRequest = {
        content: trimmed,
        chatId: chat?.id,
        workoutId,
      };

      const callbacks: WorkoutChatStreamCallbacks = {
        onMessageDelta: (delta) => {
          console.log("Received message delta:", delta);
          setStreamingAssistantMessage((prev) => prev + delta);
        },
        onStatus: (status) => {
          console.log("Received status update:", status);

          // Инструмент запускается: "invoking tool <name>"
          if (status.startsWith("invoking tool ")) {
            const toolName = status.replace("invoking tool ", "").trim();

            // Финализируем уже накопленный assistant streaming текст как отдельное сообщение
            setStreamingAssistantMessage((prev) => {
              if (prev && prev.length > 0) {
                setMessages((msgs) => [
                  ...msgs,
                  {
                    id: `assistant-chunk-${Date.now()}`,
                    chatId: chat?.id,
                    role: WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_ASSISTANT,
                    content: prev,
                    createdAt: new Date().toISOString(),
                  },
                ]);
              }

              return ""; // очистим, чтобы последующий контент после tool шёл в новое сообщение
            });

            setStreamingToolMessage((prev) => {
              // Если уже есть streaming сообщение с другим инструментом, завершаем его
              if (prev && prev.toolName && prev.toolName !== toolName) {
                setMessages((msgs) => [...msgs, prev]);

                // Создаем новое streaming сообщение для нового инструмента
                return {
                  id: `streaming-tool-${Date.now()}`,
                  chatId: chat?.id,
                  role: WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_TOOL,
                  content: "",
                  toolName,
                  createdAt: new Date().toISOString(),
                };
              }

              // Если уже есть streaming с тем же инструментом, оставляем как есть
              if (prev && prev.toolName === toolName) {
                return prev;
              }

              // Создаем новое streaming сообщение
              return {
                id: `streaming-tool-${Date.now()}`,
                chatId: chat?.id,
                role: WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_TOOL,
                content: "",
                toolName,
                createdAt: new Date().toISOString(),
              };
            });

            return;
          }

          // Инструмент завершён: "tool <name> completed"
          if (status.startsWith("tool ") && status.endsWith(" completed")) {
            const toolName = status
              .replace(/^tool /, "")
              .replace(/ completed$/, "")
              .trim();

            setStreamingToolMessage((prev) => {
              if (prev && prev.toolName === toolName) {
                // Завершаем текущее streaming сообщение инструмента
                setMessages((msgs) => [...msgs, prev]);

                return null;
              }

              return prev;
            });

            return;
          }

          setStreamState((prev) => ({ ...prev, status }));
        },
        onUsage: (usage) => {
          setStreamState((prev) => ({
            ...prev,
            usageTokens: usage.totalTokens ?? 0,
          }));
        },
        onFinal: (final) => {
          setChat(final.chat ?? undefined);
          setMessages(final.messages ?? []);
          setStreamingAssistantMessage("");
          setStreamingToolMessage(null);
          setStreamState({});
          setIsStreaming(false);
        },
        onError: (err) => {
          console.error("Workout chat stream error", err);
          setError(err.message);
          setIsStreaming(false);
          setStreamingAssistantMessage("");
          setStreamingToolMessage(null);
          // Перезагружаем чат, чтобы получить корректное состояние с сервера
          void loadChat();
        },
      };

      const session = sendWorkoutChatMessageStream(request, callbacks);

      sessionRef.current = session;

      session.done
        .catch((err) => {
          console.error("Workout chat stream failure", err);
          setError(err.message);
        })
        .finally(() => {
          setIsStreaming(false);
          setStreamingAssistantMessage("");
          setStreamState({});
          sessionRef.current = null;
        });
    },
    [chat?.id, isStreaming, workoutId],
  );

  const handleSend = useCallback(() => {
    void sendMessage(inputValue);
  }, [inputValue, sendMessage]);

  const handleKeyDown = useCallback(
    (event: ReactKeyboardEvent<HTMLTextAreaElement>) => {
      if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault();
        handleSend();
      }
    },
    [handleSend],
  );

  const streamingMessage = useMemo<WorkoutChatMessage | null>(() => {
    if (!streamingAssistantMessage) {
      return null;
    }

    return {
      id: "streaming",
      chatId: chat?.id,
      role: WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_ASSISTANT,
      content: streamingAssistantMessage,
      createdAt: new Date().toISOString(),
    };
  }, [chat?.id, streamingAssistantMessage]);

  const content = useMemo(() => {
    if (loading) {
      return (
        <div className="flex h-full flex-col items-center justify-center gap-2">
          <Spinner color="secondary" size="lg" />
          <p className="text-sm text-default-500">Загружаем чат…</p>
        </div>
      );
    }

    if (error && !hasMessages) {
      return (
        <div className="flex h-full flex-col items-center justify-center gap-4">
          <p className="text-sm text-default-500">{error}</p>
          <Button color="secondary" size="sm" onPress={() => void loadChat()}>
            Повторить
          </Button>
        </div>
      );
    }

    if (!hasMessages) {
      return (
        <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center text-default-500">
          <p className="text-sm">
            Общайтесь с тренером, чтобы настроить тренировку.
          </p>
          <p className="text-xs text-default-400">
            Задайте вопросы, попросите сгенерировать тренировку или изменить
            упражнения.
          </p>
        </div>
      );
    }

    return (
      <div className="flex h-full flex-col gap-3 p-4">
        {messages.map((message) => (
          <MessageBubble key={message.id} message={message} />
        ))}
        {streamingMessage && (
          <MessageBubble isStreaming message={streamingMessage} />
        )}
        {streamingToolMessage && (
          <MessageBubble isStreaming message={streamingToolMessage} />
        )}
      </div>
    );
  }, [
    error,
    hasMessages,
    loadChat,
    loading,
    messages,
    streamingMessage,
    streamingToolMessage,
  ]);

  const canSend = inputValue.trim().length > 0 && !isStreaming;

  return (
    <Drawer
      hideCloseButton
      backdrop="blur"
      isDismissable={false}
      isOpen={isOpen}
      placement="bottom"
      size="full"
      onClose={onClose}
    >
      <DrawerContent>
        <div className="flex h-[100dvh] flex-col bg-content1">
          <header className="flex items-center justify-between px-4 py-3">
            <div>
              <h2 className="text-lg font-semibold">Чат с тренером</h2>
              {chat?.title && (
                <p className="text-xs text-default-400">{chat.title}</p>
              )}
            </div>
            <Button color="danger" size="sm" variant="flat" onPress={onClose}>
              Закрыть
            </Button>
          </header>

          <Divider />

          <div className="flex flex-1 flex-col overflow-hidden">
            <div ref={scrollRef} className="flex-1 overflow-y-auto">
              {content}
            </div>

            <div className="px-4 pb-2 text-[11px] text-default-400 min-h-[20px] flex flex-col gap-0.5">
              {streamState.status === "assistant_thinking" && (
                <p className="flex items-center gap-1 text-xs">
                  <Spinner
                    classNames={{ wrapper: "w-3 h-3" }}
                    color="secondary"
                    size="sm"
                  />{" "}
                  Думает…
                </p>
              )}
              {streamState.status === "assistant_completed" && (
                <p className="text-xs text-success">Готово</p>
              )}
              {streamState.usageTokens ? (
                <p className="text-xs text-[10px] opacity-70">
                  Токенов: {streamState.usageTokens}
                </p>
              ) : null}
            </div>

            {error && hasMessages && (
              <div className="px-4 pb-2 text-xs text-danger">{error}</div>
            )}

            <div className="flex items-center gap-2 border-t border-default-200 px-4 py-2 mb-4">
              <Textarea
                className="flex-1"
                classNames={{
                  inputWrapper: "bg-default-100",
                }}
                minRows={2}
                placeholder="Напишите сообщение..."
                value={inputValue}
                onChange={(event) => setInputValue(event.target.value)}
                onKeyDown={(event) =>
                  handleKeyDown(
                    event as unknown as ReactKeyboardEvent<HTMLTextAreaElement>,
                  )
                }
              />
              <Button
                isIconOnly
                aria-label="Отправить сообщение"
                className="shrink-0"
                color="secondary"
                isDisabled={!canSend}
                isLoading={isStreaming}
                radius="full"
                onPress={handleSend}
              >
                <RightArrowIcon className="h-5 w-5" />
              </Button>
            </div>
          </div>
        </div>
      </DrawerContent>
    </Drawer>
  );
}
