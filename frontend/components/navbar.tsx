"use client";
import { useEffect, useState } from "react";
import { link as linkStyles } from "@nextui-org/theme";
import NextLink from "next/link";
import clsx from "clsx";

import { siteConfig } from "@/config/site";

export const Navbar = () => {
  const [keyboardOpen, setKeyboardOpen] = useState(false);

  useEffect(() => {
    // Helper to detect focusable editable elements
    const isEditable = (el: Element | null | undefined) => {
      if (!el) return false;
      const tag = (el as HTMLElement).tagName?.toLowerCase();
      const editable = (el as HTMLElement).getAttribute?.("contenteditable");

      return (
        tag === "input" ||
        tag === "textarea" ||
        tag === "select" ||
        editable === "" ||
        editable === "true"
      );
    };

    // Track whether any editable element is currently focused
    let hasEditableFocused = false;

    const updateUI = (open: boolean) => {
      setKeyboardOpen(open);
      document.documentElement.classList.toggle("kb-open", open);
    };

    const baseline =
      typeof window !== "undefined"
        ? window.visualViewport?.height || window.innerHeight
        : 0;

    const onVVResize = () => {
      const current = window.visualViewport?.height || window.innerHeight;
      // Если высота заметно уменьшилась, считаем что открылась клавиатура
      const keyboardLikelyOpen = baseline - current > 150;

      // Открыто, если либо клавиатура распознана по высоте, либо фокус на поле ввода
      const shouldOpen = keyboardLikelyOpen || hasEditableFocused;

      updateUI(shouldOpen);
    };

    const onFocusIn = (e: FocusEvent) => {
      const target = e.target as HTMLElement | null;

      if (!target) return;
      if (isEditable(target)) {
        hasEditableFocused = true;
        updateUI(true);
      }
    };

    const onFocusOut = (e: FocusEvent) => {
      // Даем время новому элементу получить фокус, чтобы не мигало при переходе
      const related = (e.relatedTarget as Element | null) ?? null;

      setTimeout(() => {
        // Если фокус переместился на другой инпут/textarea/select — считаем, что клавиатура всё еще открыта
        const nextActive =
          related ?? (document.activeElement as Element | null);

        hasEditableFocused = isEditable(nextActive);
        updateUI(hasEditableFocused);
      }, 50);
    };

    window.visualViewport?.addEventListener("resize", onVVResize);
    window.addEventListener("focusin", onFocusIn);
    // useCapture=true поможет поймать событие раньше, но здесь достаточно по умолчанию
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
