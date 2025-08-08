"use client";
import { useEffect, useState } from "react";
import { link as linkStyles } from "@nextui-org/theme";
import NextLink from "next/link";
import clsx from "clsx";

import { siteConfig } from "@/config/site";

export const Navbar = () => {
  const [keyboardOpen, setKeyboardOpen] = useState(false);

  useEffect(() => {
    const baseline =
      typeof window !== "undefined"
        ? window.visualViewport?.height || window.innerHeight
        : 0;

    const onVVResize = () => {
      const current = window.visualViewport?.height || window.innerHeight;
      // Если высота заметно уменьшилась, считаем что открылась клавиатура
      const isOpen = baseline - current > 150;

      setKeyboardOpen(isOpen);
      document.documentElement.classList.toggle("kb-open", isOpen);
    };

    const onFocusIn = (e: FocusEvent) => {
      const target = e.target as HTMLElement | null;

      if (!target) return;
      if (/(input|textarea|select)/i.test(target.tagName)) {
        setKeyboardOpen(true);
        document.documentElement.classList.add("kb-open");
      }
    };

    const onFocusOut = () => {
      // Небольшая задержка, чтобы не мигало при переходе фокуса
      setTimeout(() => {
        setKeyboardOpen(false);
        document.documentElement.classList.remove("kb-open");
      }, 100);
    };

    window.visualViewport?.addEventListener("resize", onVVResize);
    window.addEventListener("focusin", onFocusIn);
    window.addEventListener("focusout", onFocusOut);

    return () => {
      window.visualViewport?.removeEventListener("resize", onVVResize);
      window.removeEventListener("focusin", onFocusIn);
      window.removeEventListener("focusout", onFocusOut);
    };
  }, []);

  return (
    <div
      className={clsx(
        "app-navbar flex flex-col items-start justify-between fixed bottom-0 left-0 w-full bg-background h-[4.5rem] p-3 transition-transform duration-200 will-change-transform",
        // Прячем навбар когда открыта клавиатура
        keyboardOpen && "translate-y-full pointer-events-none",
        // Учет safe-area снизу
        "pb-[max(env(safe-area-inset-bottom),0.75rem)]",
      )}
    >
      <div className="flex items-center justify-around w-full">
        {siteConfig.navItems.map((item, id) => (
          <NextLink
            key={id}
            className={clsx(
              linkStyles({ color: "foreground" }),
              "data-[active=true]:text-primary data-[active=true]:font-medium",
            )}
            color="foreground"
            href={item.href}
          >
            <div className="flex flex-col items-center justify-center gap-1">
              {item.icon}
              <p className="text-xs">{item.label}</p>
            </div>
          </NextLink>
        ))}
      </div>
    </div>
  );
};
