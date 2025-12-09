"use client";

import React, { useState, useEffect } from "react";
import { AnalyticsHeader } from "@/components/Analytics/AnalyticsHeader";
import { AnalyticsFooter } from "@/components/Analytics/AnalyticsFooter";
import { AnalyticsChart } from "@/components/Analytics/AnalyticsChart";
import { authApi } from "@/api/api";
import { WorkoutAnalyticsSeries, WorkoutMuscleGroup, WorkoutExercise } from "@/api/api.generated";
import { subMonths, startOfDay, endOfDay } from "date-fns";
import { PageHeader } from "@/components/page-header";
import { Card } from "@heroui/react";

export default function AnalyticsPage() {
  // Data State
  const [muscleGroups, setMuscleGroups] = useState<WorkoutMuscleGroup[]>([]);
  const [exercises, setExercises] = useState<WorkoutExercise[]>([]);
  const [series, setSeries] = useState<WorkoutAnalyticsSeries[]>([]);
  const [loading, setLoading] = useState(false);

  // Filter State
  const [selectedMuscleGroup, setSelectedMuscleGroup] = useState<string>("");
  const [selectedExercise, setSelectedExercise] = useState<string>("");
  const [dateRangePreset, setDateRangePreset] = useState<number | null>(1);
  
  // Visibility State
  const [hiddenSeries, setHiddenSeries] = useState<Set<string>>(new Set());

  // Fetch initial data (Muscle Groups, Exercises)
  // Fetch muscle groups on mount
  useEffect(() => {
    const fetchMuscleGroups = async () => {
      try {
        const mgRes = await authApi.v1.exerciseServiceGetMuscleGroups();
        if (mgRes.data.muscleGroups) {
          setMuscleGroups(mgRes.data.muscleGroups);
        }
      } catch (e) {
        console.error("Failed to fetch muscle groups", e);
      }
    };
    fetchMuscleGroups();
  }, []);

  // Fetch exercises when muscle group changes
  useEffect(() => {
    if (!selectedMuscleGroup) {
      setExercises([]);
      setSelectedExercise("");
      return;
    }

    const fetchExercises = async () => {
      try {
        const exRes = await authApi.v1.exerciseServiceGetExercises(
          {
            muscleGroupIds: [selectedMuscleGroup],
          },
          {
            paramsSerializer: {
              indexes: null,
            },
          },
        );
        if (exRes.data.exercises) {
          setExercises(exRes.data.exercises);
        }
      } catch (e) {
        console.error("Failed to fetch exercises", e);
        setExercises([]);
      }
    };
    fetchExercises();
  }, [selectedMuscleGroup]);

  // Fetch Analytics Data
  useEffect(() => {
    const fetchAnalytics = async () => {
      setLoading(true);
      try {
        const to = new Date();
        const from = subMonths(to, dateRangePreset || 1);

        // Logic: 
        // If specific exercise is selected: splitByExercise = true (one line for the exercise)
        // If muscle group is selected but NO exercise: splitByExercise = true (multiple lines, one per exercise in the group)
        // If nothing selected: splitByExercise = false (aggregate by muscle groups)
        
        const splitByExercise = !!selectedMuscleGroup || !!selectedExercise;

        const response = await authApi.v1.workoutServiceGetAnalytics({
          from: startOfDay(from).toISOString(),
          to: endOfDay(to).toISOString(),
          muscleGroup: selectedMuscleGroup || undefined,
          exerciseId: selectedExercise || undefined,
          splitByExercise: splitByExercise,
        });
        
        if (response.data.series) {
          setSeries(response.data.series);
          // Reset hidden series when data changes
          setHiddenSeries(new Set());
        } else {
          setSeries([]);
        }
      } catch (error) {
        console.error("Failed to fetch analytics:", error);
        setSeries([]);
      } finally {
        setLoading(false);
      }
    };

    fetchAnalytics();
  }, [selectedMuscleGroup, selectedExercise, dateRangePreset]);

  const handleMuscleGroupChange = (id: string) => {
    setSelectedMuscleGroup(id);
    setSelectedExercise(""); // Reset exercise
  };

  const handleExerciseChange = (id: string) => {
    setSelectedExercise(id);
  };

  const handlePresetChange = (months: number) => {
    setDateRangePreset(months);
  };

  const handleToggleSeries = (name: string) => {
    const newHidden = new Set(hiddenSeries);
    if (newHidden.has(name)) {
      newHidden.delete(name);
    } else {
      newHidden.add(name);
    }
    setHiddenSeries(newHidden);
  };

  // Filter series passed to chart
  const visibleSeries = series.filter(s => !hiddenSeries.has(s.name!));

  return (
    <div className="py-4 flex flex-col h-full">
      <PageHeader enableBackButton title="Статистика" />
      <div className="grid grid-cols-1 gap-4 p-4">
        <Card className="bg-content1 rounded-lg shadow-sm p-3">
          <AnalyticsHeader 
            muscleGroups={muscleGroups}
            exercises={exercises}
            selectedMuscleGroup={selectedMuscleGroup}
            selectedExercise={selectedExercise}
            onMuscleGroupChange={handleMuscleGroupChange}
            onExerciseChange={handleExerciseChange}
          />

          {loading ? (
            <div className="flex h-[50vh] items-center justify-center">
              Загрузка...
            </div>
          ) : (
            <AnalyticsChart series={visibleSeries} />
          )}

          <AnalyticsFooter 
            dateRangePreset={dateRangePreset}
            onPresetChange={handlePresetChange}
            series={series}
            hiddenSeries={hiddenSeries}
            onToggleSeries={handleToggleSeries}
          />
        </Card>
      </div>
    </div>
  );
}
