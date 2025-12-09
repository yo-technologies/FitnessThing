"use client";

import React from "react";
import { Select, SelectItem } from "@heroui/react";
import { WorkoutMuscleGroup, WorkoutExercise } from "@/api/api.generated";
import { translateMuscleGroup } from "@/config/muscle-groups";

interface AnalyticsHeaderProps {
  muscleGroups: WorkoutMuscleGroup[];
  exercises: WorkoutExercise[];
  selectedMuscleGroup: string;
  selectedExercise: string;
  onMuscleGroupChange: (id: string) => void;
  onExerciseChange: (id: string) => void;
}

export const AnalyticsHeader: React.FC<AnalyticsHeaderProps> = ({
  muscleGroups,
  exercises,
  selectedMuscleGroup,
  selectedExercise,
  onMuscleGroupChange,
  onExerciseChange,
}) => {
  return (
    <div className="grid grid-cols-2 gap-3 mb-2 w-full">
      <Select
        label="Группа мышц"
        placeholder="Выберите группу"
        className="w-full"
        selectedKeys={selectedMuscleGroup ? [selectedMuscleGroup] : []}
        onChange={(e) => onMuscleGroupChange(e.target.value)}
        size="sm"
      >
        {muscleGroups.map((mg) => (
          <SelectItem key={mg.id!} textValue={translateMuscleGroup(mg.name)}>
            <span className="text-xs">
                {translateMuscleGroup(mg.name)}
            </span>
          </SelectItem>
        ))}
      </Select>

      <Select
        label="Упражнение"
        placeholder="Выберите упражнение"
        className="w-full"
        selectedKeys={selectedExercise ? [selectedExercise] : []}
        onChange={(e) => onExerciseChange(e.target.value)}
        isDisabled={!selectedMuscleGroup}
        size="sm"
      >
        {exercises.map((ex) => (
          <SelectItem key={ex.id!} textValue={ex.name}>
            <span className="text-xs">
                {ex.name}
            </span>
          </SelectItem>
        ))}
      </Select>
    </div>
  );
};
