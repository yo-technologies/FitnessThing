"use client";

import type { ThemeProviderProps } from "next-themes";

import * as React from "react";
import { HeroUIProvider } from "@heroui/system";
import { useRouter } from "next/navigation";
import { ThemeProvider as NextThemesProvider } from "next-themes";
import { init, swipeBehavior, viewport } from "@tma.js/sdk-react";
import { I18nProvider } from "@react-aria/i18n";

export interface ProvidersProps {
  children: React.ReactNode;
  themeProps?: ThemeProviderProps;
}

declare module "@react-types/shared" {
  interface RouterConfig {
    routerOptions: NonNullable<
      Parameters<ReturnType<typeof useRouter>["push"]>[1]
    >;
  }
}

function TelegramMiniAppProvider({ children }: ProvidersProps) {
  function inner() {
    try {
      init();
      swipeBehavior.mount();
      swipeBehavior.disableVertical.ifAvailable();
      
      viewport.expand.ifAvailable();  
    } catch (e) {
      console.log("Error while initializing Telegram Mini App SDK");
      console.log(e);
    }
  }

  React.useEffect(inner, []);

  return <>{children}</>;
}

export function Providers({ children, themeProps }: ProvidersProps) {
  const router = useRouter();

  return (
    <HeroUIProvider
      className="flex flex-col flex-grow h-full"
      navigate={router.push}
    >
      <NextThemesProvider {...themeProps}>
        <TelegramMiniAppProvider>
          <I18nProvider locale="ru">{children}</I18nProvider>
        </TelegramMiniAppProvider>
      </NextThemesProvider>
    </HeroUIProvider>
  );
}
