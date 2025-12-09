package domain

import "time"

type AnalyticsRawData struct {
	Date            time.Time `db:"finished_at"`
	ExerciseID      string    `db:"exercise_id"`
	ExerciseName    string    `db:"exercise_name"`
	MuscleGroupID   string    `db:"muscle_group_id"`
	MuscleGroupName string    `db:"muscle_group_name"`
	Weight          float64   `db:"weight"`
	Reps            int       `db:"reps"`
}

type AnalyticsPoint struct {
	Date  time.Time
	Value float64
}

type AnalyticsSeries struct {
	Name   string
	Points []AnalyticsPoint
}
