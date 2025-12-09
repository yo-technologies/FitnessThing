"use client";

import React, { useMemo } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { WorkoutAnalyticsSeries } from "@/api/api.generated";
import { format, parseISO } from "date-fns";
import { useTheme } from "next-themes";
import { translateMuscleGroup } from "@/config/muscle-groups";

interface AnalyticsChartProps {
  series: WorkoutAnalyticsSeries[];
}

export const AnalyticsChart: React.FC<AnalyticsChartProps> = ({ series }) => {
  const { theme } = useTheme();
  const isDark = theme === "dark";

  const data = useMemo(() => {
    const dateMap = new Map<string, any>();

    series.forEach((s) => {
      if (!s.points || !s.name) return;
      s.points.forEach((p) => {
        if (!p.date || p.value === undefined) return;
        const dateStr = p.date; 
        if (!dateMap.has(dateStr)) {
          dateMap.set(dateStr, { date: dateStr });
        }
        const entry = dateMap.get(dateStr);
        entry[s.name!] = Math.round(p.value * 100) / 100;
      });
    });

    return Array.from(dateMap.values()).sort((a, b) =>
      a.date.localeCompare(b.date)
    );
  }, [series]);

  const colors = [
    "#8884d8",
    "#82ca9d",
    "#ffc658",
    "#ff7300",
    "#0088fe",
    "#00c49f",
    "#ffbb28",
    "#ff8042",
    "#a83279",
    "#32a852",
    "#3252a8",
    "#a87e32",
  ];

  if (series.length === 0) {
    return (
      <div className="h-[50vh] flex items-center justify-center text-gray-500">
        Нет данных для отображения
      </div>
    );
  }

  return (
    <div className="h-[50vh] w-full">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart
          data={data}
        >
          <CartesianGrid strokeDasharray="3 3" stroke={isDark ? "#374151" : "#e5e7eb"} />
          <XAxis
            dataKey="date"
            tickFormatter={(str) => format(parseISO(str), "dd MMM")}
            stroke={isDark ? "#9ca3af" : "#4b5563"}
            tick={{ fontSize: 12 }}
          />
          <YAxis
            stroke={isDark ? "#9ca3af" : "#4b5563"}
            tick={{ fontSize: 12 }}
            width={20}
            padding={{bottom: 10}}
          />
          <Tooltip
            labelFormatter={(str) => format(parseISO(str), "dd MMMM yyyy")}
            contentStyle={{
              backgroundColor: isDark ? "#1f2937" : "#ffffff",
              borderColor: isDark ? "#374151" : "#e5e7eb",
              color: isDark ? "#f3f4f6" : "#111827",
              fontSize: "12px",
            }}
          />
          {/* <Legend wrapperStyle={{ fontSize: "12px", alignItems: "left" }} /> */}
          {series.map((s, index) => (
            <Line
            key={s.name}
            name={translateMuscleGroup(s.name)}
            type="monotone"
            dataKey={s.name!}
            stroke={colors[index % colors.length]}
            activeDot={{ r: 8 }}
            connectNulls
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
};
