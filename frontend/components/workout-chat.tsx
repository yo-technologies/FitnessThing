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
import { ScrollShadow } from "@nextui-org/react";

import { AssistantMarkdown } from "./assistant-markdown";

import {
  ToolEventState,
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
import { formatToolLabel } from "@/config/tools";

type WorkoutChatPanelProps = {
  workoutId: string;
  isOpen: boolean;
  onClose: () => void;
  prefill?: string;
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

function formatTime(iso?: string | null) {
  if (!iso) return "";
  try {
    const d = new Date(iso);

    // Формат HH:MM по локали пользователя
    return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  } catch {
    return "";
  }
}

function ToolChip({
  toolName,
  isError,
  isStreaming,
  createdAt,
}: {
  toolName: string;
  isError?: boolean;
  isStreaming?: boolean;
  createdAt?: string;
}) {
  const label = useMemo(() => formatToolLabel(toolName), [toolName]);

  return (
    <div className="flex w-full flex-col gap-1 items-start">
      <div
        className={
          `flex items-baseline gap-2 rounded-medium px-2 py-2 text-sm w-fit max-w-full bg-default-100 text-default-600 border ` +
          (isError ? "border-danger-200" : "border-default-200")
        }
      >
        <GearIcon className={`h-3 w-3 self-center`} />
        <span className="text-xs">{label}</span>

        {isStreaming && (
          <span className="flex items-center gap-1 text-warning text-xs ml-2">
            <Spinner
              classNames={{ wrapper: "w-3 h-3" }}
              color="warning"
              size="sm"
            />
          </span>
        )}
        {createdAt && (
          <span className="ml-2 text-[10px] font-light leading-none text-default-500 whitespace-nowrap">
            {formatTime(createdAt)}
          </span>
        )}
      </div>
    </div>
  );
}

function MessageBubble({
  message,
  isStreaming,
  showHeader,
}: {
  message: WorkoutChatMessage;
  isStreaming?: boolean;
  showHeader?: boolean;
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

  if (isTool) {
    const legacyError =
      (message.content ?? "").trim().toLowerCase().startsWith("error:") ||
      false;

    return (
      <ToolChip
        createdAt={message.createdAt ?? undefined}
        isError={Boolean(message.error) || legacyError}
        isStreaming={isStreaming}
        toolName={message.toolName ?? "Инструмент"}
      />
    );
  }

  if (!message.content && !message.error && !isStreaming && !showHeader) {
    return null;
  }

  return (
    <div
      className={`flex w-full flex-col gap-2 ${alignment}`}
      data-message-id={message.id ?? undefined}
    >
      {showHeader && (
        <span className="text-xs font-bold text-default-500">
          {roleLabel(message.role)}
        </span>
      )}
      {message.content && (
        <div
          className={`flex max-w-[90%] items-end gap-2 ${
            isUser ? "self-end" : "self-start"
          }`}
        >
          <div
            className={`max-w-full rounded-large px-3 py-2 text-sm shadow-sm ${bubbleClasses} [&_p]:my-0 [&_ul]:my-0 [&_ol]:my-0 [&_pre]:my-0 [&_blockquote]:my-0 [&_h1]:mt-0 [&_h2]:mt-0 [&_h3]:mt-0 [&_h4]:mt-0`}
          >
            <AssistantMarkdown content={message.content} />
            {isStreaming && <span className="ml-1 animate-pulse">▍</span>}
          </div>
          {message.createdAt && (
            <span className="text-[10px] font-light leading-none text-default-500 whitespace-nowrap">
              {formatTime(message.createdAt)}
            </span>
          )}
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
  prefill,
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
  const [isGenerating, setIsGenerating] = useState(false);
  // Динамический нижний паддинг, чтобы можно было проскроллить так,
  // чтобы последнее пользовательское сообщение оказалось у верхней границы
  const [dynamicBottomPadding, setDynamicBottomPadding] = useState(0);

  // Ref на скроллируемый контейнер (overflow-y-auto)
  const scrollRef = useRef<HTMLDivElement>(null);
  const sessionRef = useRef<ChatSessionHandle>(null);
  const prefillAppliedRef = useRef(false);
  const latestUserMessageIdRef = useRef<string | null>(null);
  const autoScrollOnOpenRef = useRef(false);
  // Последовательный счётчик для генерации уникальных id стриминговых messages инструментов
  const toolEventSeqRef = useRef(0);

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

    autoScrollOnOpenRef.current = true;
    latestUserMessageIdRef.current = null;
    void loadChat();
  }, [isOpen, loadChat]);

  // Однократно подставляем текст в инпут при первом открытии, если передан prefill
  useEffect(() => {
    if (
      isOpen &&
      prefill &&
      !prefillAppliedRef.current &&
      inputValue.trim() === ""
    ) {
      setInputValue(prefill);
      prefillAppliedRef.current = true;
    }
  }, [isOpen, prefill, inputValue]);

  // После успешного применения префилла очищаем его из URL, чтобы не автозаполняло повторно
  useEffect(() => {
    if (!isOpen) return;
    if (!prefillAppliedRef.current) return;

    try {
      const url = new URL(window.location.href);

      if (!url.searchParams.has("prefill")) return;

      url.searchParams.delete("prefill");

      const newUrl = `${url.pathname}${url.search ? `?${url.searchParams.toString()}` : ""}${url.hash}`;

      window.history.replaceState(null, "", newUrl);
    } catch {
      // no-op
    }
    // Один раз после применения
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, inputValue, prefill]);

  useEffect(() => {
    if (!isOpen && sessionRef.current) {
      sessionRef.current.close();
      sessionRef.current = null;
    }
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen || !hasMessages) return;

    const el = scrollRef.current;

    if (!el) return;

    const handleScroll = () => {
      if (autoScrollOnOpenRef.current) {
        autoScrollOnOpenRef.current = false;
        el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });

        return;
      }

      const latestId = latestUserMessageIdRef.current;

      if (!latestId) {
        return;
      }

      const target = el.querySelector<HTMLElement>(
        `[data-message-id="${latestId}"]`,
      );

      if (!target) {
        latestUserMessageIdRef.current = null;

        return;
      }

      const containerTop = el.getBoundingClientRect().top;
      const targetTop = target.getBoundingClientRect().top;
      const offset = targetTop - containerTop;
      const paddingOffset = 20;
      const desiredTop = Math.max(el.scrollTop + offset - paddingOffset, 0);
      const maxScrollTop = el.scrollHeight - el.clientHeight;

      if (desiredTop <= maxScrollTop) {
        // Можем проскроллить уже сейчас
        el.scrollTo({ top: desiredTop, behavior: "smooth" });
        latestUserMessageIdRef.current = null;

        return;
      }

      // Контента не хватает, чтобы поставить сообщение в самый верх.
      // Увеличим нижний паддинг на недостающую величину, чтобы desiredTop стал достижимым.
      const needPadding = Math.ceil(desiredTop - maxScrollTop);

      if (needPadding > 0) {
        setDynamicBottomPadding((prev) => Math.max(prev, needPadding));
        // После применения паддинга эффект запустится ещё раз и выполнит прокрутку.
      }
    };

    requestAnimationFrame(() => {
      requestAnimationFrame(handleScroll);
    });
  }, [messages, isOpen, hasMessages, dynamicBottomPadding]);

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

      latestUserMessageIdRef.current = userMessage.id ?? null;
      setMessages((prev) => [...prev, userMessage]);
      setInputValue("");
      setStreamingAssistantMessage("");
      setStreamState({ status: "assistant_thinking" });
      setIsStreaming(true);
      setIsGenerating(true);
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
          // Оставляем только общие статусы (например, assistant_thinking)
          setStreamState((prev) => ({ ...prev, status }));
        },
        onToolEvent: (ev) => {
          console.log("Received tool event:", ev);

          const state = ev.state ?? ToolEventState.STATE_UNSPECIFIED;
          const toolName = (ev.toolName ?? "").trim();

          if (state === ToolEventState.INVOKING) {
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

              return "";
            });

            // Создаём стриминговый чип инструмента как TOOL-сообщение с toolName
            setStreamingToolMessage((prev) => {
              if (prev && prev.toolName && prev.toolName !== toolName) {
                // Завершаем предыдущее
                setMessages((msgs) => [...msgs, prev]);
              }

              if (prev && prev.toolName === toolName) {
                return prev;
              }

              // Стабильный id на время стрима конкретного инструмента
              return {
                id: `streaming-tool-${toolName || "unknown"}-${toolEventSeqRef.current++}`,
                chatId: chat?.id,
                role: WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_TOOL,
                content: "",
                toolName,
                createdAt: new Date().toISOString(),
              };
            });

            return;
          }

          if (
            state === ToolEventState.COMPLETED ||
            state === ToolEventState.ERROR
          ) {
            // Завершаем стриминговый чип инструмента; если ERROR — пометим ошибку без вывода текста
            setStreamingToolMessage((prev) => {
              if (prev && (!toolName || prev.toolName === toolName)) {
                const finalized =
                  state === ToolEventState.ERROR
                    ? { ...prev, error: prev.error ?? "error" }
                    : prev;

                setMessages((msgs) => [...msgs, finalized]);

                return null;
              }

              return prev;
            });

            return;
          }
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
          setIsGenerating(false);
        },
        onError: (err) => {
          console.error("Workout chat stream error", err);
          setError(err.message);
          setIsStreaming(false);
          setIsGenerating(false);
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
          setIsGenerating(false);
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

  const combinedMessages = useMemo(() => {
    // Объединяем обычные и стриминговые сообщения, сохраняя порядок
    const list: WorkoutChatMessage[] = [...messages];

    if (streamingToolMessage) {
      list.push(streamingToolMessage);
    }

    if (streamingMessage) {
      list.push(streamingMessage);
    }

    return list;
  }, [messages, streamingMessage, streamingToolMessage]);

  function senderKey(role?: WorkoutChatMessageRole | null) {
    switch (role) {
      case WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_USER:
        return "user";
      case WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_ASSISTANT:
        return "assistant";
      case WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_TOOL:
        // Считаем tool частью ассистента для группировки заголовков
        return "assistant";
      default:
        return "other";
    }
  }

  const content = useMemo(() => {
    if (loading) {
      return (
        <div className="flex w-full flex-col items-center justify-center gap-2">
          <Spinner color="secondary" size="lg" />
          <p className="text-sm text-default-500">Загружаем чат…</p>
        </div>
      );
    }

    if (error && !hasMessages) {
      return (
        <div className="flex w-full flex-col items-center justify-center gap-4">
          <p className="text-sm text-default-500">{error}</p>
          <Button color="secondary" size="sm" onPress={() => void loadChat()}>
            Повторить
          </Button>
        </div>
      );
    }

    if (!hasMessages) {
      return (
        <div className="flex w-full flex-col items-center justify-center gap-3 p-6 text-center text-default-500">
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
      <div
        className={`flex w-full flex-col gap-3 px-4 ${isGenerating ? "pt-6" : "p-4"}`}
        style={
          isGenerating ? { paddingBottom: dynamicBottomPadding } : undefined
        }
      >
        {combinedMessages.map((message, idx, arr) => {
          // Пропускаем системные сообщения
          const isSystemMessage =
            message.role === WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_SYSTEM;

          if (isSystemMessage) {
            return null;
          }

          const prev = idx > 0 ? arr[idx - 1] : undefined;
          const showHeader =
            !prev || senderKey(prev.role) !== senderKey(message.role);

          const isStreamingItem =
            message.id === "streaming" ||
            streamingToolMessage?.id === message.id;

          const reactKey = message.id
            ? `${message.id}${isStreamingItem ? "-stream" : ""}`
            : `${senderKey(message.role)}-${idx}${isStreamingItem ? "-stream" : ""}`;

          return (
            <MessageBubble
              key={reactKey}
              isStreaming={isStreamingItem}
              message={message}
              showHeader={showHeader}
            />
          );
        })}
      </div>
    );
  }, [
    error,
    hasMessages,
    loadChat,
    loading,
    combinedMessages,
    streamingToolMessage,
    isGenerating,
    dynamicBottomPadding,
  ]);

  const canSend = inputValue.trim().length > 0 && !isStreaming;
  const showThinking = streamState.status === "assistant_thinking";

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
            <div className="h-full relative flex flex-1 flex-col overflow-hidden">
              <ScrollShadow
                ref={scrollRef}
                className="flex-1 overflow-y-auto h-full"
                size={60}
              >
                {/* Контент сообщений; добавляем нижний паддинг, чтобы оверлей не перекрывал последние сообщения */}
                <div
                  className={`flex min-h-full min-w-full ${showThinking ? "pb-4" : ""}`}
                >
                  {content}
                </div>
              </ScrollShadow>

              {/* Статус «Думает…» как оверлей поверх контента внизу */}
              {showThinking && (
                <div className="absolute inset-x-0 bottom-0 z-20 px-4 pb-2 shadow-md">
                  <div
                    aria-live="polite"
                    className="flex items-center gap-1 text-xs text-default-400"
                    role="status"
                  >
                    <Spinner
                      classNames={{ wrapper: "w-3 h-3" }}
                      color="secondary"
                      size="sm"
                    />
                    <span>Думает…</span>
                    {streamState.usageTokens ? (
                      <span className="ml-2 text-[10px] opacity-70">
                        Токенов: {streamState.usageTokens}
                      </span>
                    ) : null}
                  </div>
                </div>
              )}

              {error && hasMessages && (
                <div className="absolute inset-x-0 bottom-0 z-20 px-4 pb-2 text-xs text-danger">
                  {error}
                </div>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2 border-t border-default-200 px-4 py-2 mb-4">
            <Textarea
              autoFocus
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
      </DrawerContent>
    </Drawer>
  );
}
