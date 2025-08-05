import React from "react";
import { Card, CardBody } from "@nextui-org/card";
import { Chip } from "@nextui-org/chip";
import { Slider } from "@nextui-org/slider";

import { SliderWithMarksProps } from "../types";

import { SectionHeader } from "./SectionHeader";

export function SliderWithMarks({
  title,
  tooltip,
  value,
  minValue,
  maxValue,
  step,
  marks,
  color,
  formatValue,
  onChange,
}: SliderWithMarksProps) {
  // Преобразуем цвет для Chip и Slider
  const chipColor = color === "default" ? "primary" : color;
  const sliderColor = color === "default" ? "foreground" : color;

  return (
    <Card>
      <CardBody className="flex flex-col gap-4">
        <SectionHeader
          badge={
            <Chip color={chipColor} size="sm" variant="flat">
              {formatValue(value)}
            </Chip>
          }
          title={title}
          tooltip={tooltip}
        />
        <Slider
          className="w-full"
          color={sliderColor}
          marks={marks}
          maxValue={maxValue}
          minValue={minValue}
          size="sm"
          step={step}
          value={value}
          onChange={onChange}
        />
      </CardBody>
    </Card>
  );
}
