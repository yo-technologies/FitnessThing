/* eslint-disable @next/next/no-img-element */
import "@/styles/globals.css";
import { Metadata, Viewport } from "next";
import clsx from "clsx";
import { ToastContainer } from "react-toastify";
import Script from "next/script";
import { Suspense } from "react";

import { Providers } from "./providers";

import { siteConfig } from "@/config/site";
import { fontSans } from "@/config/fonts";
import { Navbar } from "@/components/navbar";
import YandexMetrika from "@/components/yandex-metrica";

export const metadata: Metadata = {
  title: {
    default: siteConfig.name,
    template: `%s - ${siteConfig.name}`,
  },
  description: siteConfig.description,
  icons: {
    icon: "/favicon.ico",
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "white" },
    { media: "(prefers-color-scheme: dark)", color: "black" },
  ],
  userScalable: false,
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html suppressHydrationWarning className="overflow-hidden" lang="ru">
      <head>
        <Script id="metrika-counter" strategy="afterInteractive">
          {`(function(m,e,t,r,i,k,a){m[i]=m[i]||function(){(m[i].a=m[i].a||[]).push(arguments)};
            m[i].l=1*new Date();
            for (var j = 0; j < document.scripts.length; j++) {if (document.scripts[j].src === r) { return; }}
            k=e.createElement(t),a=e.getElementsByTagName(t)[0],k.async=1,k.src=r,a.parentNode.insertBefore(k,a)})
            (window, document, "script", "https://mc.yandex.ru/metrika/tag.js", "ym");

            ym(99867208, "init", {
                  defer:true,
                  clickmap:true,
                  trackLinks:true,
                  accurateTrackBounce:true,
                  webvisor:true
            });
          `}
        </Script>
      </head>
      <body
        className={clsx(
          "h-full w-full overflow-hidden font-sans antialiased ",
          fontSans.variable,
        )}
      >
        <Suspense fallback={<></>}>
          <YandexMetrika />
        </Suspense>
        <div className="flex flex-col justify-between h-dvh overflow-y-scroll max-w-full">
          <Providers
            themeProps={{
              attribute: "class",
              defaultTheme: "dark",
              enableSystem: true,
            }}
          >
            <main
              className={clsx(
                "app-main flex mx-auto flex-grow overflow-y-auto w-full h-full",
                // Отступ под navbar; убираем при открытой клавиатуре (см. globals.css)
                "mb-[max(4.5rem,env(safe-area-inset-bottom))]",
              )}
            >
              <div className="flex flex-grow max-h-full flex-col w-full">
                {children}
              </div>
              <ToastContainer />
              <Navbar />
            </main>
          </Providers>
        </div>
      </body>
    </html>
  );
}
