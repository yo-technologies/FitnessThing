import React from "react";
import { Card, CardBody } from "@nextui-org/card";
import { Chip } from "@nextui-org/chip";

import { ChipSelectorProps } from "../types";

import { SectionHeader } from "./SectionHeader";

export function ChipSelector<T>({
  title,
  tooltip,
  options,
  value,
  color,
  onChange,
}: ChipSelectorProps<T>) {
  return (
    <Card>
      <CardBody className="flex flex-col gap-4">
        <SectionHeader title={title} tooltip={tooltip} />
        <div className="flex flex-wrap gap-2">
          {options.map((option) => (
            <Chip
              key={String(option.key)}
              className="cursor-pointer transition-all"
              color={value === option.key ? color : "default"}
              variant={value === option.key ? "solid" : "bordered"}
              onClick={() => onChange(option.key)}
            >
              {option.label}
            </Chip>
          ))}
        </div>
      </CardBody>
    </Card>
  );
}
