import React from "react";
import { Card, CardBody } from "@nextui-org/card";
import { Chip } from "@nextui-org/chip";

import { MultiChipSelectorProps } from "../types";

import { SectionHeader } from "./SectionHeader";

export function MultiChipSelector({
  title,
  tooltip,
  options,
  selectedIds,
  color,
  onChange,
}: MultiChipSelectorProps) {
  const handleToggle = (id: string) => {
    const newSelection = selectedIds.includes(id)
      ? selectedIds.filter((selectedId) => selectedId !== id)
      : [...selectedIds, id];

    onChange(newSelection);
  };

  return (
    <Card>
      <CardBody className="flex flex-col gap-2">
        <SectionHeader title={title} tooltip={tooltip} />

        <div className="flex flex-wrap gap-2">
          {options.map((option) => (
            <Chip
              key={option.id}
              className="cursor-pointer transition-all hover:scale-105"
              color={selectedIds.includes(option.id) ? color : "default"}
              variant={selectedIds.includes(option.id) ? "solid" : "bordered"}
              onClick={() => handleToggle(option.id)}
            >
              {option.name}
            </Chip>
          ))}
        </div>

        {options.length === 0 && (
          <div className="text-center py-4 text-default-500">
            <p>Загрузка опций...</p>
          </div>
        )}
      </CardBody>
    </Card>
  );
}
