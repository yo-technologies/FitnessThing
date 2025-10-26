"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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
  WorkoutGetLLMLimitsResponse,
} from "@/api/api.generated";
import { authApi } from "@/api/api";
import {
  WorkoutChatSendRequest,
  WorkoutChatStreamCallbacks,
  sendWorkoutChatMessageStream,
  ChatStreamError,
} from "@/api/chat-stream";
import { GearIcon, UpArrowIcon } from "@/config/icons";
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

type LimitsState = {
  loading: boolean;
  data?: WorkoutGetLLMLimitsResponse;
  error?: string | null;
  cooldownUntil?: number; // epoch ms until retry allowed
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
  ref,
}: {
  message: WorkoutChatMessage;
  isStreaming?: boolean;
  showHeader?: boolean;
  ref?: React.Ref<HTMLDivElement>;
}) {
  const isUser = message.role === WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_USER;
  const isTool = message.role === WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_TOOL;
  const isSystem =
    message.role === WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_SYSTEM;
  const bubbleWidth = isUser ? "max-w-[90%]" : "max-w-full";
  const alignment = isUser ? "items-end" : "items-start";
  const bubbleClasses = isUser
    ? "bg-primary text-primary-foreground px-3 py-2"
    : "text-default-800";

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
      ref={ref}
      className={`flex w-full flex-col gap-2 ${alignment}`}
      data-message-id={message.id ?? undefined}
    >
      {showHeader && (
        <span className="text-xs font-bold text-default-500">
          {roleLabel(message.role)}
        </span>
      )}
      {message.content && (
        <div className={`flex items-end gap-2 ${bubbleWidth}`}>
          <div
            className={`max-w-full rounded-large text-sm shadow-sm ${bubbleClasses}`}
          >
            <AssistantMarkdown content={message.content} />
            {isStreaming && <span className="ml-1 animate-pulse">▍</span>}
          </div>
          {message.createdAt && isUser && (
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
  // Список завершённых tool-сообщений, которые были созданы во время стриминга
  const [streamingCompletedTools, setStreamingCompletedTools] = useState<
    WorkoutChatMessage[]
  >([]);
  const [isGenerating, setIsGenerating] = useState(false);
  const [limits, setLimits] = useState<LimitsState>({ loading: false });
  // Показывать кнопку быстро перейти вниз
  const [showScrollToBottom, setShowScrollToBottom] = useState(false);
  // Применять ли нижний паддинг к контейнеру сообщений (при стриминге)
  const [shouldApplyPadding, setShouldApplyPadding] = useState(false);
  // Размер паддинга в px (будет пересчитываться на основе размера контейнера)
  const [paddingSize, setPaddingSize] = useState(0);

  // Ref на скроллируемый контейнер (overflow-y-auto)
  const scrollRef = useRef<HTMLDivElement>(null);
  const scrollElementId = useMemo(
    () => `workout-chat-scroll-${workoutId}`,
    [workoutId],
  );
  const sessionRef = useRef<ChatSessionHandle>(null);
  const prefillAppliedRef = useRef(false);
  // Последовательный счётчик для генерации уникальных id стриминговых messages инструментов
  const toolEventSeqRef = useRef(0);
  // Ref на последнее сообщение пользователя для скролла к нему
  const lastUserMessageRef = useRef<HTMLDivElement>(null);
  // Ref на textarea для скрытия клавиатуры
  const textareaRef = useRef<HTMLTextAreaElement>(null);

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

  const loadLimits = useCallback(async () => {
    setLimits((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const response = await authApi.v1.chatServiceGetLlmLimits();

      setLimits({ loading: false, data: response.data });
    } catch (err) {
      console.error("Failed to load LLM limits", err);
      setLimits((prev) => ({
        ...prev,
        loading: false,
        error: "Не удалось получить лимиты",
      }));
    }
  }, []);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    void loadChat();
    void loadLimits();
  }, [isOpen, loadChat, loadLimits]);

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

  // Закрытие сессии при закрытии панели
  useEffect(() => {
    if (!isOpen && sessionRef.current) {
      sessionRef.current.close();
      sessionRef.current = null;
    }
  }, [isOpen]);

  // Следим за скроллом контейнера и показываем кнопку «вниз», если пользователь отскроллил
  useEffect(() => {
    const el = document.getElementById(
      scrollElementId,
    ) as HTMLDivElement | null;

    if (!isOpen || !el) {
      setShowScrollToBottom(false);

      return;
    }

    const threshold = 80; // px от низа, при которых считаем что не внизу

    const handleScroll = () => {
      try {
        const atBottom =
          el.scrollHeight - el.scrollTop - el.clientHeight <= threshold;

        setShowScrollToBottom(!atBottom);
      } catch {
        // no-op
      }
    };

    el.addEventListener("scroll", handleScroll, { passive: true });

    // начальная проверка
    handleScroll();

    return () => {
      el.removeEventListener("scroll", handleScroll);
      setShowScrollToBottom(false);
    };
  }, [isOpen, messages]);

  // Прокручиваем к концу при загрузке чата (один раз)
  useEffect(() => {
    if (!isOpen || loading) {
      return;
    }

    // Даём браузеру время на рендер
    const timer = setTimeout(() => {
      const el = document.getElementById(
        scrollElementId,
      ) as HTMLDivElement | null;

      if (!el) return;

      el.scrollTo({ top: el.scrollHeight, behavior: "instant" });
    }, 0);

    return () => clearTimeout(timer);
  }, [isOpen, loading, scrollElementId]);

  // Скролл к последнему сообщению пользователя один раз при начале стриминга
  useEffect(() => {
    if (!shouldApplyPadding || !isOpen || !lastUserMessageRef.current) {
      return;
    }

    // Даём React время на рендер нового сообщения в DOM
    const timer = requestAnimationFrame(() => {
      if (lastUserMessageRef.current) {
        lastUserMessageRef.current.scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      }
    });

    return () => {
      if (typeof timer === "number") {
        cancelAnimationFrame(timer);
      }
    };
  }, [shouldApplyPadding, isOpen]);

  // Динамический паддинг во время генерации
  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const containerEl = document.getElementById(
      scrollElementId,
    ) as HTMLDivElement | null;

    if (!containerEl || !lastUserMessageRef.current) return;

    const updatePadding = () => {
      const containerHeight = containerEl.clientHeight;
      const userMessageEl = lastUserMessageRef.current;

      if (!userMessageEl) return;

      // Если паддинг должен быть применён, вычисляем его на основе сообщений
      if (shouldApplyPadding) {
        // Считаем высоту всех сообщений после пользовательского
        let contentHeightAfterUserMsg = 0;
        let foundUserMsg = false;

        const messageElements =
          containerEl.querySelectorAll("[data-message-id]");

        messageElements.forEach((el) => {
          if (foundUserMsg) {
            contentHeightAfterUserMsg += (el as HTMLElement).offsetHeight;
          }

          // Проверяем, это ли наше пользовательское сообщение
          if (el === userMessageEl) {
            foundUserMsg = true;
          }
        });

        // Паддинг = оставшееся место в контейнере - высота содержимого после пользовательского сообщения
        const availableSpace =
          containerHeight - (userMessageEl.offsetHeight + 12); // 12px - gap между сообщениями
        const newPadding = Math.max(
          availableSpace - contentHeightAfterUserMsg - 25, // дополнительный gap
          0,
        );

        setPaddingSize(newPadding);
      }
    };

    // Первое вычисление сразу
    updatePadding();

    // Слушаем изменения размера и обновляем паддинг во время генерации
    const observer = new MutationObserver(updatePadding);

    observer.observe(containerEl, {
      childList: true,
      subtree: true,
      characterData: true,
    });

    return () => {
      observer.disconnect();
    };
  }, [shouldApplyPadding, isOpen, scrollElementId]);

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
      setStreamingToolMessage(null);
      setStreamingCompletedTools([]); // Очищаем предыдущие временные tool-сообщения
      setStreamState({ status: "assistant_thinking" });
      setIsStreaming(true);
      setIsGenerating(true);
      setShouldApplyPadding(true);
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
                // Завершаем предыдущее, добавляем в список завершённых
                setStreamingCompletedTools((tools) => [...tools, prev]);
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

                // Добавляем в список завершённых tool-сообщений, НЕ в основной messages
                setStreamingCompletedTools((tools) => [...tools, finalized]);

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
          setStreamingCompletedTools([]); // Очищаем временные tool-сообщения
          setStreamState({});
          setIsStreaming(false);
          setIsGenerating(false);
          void loadLimits();
        },
        onError: (err) => {
          console.error("Workout chat stream error", err);
          if (err instanceof ChatStreamError) {
            if (err.type === "rate_limit" || err.code === "429") {
              const retryAfterSec = err.retryAfterSeconds ?? 0;
              const until = Date.now() + Math.max(retryAfterSec, 0) * 1000;

              setLimits((prev) => ({ ...prev, cooldownUntil: until }));

              setError(
                retryAfterSec > 0
                  ? `Превышен лимит запросов. Попробуйте через ${retryAfterSec} сек.`
                  : "Превышен лимит запросов. Попробуйте позже.",
              );
            } else if (err.type === "quota_exceeded") {
              setError("Исчерпан дневной лимит токенов. Попробуйте завтра.");
            } else {
              setError(err.message);
            }
          } else {
            setError(err.message);
          }
          setIsStreaming(false);
          setIsGenerating(false);
          setStreamingAssistantMessage("");
          setStreamingToolMessage(null);
          setStreamingCompletedTools([]); // Очищаем временные tool-сообщения
          // Перезагружаем чат, чтобы получить корректное состояние с сервера
          void loadChat();
          void loadLimits();
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
          setStreamingToolMessage(null);
          setStreamingCompletedTools([]); // Очищаем временные tool-сообщения
          setStreamState({});
          sessionRef.current = null;
        });
    },
    [chat?.id, isStreaming, workoutId, loadChat, loadLimits],
  );

  const handleSend = useCallback(() => {
    // Скрываем клавиатуру перед отправкой сообщения
    textareaRef.current?.blur();
    void sendMessage(inputValue);
  }, [inputValue, sendMessage]);

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

    // Добавляем завершённые tool-сообщения из стриминга
    streamingCompletedTools.forEach((tool) => {
      list.push(tool);
    });

    if (streamingToolMessage) {
      list.push(streamingToolMessage);
    }

    if (streamingMessage) {
      list.push(streamingMessage);
    }

    // Сортируем по времени создания для корректного порядка
    return list.sort((a, b) => {
      const timeA = a.createdAt ? new Date(a.createdAt).getTime() : 0;
      const timeB = b.createdAt ? new Date(b.createdAt).getTime() : 0;

      return timeA - timeB;
    });
  }, [
    messages,
    streamingMessage,
    streamingToolMessage,
    streamingCompletedTools,
  ]);

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
        style={{
          paddingBottom: shouldApplyPadding ? `${paddingSize}px` : undefined,
        }}
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

          // Проверяем, является ли это последним сообщением пользователя
          const isUser =
            message.role === WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_USER;
          const isLastUserMessage =
            isUser &&
            idx ===
              arr.reduceRight((lastIdx, msg, i) => {
                if (lastIdx !== -1) return lastIdx;

                return msg.role ===
                  WorkoutChatMessageRole.CHAT_MESSAGE_ROLE_USER
                  ? i
                  : -1;
              }, -1);

          return (
            <MessageBubble
              key={reactKey}
              ref={isLastUserMessage ? lastUserMessageRef : undefined}
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
    shouldApplyPadding,
    paddingSize,
  ]);

  const now = Date.now();
  const inCooldown = Boolean(
    limits.cooldownUntil && now < (limits.cooldownUntil ?? 0),
  );
  const outOfQuota = Boolean(
    limits.data &&
      typeof limits.data.remaining === "number" &&
      (limits.data.remaining ?? 0) <= 0,
  );
  const canSend =
    inputValue.trim().length > 0 && !isStreaming && !inCooldown && !outOfQuota;

  const sendTooltip = inCooldown
    ? "Подождите окончания кулдауна"
    : outOfQuota
      ? "Исчерпан дневной лимит"
      : undefined;
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
                id={scrollElementId}
                size={60}
              >
                {/* Контент сообщений; добавляем нижний паддинг, чтобы оверлей не перекрывал последние сообщения */}
                <div className="flex min-h-full min-w-full pb-4">{content}</div>
              </ScrollShadow>

              {/* Кнопка быстро вниз */}
              {showScrollToBottom && (
                <div className="absolute right-4 bottom-2 z-30">
                  <Button
                    isIconOnly
                    aria-label="Прокрутить вниз"
                    color="secondary"
                    radius="full"
                    size="sm"
                    variant="flat"
                    onPress={() => {
                      const el = document.getElementById(
                        scrollElementId,
                      ) as HTMLDivElement | null;

                      if (!el) return;

                      el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
                    }}
                  >
                    <UpArrowIcon className="h-4 w-4 rotate-180" />
                  </Button>
                </div>
              )}

              {(error || inCooldown || outOfQuota || showThinking) && (
                <div className="absolute inset-x-0 bottom-0 z-20 px-4 pb-2 text-xs shadow-md">
                  {showThinking && (
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
                    </div>
                  )}
                  {error && <div className="text-danger mb-1">{error}</div>}
                  {inCooldown && (
                    <div className="text-warning">
                      {"Слишком часто. Подождите немного перед следующим "}
                      {"запросом."}
                    </div>
                  )}
                  {outOfQuota && (
                    <div className="text-warning">
                      Исчерпан дневной лимит. Продолжите завтра.
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2 border-t border-default-200 px-4 py-2 mb-4">
            <Textarea
              ref={textareaRef}
              className="flex-1"
              classNames={{
                inputWrapper: "bg-default-100",
              }}
              minRows={2}
              placeholder="Напишите сообщение..."
              value={inputValue}
              onChange={(event) => setInputValue(event.target.value)}
            />
            <Button
              isIconOnly
              aria-label="Отправить сообщение"
              className="shrink-0"
              color="secondary"
              isDisabled={!canSend}
              isLoading={isStreaming}
              radius="full"
              title={sendTooltip}
              onPress={handleSend}
            >
              <UpArrowIcon className="h-6 w-6" />
            </Button>
          </div>
        </div>
      </DrawerContent>
    </Drawer>
  );
}
