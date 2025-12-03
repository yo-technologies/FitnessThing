import React, { useState, useRef, useEffect } from "react";
import { Card, CardBody } from "@heroui/card";
import { Chip } from "@heroui/chip";
import clsx from "clsx";

import { SectionHeader } from "./SectionHeader";

export interface EditableChipFieldProps {
  title: string;
  tooltip: string;
  value: string[];
  placeholder: string;
  color: "primary" | "secondary" | "success" | "warning" | "danger" | "default";
  onChange: (value: string[]) => void;
}

export function EditableChipField({
  title,
  tooltip,
  value,
  placeholder,
  color,
  onChange,
}: EditableChipFieldProps) {
  const [inputValue, setInputValue] = useState("");
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Автофокус при начале редактирования
  useEffect(() => {
    if (editingIndex !== null && textareaRef.current) {
      textareaRef.current.focus();
    }
  }, [editingIndex]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      addChip();
    } else if (e.key === "Backspace" && inputValue === "" && value.length > 0) {
      e.preventDefault();
      removeChip(value.length - 1);
    }
  };

  const addChip = () => {
    const trimmedValue = inputValue.trim();

    if (trimmedValue) {
      if (editingIndex !== null) {
        // Редактируем существующий чип
        const newValue = [...value];

        newValue[editingIndex] = trimmedValue;
        onChange(newValue);
        setEditingIndex(null);
      } else {
        // Добавляем новый чип
        onChange([...value, trimmedValue]);
      }
      setInputValue("");

      // Возвращаем фокус на textarea для продолжения ввода
      setTimeout(() => {
        if (textareaRef.current) {
          textareaRef.current.focus();
        }
      }, 0);
    }
  };

  const removeChip = (index: number) => {
    const newValue = value.filter((_, i) => i !== index);

    onChange(newValue);
  };

  const startEditing = (index: number) => {
    setEditingIndex(index);
    setInputValue(value[index]);
  };

  const cancelEditing = () => {
    setEditingIndex(null);
    setInputValue("");
  };

  return (
    <Card>
      <CardBody className="flex flex-col gap-4">
        <SectionHeader title={title} tooltip={tooltip} />

        <div className="min-h-[100px] p-3 border-2 border-default-200 rounded-lg focus-within:border-primary transition-colors">
          <div className="flex flex-wrap gap-2 items-start min-h-[70px]">
            {value.map((chip, index) =>
              editingIndex === index ? (
                <textarea
                  key={index}
                  ref={textareaRef}
                  className="resize-none outline-none bg-transparent text-xs p-1 border-b border-primary h-fit"
                  rows={1}
                  value={inputValue}
                  onBlur={addChip}
                  onChange={(e) => setInputValue(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && !e.shiftKey) {
                      e.preventDefault();
                      addChip();
                    } else if (e.key === "Escape") {
                      e.preventDefault();
                      cancelEditing();
                    }
                  }}
                />
              ) : (
                <Chip
                  key={index}
                  className="cursor-pointer transition-all hover:scale-105"
                  color={color}
                  size="sm"
                  variant="solid"
                  onClick={() => startEditing(index)}
                  onClose={() => removeChip(index)}
                >
                  {chip}
                </Chip>
              ),
            )}

            {editingIndex === null && (
              <textarea
                ref={textareaRef}
                className={clsx(
                  "flex-1 h-fit resize-none outline-none bg-transparent text-xs placeholder:text-default-400 placeholder:text-xs",
                  value.length > 0 && "leading-6 placeholder:leading-6",
                )}
                placeholder={
                  value.length === 0 ? placeholder : " добавить цель..."
                }
                rows={1}
                style={{
                  minHeight: "50px",
                }}
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                onInput={(e) => {
                  const target = e.target as HTMLTextAreaElement;

                  target.style.height = "auto";
                  target.style.height = target.scrollHeight + "px";
                }}
                onKeyDown={handleKeyDown}
              />
            )}
          </div>
        </div>

        <div className="text-xs text-default-400">
          <p>• Нажмите Enter для добавления цели</p>
          <p>• Кликните на чип для редактирования</p>
          <p>• Backspace в пустом поле удаляет последний чип</p>
        </div>
      </CardBody>
    </Card>
  );
}
