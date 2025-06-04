import { retrieveRawInitData } from "@telegram-apps/sdk-react";

import { Api, ApiConfig } from "./api.generated";

export const errUnauthorized = new Error("Unauthorized");

class TelegramApi<T = any> extends Api<T> {
  config: ApiConfig;

  constructor(config: ApiConfig = {}) {
    super(config);
    this.config = config;

    // Добавляем interceptor для запросов с Telegram данными
    this.instance.interceptors.request.use(
      (config) => {
        try {
          const initDataRaw = retrieveRawInitData();

          if (initDataRaw) {
            config.headers["Authorization"] = `tma ${initDataRaw}`;
          }
        } catch (error) {
          console.warn("Failed to retrieve Telegram init data:", error);
        }

        return config;
      },
      (error) => Promise.reject(error),
    );

    // Добавляем interceptor для ответов
    this.instance.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          console.log("Unauthorized request - invalid Telegram data");

          return Promise.reject(errUnauthorized);
        }

        return Promise.reject(error);
      },
    );
  }
}

export const authApi = new TelegramApi({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "/api",
});
