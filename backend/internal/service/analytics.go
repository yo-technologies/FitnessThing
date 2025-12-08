package service

import (
	"context"
	"sort"
	"time"

	"fitness-trainer/internal/domain"
)

func (s *Service) GetAnalytics(ctx context.Context, userID domain.ID, from, to time.Time, muscleGroupIDs []domain.ID, exerciseIDs []domain.ID, splitByExercise bool) ([]domain.AnalyticsSeries, error) {
	rawData, err := s.repository.GetAnalyticsRawData(ctx, userID, from, to, muscleGroupIDs, exerciseIDs)
	if err != nil {
		return nil, err
	}

	// Map: Date -> Key -> Max1RM
	// Key depends on splitByExercise.
	// If splitByExercise: Key = ExerciseID
	// Else: Key = MuscleGroupID

	// First, calculate 1RM for each raw entry and find max per exercise per day.
	type exerciseDayKey struct {
		Date       time.Time
		ExerciseID string
	}
	
	// Helper to normalize date to day (strip time)
	normalizeDate := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}

	max1RMPerExerciseDay := make(map[exerciseDayKey]float64)
	exerciseNames := make(map[string]string)
	exerciseMuscleGroups := make(map[string]string) // ExerciseID -> MuscleGroupID
	muscleGroupNames := make(map[string]string)

	for _, row := range rawData {
		oneRM := row.Weight * (1 + float64(row.Reps)/30.0)
		date := normalizeDate(row.Date)
		key := exerciseDayKey{Date: date, ExerciseID: row.ExerciseID}

		if val, ok := max1RMPerExerciseDay[key]; !ok || oneRM > val {
			max1RMPerExerciseDay[key] = oneRM
		}

		exerciseNames[row.ExerciseID] = row.ExerciseName
		exerciseMuscleGroups[row.ExerciseID] = row.MuscleGroupID
		muscleGroupNames[row.MuscleGroupID] = row.MuscleGroupName
	}

	// Now aggregate based on splitByExercise
	seriesMap := make(map[string][]domain.AnalyticsPoint) // SeriesName -> Points
	seriesNames := make(map[string]string) // ID -> Name

	if splitByExercise {
		for key, val := range max1RMPerExerciseDay {
			seriesID := key.ExerciseID
			seriesName := exerciseNames[seriesID]
			seriesMap[seriesID] = append(seriesMap[seriesID], domain.AnalyticsPoint{
				Date:  key.Date,
				Value: val,
			})
			seriesNames[seriesID] = seriesName
		}
	} else {
		// Aggregate by Muscle Group
		// For each day and muscle group, sum the Max1RMs of exercises
		type muscleGroupDayKey struct {
			Date          time.Time
			MuscleGroupID string
		}
		
		muscleGroupSums := make(map[muscleGroupDayKey]float64)

		for key, val := range max1RMPerExerciseDay {
			mgID := exerciseMuscleGroups[key.ExerciseID]
			mgKey := muscleGroupDayKey{Date: key.Date, MuscleGroupID: mgID}
			muscleGroupSums[mgKey] += val
		}

		for key, val := range muscleGroupSums {
			seriesID := key.MuscleGroupID
			seriesName := muscleGroupNames[seriesID]
			seriesMap[seriesID] = append(seriesMap[seriesID], domain.AnalyticsPoint{
				Date:  key.Date,
				Value: val,
			})
			seriesNames[seriesID] = seriesName
		}
	}

	// Convert map to slice
	var result []domain.AnalyticsSeries
	for id, points := range seriesMap {
		// Sort points by date
		sort.Slice(points, func(i, j int) bool {
			return points[i].Date.Before(points[j].Date)
		})
		
		result = append(result, domain.AnalyticsSeries{
			Name:   seriesNames[id],
			Points: points,
		})
	}
	
	// Sort series by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}
