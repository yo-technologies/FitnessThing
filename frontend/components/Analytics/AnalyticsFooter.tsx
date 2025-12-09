"use client";

import React from "react";
import { Button, ButtonGroup, Checkbox } from "@heroui/react";
import { WorkoutAnalyticsSeries } from "@/api/api.generated";
import { translateMuscleGroup } from "@/config/muscle-groups";

interface AnalyticsFooterProps {
  dateRangePreset: number | null;
  onPresetChange: (months: number) => void;
  series: WorkoutAnalyticsSeries[];
  hiddenSeries: Set<string>;
  onToggleSeries: (name: string) => void;
}

export const AnalyticsFooter: React.FC<AnalyticsFooterProps> = ({
  dateRangePreset,
  onPresetChange,
  series,
  hiddenSeries,
  onToggleSeries,
}) => {
  const colors = [
    "#8884d8",
    "#82ca9d",
    "#ffc658",
    "#ff7300",
    "#0088fe",
    "#00c49f",
    "#ffbb28",
    "#ff8042",
  ];

  return (
    <div className="flex flex-col gap-4 mt-1">
      <div className="flex flex-wrap items-center gap-3">
        <ButtonGroup>
          <Button
            color={dateRangePreset === 1 ? "primary" : "default"}
            onPress={() => onPresetChange(1)}
            size="sm"
          >
            1 мес
          </Button>
          <Button
            color={dateRangePreset === 6 ? "primary" : "default"}
            onPress={() => onPresetChange(6)}
            size="sm"
          >
            6 мес
          </Button>
          <Button
            color={dateRangePreset === 12 ? "primary" : "default"}
            onPress={() => onPresetChange(12)}
            size="sm"
          >
            12 мес
          </Button>
        </ButtonGroup>
      </div>

      {series.length > 0 && (
        <div className="flex flex-col gap-2">
          {series.map((s, index) => (
            <div key={s.name} className="flex items-center gap-2">
              <Checkbox
                isSelected={!hiddenSeries.has(s.name!)}
                onValueChange={() => onToggleSeries(s.name!)}
                color="primary"
                size="sm"
              >
                <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full text-xs" style={{ backgroundColor: colors[index % colors.length] }} />
                    <span>{translateMuscleGroup(s.name)}</span>
                </div>
              </Checkbox>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
