import { retrieveRawInitData } from "@telegram-apps/sdk-react";

import {
  WorkoutChatUsage,
  WorkoutSendChatMessageResponse,
  WorkoutSendChatMessageStreamResponse,
  WorkoutToolEvent,
} from "./api.generated";

export type WorkoutChatStreamCallbacks = {
  onMessageDelta?: (delta: string) => void;
  onUsage?: (usage: WorkoutChatUsage) => void;
  onStatus?: (status: string) => void;
  onToolEvent?: (ev: WorkoutToolEvent) => void;
  onFinal?: (final: WorkoutSendChatMessageResponse) => void;
  onError?: (error: Error) => void;
};

export type WorkoutChatStreamSession = {
  close: () => void;
  done: Promise<WorkoutSendChatMessageResponse | undefined>;
};

export type WorkoutChatSendRequest = {
  chatId?: string;
  workoutId?: string;
  content: string;
};

function resolveWsBaseUrl(): string {
  const configuredApiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (configuredApiUrl) {
    try {
      const url = new URL(configuredApiUrl);
      const isSecure = url.protocol === "https:";

      url.protocol = isSecure ? "wss:" : "ws:";

      const normalizedPath = url.pathname.replace(/\/$/, "");

      url.pathname = `${normalizedPath}/ws/chat`;
      url.search = "";
      url.hash = "";

      return url.toString();
    } catch (error) {
      console.warn(
        "Failed to parse NEXT_PUBLIC_API_URL as absolute URL",
        error,
      );
    }
  }

  if (typeof window === "undefined") {
    throw new Error("Cannot resolve WebSocket URL on the server");
  }

  const origin = window.location.origin;
  const isSecure = origin.startsWith("https:");
  const baseApiPath = (configuredApiUrl ?? "/api").replace(/\/$/, "");

  const wsProtocol = isSecure ? "wss" : "ws";
  const path = `${baseApiPath}/ws/chat`.replace(/\/\//g, "/");

  return `${wsProtocol}://${window.location.host}${path.startsWith("/") ? path : `/${path}`}`;
}

function buildWebSocketUrl(): string {
  const baseUrl = resolveWsBaseUrl();
  const initDataRaw = retrieveRawInitData();

  if (!initDataRaw) {
    return baseUrl;
  }

  try {
    const url = new URL(baseUrl);

    url.searchParams.set("init_data", initDataRaw);

    return url.toString();
  } catch (error) {
    console.warn("Failed to append init_data to websocket URL", error);

    const separator = baseUrl.includes("?") ? "&" : "?";

    return `${baseUrl}${separator}init_data=${encodeURIComponent(initDataRaw)}`;
  }
}

export class ChatStreamError extends Error {
  readonly type?: string;
  readonly code?: string;
  readonly retryAfterSeconds?: number;

  constructor(
    message: string,
    opts?: { type?: string; code?: string; retryAfterSeconds?: number },
  ) {
    super(message);
    this.name = "ChatStreamError";
    this.type = opts?.type;
    this.code = opts?.code;
    this.retryAfterSeconds = opts?.retryAfterSeconds;
  }
}

function isStreamResponse(
  payload: unknown,
): payload is WorkoutSendChatMessageStreamResponse {
  return typeof payload === "object" && payload !== null;
}

function parseStreamMessage(
  data: string,
): WorkoutSendChatMessageStreamResponse | { error: string } | null {
  try {
    const parsed = JSON.parse(data);

    if (parsed && typeof parsed === "object" && "error" in parsed) {
      const errorValue = (parsed as { error: unknown }).error;

      if (typeof errorValue === "string") {
        return { error: errorValue };
      }

      return { error: "Unknown chat stream error" };
    }

    // Handle protobuf oneof structure where payload is nested
    if (parsed && typeof parsed === "object" && "payload" in parsed) {
      const payload = (parsed as { payload: unknown }).payload;

      if (isStreamResponse(payload)) {
        return payload;
      } else {
        console.warn("Payload does not match expected structure:", payload);
      }
    }

    // Handle case where payload is capitalized (Payload instead of payload)
    if (parsed && typeof parsed === "object" && "Payload" in parsed) {
      const payload = (parsed as { Payload: unknown }).Payload;

      if (isStreamResponse(payload)) {
        return payload;
      } else {
        console.warn("Payload does not match expected structure:", payload);
      }
    }
    if (isStreamResponse(parsed)) {
      return parsed;
    }

    return null;
  } catch (error) {
    console.warn("Failed to parse chat stream message", error);

    return null;
  }
}

export function sendWorkoutChatMessageStream(
  request: WorkoutChatSendRequest,
  callbacks: WorkoutChatStreamCallbacks = {},
): WorkoutChatStreamSession {
  const url = buildWebSocketUrl();
  const socket = new WebSocket(url);

  let isClosed = false;
  let finalResponse: WorkoutSendChatMessageResponse | undefined;

  const close = () => {
    if (isClosed) {
      return;
    }

    isClosed = true;

    try {
      socket.close();
    } catch (error) {
      console.warn("Failed to close chat websocket", error);
    }
  };

  const done = new Promise<WorkoutSendChatMessageResponse | undefined>(
    (resolve, reject) => {
      socket.addEventListener("open", () => {
        // Convert camelCase to snake_case for protobuf compatibility
        const protobufRequest = {
          chat_id: request.chatId,
          workout_id: request.workoutId,
          content: request.content,
        };

        socket.send(JSON.stringify(protobufRequest));
      });

      socket.addEventListener("message", (event) => {
        const parsed = parseStreamMessage(event.data);

        if (!parsed) {
          console.warn("Failed to parse message:", event.data);

          return;
        }

        if ("error" in parsed && typeof (parsed as any).error === "string") {
          const error = new Error((parsed as any).error);

          callbacks.onError?.(error);
          reject(error);
          close();

          return;
        }

        const payload = parsed as WorkoutSendChatMessageStreamResponse;

        // Structured error payload from server stream
        if (payload.error) {
          const msg = payload.error.message || "Ошибка в чате";
          const error = new ChatStreamError(msg, {
            type: payload.error.type,
            code: payload.error.code,
            retryAfterSeconds: payload.error.retryAfterSeconds,
          });

          callbacks.onError?.(error);
          reject(error);
          close();

          return;
        }

        if (payload.messageDelta?.content) {
          callbacks.onMessageDelta?.(payload.messageDelta.content);
        }

        if (payload.usage) {
          callbacks.onUsage?.(payload.usage);
        }

        if (payload.status) {
          callbacks.onStatus?.(payload.status);
        }

        if (payload.toolEvent) {
          callbacks.onToolEvent?.(payload.toolEvent);
        }

        if (payload.final) {
          finalResponse = payload.final;
          callbacks.onFinal?.(payload.final);
          resolve(payload.final);
          close();
        }
      });

      socket.addEventListener("error", (event) => {
        const error = new Error("Chat stream connection error");

        console.error("Workout chat websocket error", event);
        callbacks.onError?.(error);
        reject(error);
        close();
      });

      socket.addEventListener("close", () => {
        if (!isClosed) {
          isClosed = true;
          resolve(finalResponse);
        }
      });
    },
  );

  return { close, done };
}
