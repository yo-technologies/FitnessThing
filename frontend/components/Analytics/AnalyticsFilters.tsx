"use client";

import React, { useEffect, useState } from "react";
import { Select, SelectItem, Button, ButtonGroup, Checkbox, Input } from "@heroui/react";
import { authApi } from "@/api/api";
import { WorkoutMuscleGroup, WorkoutExercise } from "@/api/api.generated";
import { subMonths, format, parseISO, startOfDay, endOfDay } from "date-fns";
import { translateMuscleGroup } from "@/config/muscle-groups";

interface AnalyticsFiltersProps {
  onFilterChange: (filters: {
    muscleGroupId?: string;
    exerciseId?: string;
    from: Date;
    to: Date;
    splitByExercise: boolean;
  }) => void;
}

export const AnalyticsFilters: React.FC<AnalyticsFiltersProps> = ({
  onFilterChange,
}) => {
  const [muscleGroups, setMuscleGroups] = useState<WorkoutMuscleGroup[]>([]);
  const [exercises, setExercises] = useState<WorkoutExercise[]>([]);
  
  const [selectedMuscleGroup, setSelectedMuscleGroup] = useState<string>("");
  const [selectedExercise, setSelectedExercise] = useState<string>("");
  const [splitByExercise, setSplitByExercise] = useState<boolean>(false);
  
  const [dateRangePreset, setDateRangePreset] = useState<number | null>(1); // 1, 6, 12 months
  const [customFrom, setCustomFrom] = useState<string>(
    format(subMonths(new Date(), 1), "yyyy-MM-dd")
  );
  const [customTo, setCustomTo] = useState<string>(
    format(new Date(), "yyyy-MM-dd")
  );

  useEffect(() => {
    const fetchData = async () => {
      try {
        const mgRes = await authApi.v1.exerciseServiceGetMuscleGroups();
        if (mgRes.data.muscleGroups) {
          setMuscleGroups(mgRes.data.muscleGroups);
        }
        
        const exRes = await authApi.v1.exerciseServiceGetExercises();
        if (exRes.data.exercises) {
          setExercises(exRes.data.exercises);
        }
      } catch (e) {
        console.error("Failed to fetch filter data", e);
      }
    };
    fetchData();
  }, []);

  // Filter exercises based on selected muscle group
  const filteredExercises = React.useMemo(() => {
    if (!selectedMuscleGroup) return exercises;
    return exercises.filter((ex) =>
      ex.targetMuscleGroups?.includes(selectedMuscleGroup)
    );
  }, [exercises, selectedMuscleGroup]);

  // Handle filter changes
  useEffect(() => {
    let from: Date;
    let to: Date;

    if (dateRangePreset) {
      to = new Date();
      from = subMonths(to, dateRangePreset);
    } else {
      from = parseISO(customFrom);
      to = parseISO(customTo);
    }

    // Ensure valid dates
    if (isNaN(from.getTime())) from = subMonths(new Date(), 1);
    if (isNaN(to.getTime())) to = new Date();

    onFilterChange({
      muscleGroupId: selectedMuscleGroup || undefined,
      exerciseId: selectedExercise || undefined,
      from: startOfDay(from),
      to: endOfDay(to),
      splitByExercise: splitByExercise,
    });
  }, [
    selectedMuscleGroup,
    selectedExercise,
    splitByExercise,
    dateRangePreset,
    customFrom,
    customTo,
  ]);

  const handlePresetChange = (months: number) => {
    setDateRangePreset(months);
    setCustomFrom(format(subMonths(new Date(), months), "yyyy-MM-dd"));
    setCustomTo(format(new Date(), "yyyy-MM-dd"));
  };

  const handleManualDateChange = (type: "from" | "to", value: string) => {
    setDateRangePreset(null);
    if (type === "from") setCustomFrom(value);
    else setCustomTo(value);
  };

  return (
    <div className="flex flex-col gap-4 p-4 bg-content1 rounded-lg shadow-sm">
      <div className="flex flex-wrap gap-4">
        <Select
          label="Группа мышц"
          placeholder="Выберите группу"
          className="max-w-xs"
          selectedKeys={selectedMuscleGroup ? [selectedMuscleGroup] : []}
          onChange={(e) => {
            setSelectedMuscleGroup(e.target.value);
            setSelectedExercise(""); // Reset exercise when muscle group changes
          }}
        >
          {muscleGroups.map((mg) => (
            <SelectItem key={mg.id!}>
              {translateMuscleGroup(mg.name)}
            </SelectItem>
          ))}
        </Select>

        <Select
          label="Упражнение"
          placeholder="Выберите упражнение"
          className="max-w-xs"
          selectedKeys={selectedExercise ? [selectedExercise] : []}
          onChange={(e) => setSelectedExercise(e.target.value)}
          isDisabled={!selectedMuscleGroup && false} // Can select exercise without muscle group? Prompt says "choose muscle group AND specific exercise". Usually exercise implies muscle group.
        >
          {filteredExercises.map((ex) => (
            <SelectItem key={ex.id!}>
              {ex.name}
            </SelectItem>
          ))}
        </Select>
      </div>

      <div className="flex flex-wrap items-center gap-4">
        <ButtonGroup>
          <Button
            color={dateRangePreset === 1 ? "primary" : "default"}
            onPress={() => handlePresetChange(1)}
          >
            1 мес
          </Button>
          <Button
            color={dateRangePreset === 6 ? "primary" : "default"}
            onPress={() => handlePresetChange(6)}
          >
            6 мес
          </Button>
          <Button
            color={dateRangePreset === 12 ? "primary" : "default"}
            onPress={() => handlePresetChange(12)}
          >
            12 мес
          </Button>
        </ButtonGroup>

        <div className="flex gap-2 items-center">
          <Input
            type="date"
            label="От"
            value={customFrom}
            onChange={(e) => handleManualDateChange("from", e.target.value)}
            className="w-40"
          />
          <Input
            type="date"
            label="До"
            value={customTo}
            onChange={(e) => handleManualDateChange("to", e.target.value)}
            className="w-40"
          />
        </div>
      </div>

      {selectedMuscleGroup && !selectedExercise && (
        <div className="flex items-center gap-2">
          <Checkbox
            isSelected={splitByExercise}
            onValueChange={setSplitByExercise}
          >
            Показывать каждое упражнение отдельно
          </Checkbox>
        </div>
      )}
    </div>
  );
};
