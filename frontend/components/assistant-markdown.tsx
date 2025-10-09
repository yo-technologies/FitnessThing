"use client";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

export type AssistantMarkdownProps = {
  content?: string | null;
};

// Нормализуем случаи вида "- 1) пункт" → "1. пункт",
// чтобы не рендерились одновременно маркеры и нумерация.
function normalizeAssistantMarkdown(input?: string | null): string {
  if (!input) return "";
  // Заменяем в начале строк шаблон "- 1) " → "1. " (и для любых чисел)

  return input.replace(/(^|\n)[\t ]*-\s*(\d+)\)\s/g, "$1$2. ");
}

export function AssistantMarkdown({ content }: AssistantMarkdownProps) {
  return (
    <Markdown
      components={{
        p: (props) => <p className="my-1 leading-relaxed" {...props} />,
        ul: (props) => (
          <ul
            className="list-disc list-outside ml-4 my-2 space-y-1"
            {...props}
          />
        ),
        ol: (props) => (
          <ol
            className="list-decimal list-outside ml-4 my-2 space-y-1"
            {...props}
          />
        ),
        li: (props) => <li className="my-0.5 [&>p]:m-0" {...props} />,
        h1: ({ children, ...props }) => (
          <h1 className="mt-2 mb-1 font-semibold text-base" {...props}>
            {children}
          </h1>
        ),
        h2: ({ children, ...props }) => (
          <h2 className="mt-2 mb-1 font-semibold text-base" {...props}>
            {children}
          </h2>
        ),
        h3: ({ children, ...props }) => (
          <h3 className="mt-2 mb-1 font-semibold text-sm" {...props}>
            {children}
          </h3>
        ),
      }}
      remarkPlugins={[remarkGfm]}
    >
      {normalizeAssistantMarkdown(content)}
    </Markdown>
  );
}
