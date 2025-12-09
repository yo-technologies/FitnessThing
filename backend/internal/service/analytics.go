package service

import (
	"context"
	"sort"
	"time"

	"fitness-trainer/internal/domain"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetAnalytics(ctx context.Context, userID domain.ID, from, to time.Time, muscleGroupIDs []domain.ID, exerciseIDs []domain.ID, splitByExercise bool) ([]domain.AnalyticsSeries, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetAnalytics")
	defer span.Finish()

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

	calcOneRM := func(weight float64, reps int) float64 {
		if weight <= 0 || reps <= 0 {
			return 0
		}

		if reps > 12 {
			reps = 12
		}

		return weight * 36.0 / float64(37-reps)
	}

	max1RMPerExerciseDay := make(map[exerciseDayKey]float64)
	exerciseNames := make(map[string]string)
	exerciseMuscleGroups := make(map[string]string) // ExerciseID -> MuscleGroupID
	muscleGroupNames := make(map[string]string)

	// Pick bucket resolution based on requested range
	bucketDays := func() int {
		days := int(to.Sub(from).Hours()/24) + 1
		switch {
		case days <= 45:
			return 1 // daily
		case days <= 200:
			return 7 // weekly
		default:
			return 30 // monthly-ish
		}
	}()

	bucketDate := func(t time.Time) time.Time {
		if bucketDays == 1 {
			return normalizeDate(t)
		}

		t = normalizeDate(t)
		switch bucketDays {
		case 7:
			offset := (int(t.Weekday()) + 6) % 7 // Monday-start
			return t.AddDate(0, 0, -offset)
		default:
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}
	}

	for _, row := range rawData {
		oneRM := calcOneRM(row.Weight, row.Reps)
		if oneRM == 0 {
			continue
		}
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
	seriesNames := make(map[string]string)                // ID -> Name

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
		// For each day and muscle group, average the Max1RMs of exercises to avoid spikes from exercise count
		type muscleGroupDayKey struct {
			Date          time.Time
			MuscleGroupID string
		}

		type agg struct {
			sum   float64
			count int
		}

		muscleGroupAgg := make(map[muscleGroupDayKey]agg)

		for key, val := range max1RMPerExerciseDay {
			mgID := exerciseMuscleGroups[key.ExerciseID]
			mgKey := muscleGroupDayKey{Date: key.Date, MuscleGroupID: mgID}
			current := muscleGroupAgg[mgKey]
			current.sum += val
			current.count++
			muscleGroupAgg[mgKey] = current
		}

		for key, val := range muscleGroupAgg {
			seriesID := key.MuscleGroupID
			seriesName := muscleGroupNames[seriesID]
			seriesMap[seriesID] = append(seriesMap[seriesID], domain.AnalyticsPoint{
				Date:  key.Date,
				Value: val.sum / float64(val.count),
			})
			seriesNames[seriesID] = seriesName
		}
	}

	// Convert map to slice
	var result []domain.AnalyticsSeries
	const (
		smoothingWindow = 7
		emaAlpha        = 0.3
	)

	for id, points := range seriesMap {
		// bucket by chosen resolution
		bucketed := make(map[time.Time]struct {
			sum   float64
			count int
		})

		for _, p := range points {
			b := bucketDate(p.Date)
			cur := bucketed[b]
			cur.sum += p.Value
			cur.count++
			bucketed[b] = cur
		}

		points = points[:0]
		for b, agg := range bucketed {
			points = append(points, domain.AnalyticsPoint{
				Date:  b,
				Value: agg.sum / float64(agg.count),
			})
		}
		// Sort points by date
		sort.Slice(points, func(i, j int) bool {
			return points[i].Date.Before(points[j].Date)
		})

		if len(points) >= 2 {
			smoothed := make([]domain.AnalyticsPoint, len(points))
			for i, point := range points {
				start := i - (smoothingWindow - 1)
				if start < 0 {
					start = 0
				}

				end := i + 1
				var sum float64
				for j := start; j < end; j++ {
					sum += points[j].Value
				}

				windowSize := float64(end - start)
				smoothed[i] = domain.AnalyticsPoint{
					Date:  point.Date,
					Value: sum / windowSize,
				}
			}

			// Apply trailing EMA on top of the moving average to further damp spikes
			var ema float64
			for i, p := range smoothed {
				if i == 0 {
					ema = p.Value
				} else {
					ema = emaAlpha*p.Value + (1-emaAlpha)*ema
				}
				smoothed[i].Value = ema
			}

			points = smoothed
		}

		// For aggregated view (not split by exercise), show relative progress per series
		if !splitByExercise {
			var base float64
			for _, p := range points {
				if p.Value > 0 {
					base = p.Value
					break
				}
			}

			if base == 0 && len(points) > 0 {
				base = points[0].Value
			}

			if base != 0 {
				for i := range points {
					points[i].Value = points[i].Value / base * 100.0
				}
			}
		}

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
