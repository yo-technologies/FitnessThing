package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"

	openai_client "fitness-trainer/internal/clients/openai"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
	"github.com/opentracing/opentracing-go"
)

type agentToolContext struct {
	userID    domain.ID
	workoutID utils.Nullable[domain.ID]
}

type agentToolHandler func(ctx context.Context, toolCtx agentToolContext, raw json.RawMessage) (string, error)

type agentTool struct {
	name       string
	definition openai.ChatCompletionToolParam
	handler    agentToolHandler
}

func annotateAgentToolSpan(span opentracing.Span, toolCtx agentToolContext, toolName string) {
	span.SetTag("tool", toolName)
	if toolCtx.userID != (domain.ID{}) {
		span.SetTag("user_id", toolCtx.userID.String())
	}
	if toolCtx.workoutID.IsValid {
		span.SetTag("workout_id", toolCtx.workoutID.V.String())
	}
}

func (s *Service) ensureChatTools() {
	s.chatToolsOnce.Do(func() {
		s.chatTools = map[string]agentTool{}

		for _, tool := range []agentTool{
			s.newListMuscleGroupsTool(),
			s.newListExercisesTool(),
			s.newGetWorkoutHistoryTool(),
			s.newGetExerciseHistoryTool(),
			s.newGetWorkoutPlanTool(),
			s.newAddExercisesToWorkoutTool(),
			s.newRemoveExerciseLogsTool(),
			s.newReplaceExpectedSetsTool(),
			s.newUpdateSetLogTool(),
		} {
			if tool.name == "" {
				panic("agent tool definition missing name")
			}
			s.chatTools[tool.name] = tool
		}
	})
}

func (s *Service) chatToolDefinitions() []openai.ChatCompletionToolParam {
	s.ensureChatTools()

	defs := make([]openai.ChatCompletionToolParam, 0, len(s.chatTools))
	for _, tool := range s.chatTools {
		defs = append(defs, tool.definition)
	}
	return defs
}

func (s *Service) newListMuscleGroupsTool() agentTool {
	return agentTool{
		name: "list_muscle_groups",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("list_muscle_groups"),
				Description: openai.String("Return the available muscle groups with their identifiers."),
				Parameters: openai.F(shared.FunctionParameters{
					"type": "object",
				}),
				Strict: openai.Bool(true),
			}),
		},
		handler: func(ctx context.Context, toolCtx agentToolContext, _ json.RawMessage) (string, error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.list_muscle_groups")
			defer span.Finish()
			annotateAgentToolSpan(span, toolCtx, "list_muscle_groups")

			groups, err := s.repository.GetMuscleGroups(ctx)
			if err != nil {
				return "", fmt.Errorf("failed to load muscle groups: %w", err)
			}

			payload := struct {
				MuscleGroups []dto.MuscleGroupDTO `json:"muscle_groups"`
			}{MuscleGroups: groups}

			raw, err := json.Marshal(payload)
			if err != nil {
				return "", fmt.Errorf("failed to marshal muscle groups: %w", err)
			}

			return string(raw), nil
		},
	}
}

type listExercisesArgs struct {
	MuscleGroupIDs     []string `json:"muscle_group_ids"`
	ExcludeExerciseIDs []string `json:"exclude_exercise_ids"`
	Limit              *int     `json:"limit"`
}

func (s *Service) newListExercisesTool() agentTool {
	schema := shared.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"muscle_group_ids": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string", "description": "UUID of the muscle group"},
			},
			"exclude_exercise_ids": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string", "description": "UUID of exercises to exclude"},
			},
			"limit": map[string]any{
				"type":        "integer",
				"minimum":     1,
				"maximum":     50,
				"description": "Maximum number of exercises to return. Defaults to 10.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "list_exercises",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("list_exercises"),
				Description: openai.String("Fetch exercises optionally filtered by muscle groups or search query."),
				Parameters:  openai.F(schema),
				Strict:      openai.Bool(true),
			}),
		},
		handler: func(ctx context.Context, toolCtx agentToolContext, raw json.RawMessage) (string, error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.list_exercises")
			defer span.Finish()
			annotateAgentToolSpan(span, toolCtx, "list_exercises")
			if len(raw) > 0 {
				span.SetTag("arguments.present", true)
			}

			var args listExercisesArgs
			if len(raw) > 0 {
				if err := json.Unmarshal(raw, &args); err != nil {
					return "", fmt.Errorf("invalid arguments for list_exercises: %w", err)
				}
			}

			limit := 10
			if args.Limit != nil {
				limit = *args.Limit
			}

			toIDs := func(values []string) ([]domain.ID, error) {
				ids := make([]domain.ID, 0, len(values))
				for _, v := range values {
					if v == "" {
						continue
					}
					id, err := domain.ParseID(v)
					if err != nil {
						logger.Errorf("failed to parse id %s: %w", v, err)
						continue
					}
					ids = append(ids, id)
				}
				return ids, nil
			}

			muscleGroupIDs, err := toIDs(args.MuscleGroupIDs)
			if err != nil {
				return "", err
			}

			excludedExerciseIDs, err := toIDs(args.ExcludeExerciseIDs)
			if err != nil {
				return "", err
			}

			exercises, err := s.repository.GetExercises(ctx, muscleGroupIDs, excludedExerciseIDs)
			if err != nil {
				return "", fmt.Errorf("failed to get exercises: %w", err)
			}

			if len(exercises) > limit {
				exercises = exercises[:limit]
			}

			type exercisePayload struct {
				ID                 string               `json:"id"`
				Name               string               `json:"name"`
				Description        string               `json:"description"`
				TargetMuscleGroups []domain.MuscleGroup `json:"target_muscle_groups"`
			}

			payload := struct {
				Exercises []exercisePayload `json:"exercises"`
			}{Exercises: make([]exercisePayload, 0, len(exercises))}

			for _, exercise := range exercises {
				payload.Exercises = append(payload.Exercises, exercisePayload{
					ID:                 exercise.ID.String(),
					Name:               exercise.Name,
					Description:        exercise.Description,
					TargetMuscleGroups: exercise.TargetMuscleGroups,
				})
			}

			rawResp, err := json.Marshal(payload)
			if err != nil {
				return "", fmt.Errorf("failed to marshal exercises: %w", err)
			}

			return string(rawResp), nil
		},
	}
}

type getWorkoutHistoryArgs struct {
	Limit int `json:"limit"`
}

type getExerciseHistoryArgs struct {
	MuscleGroupIDs     []string `json:"muscle_group_ids"`
	ExerciseLimit      *int     `json:"exercise_limit"`
	LogsPerExercise    *int     `json:"logs_per_exercise"`
	ExcludeExerciseIDs []string `json:"exclude_exercise_ids"`
}

func (s *Service) newGetWorkoutHistoryTool() agentTool {
	schema := shared.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"limit": map[string]any{
				"type":        "integer",
				"minimum":     1,
				"maximum":     20,
				"description": "Maximum number of workouts to return. Defaults to 5.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "get_workout_history",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("get_workout_history"),
				Description: openai.String("Return recent workouts for the current user including exercises."),
				Parameters:  openai.F(schema),
				Strict:      openai.Bool(true),
			}),
		},
		handler: func(ctx context.Context, toolCtx agentToolContext, raw json.RawMessage) (string, error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.get_workout_history")
			defer span.Finish()
			annotateAgentToolSpan(span, toolCtx, "get_workout_history")
			if len(raw) > 0 {
				span.SetTag("arguments.present", true)
			}

			if toolCtx.userID == (domain.ID{}) {
				return "", fmt.Errorf("user context is required for get_workout_history")
			}

			args := getWorkoutHistoryArgs{Limit: 5}
			if len(raw) > 0 {
				if err := json.Unmarshal(raw, &args); err != nil {
					return "", fmt.Errorf("invalid arguments for get_workout_history: %w", err)
				}
				if args.Limit <= 0 {
					args.Limit = 5
				}
			}

			workouts, err := s.repository.GetWorkouts(ctx, toolCtx.userID, args.Limit, 0)
			if err != nil {
				return "", fmt.Errorf("failed to load workouts: %w", err)
			}

			type workoutPayload struct {
				ID            string    `json:"id"`
				CreatedAt     time.Time `json:"created_at"`
				FinishedAt    time.Time `json:"finished_at"`
				ExerciseNames []string  `json:"exercise_names"`
				Reasoning     string    `json:"reasoning"`
			}

			payload := struct {
				Workouts []workoutPayload `json:"workouts"`
			}{Workouts: make([]workoutPayload, 0, len(workouts))}

			for _, workout := range workouts {
				exerciseLogs, err := s.repository.GetExerciseLogsByWorkoutID(ctx, workout.ID)
				if err != nil {
					return "", fmt.Errorf("failed to load exercise logs for workout %s: %w", workout.ID, err)
				}

				names := make([]string, 0, len(exerciseLogs))
				for _, log := range exerciseLogs {
					exercise, err := s.repository.GetExerciseByID(ctx, log.ExerciseID)
					if err != nil {
						return "", fmt.Errorf("failed to load exercise %s: %w", log.ExerciseID, err)
					}
					names = append(names, exercise.Name)
				}

				payload.Workouts = append(payload.Workouts, workoutPayload{
					ID:            workout.ID.String(),
					CreatedAt:     workout.CreatedAt,
					FinishedAt:    workout.FinishedAt,
					ExerciseNames: names,
					Reasoning:     workout.Reasoning,
				})
			}

			rawResp, err := json.Marshal(payload)
			if err != nil {
				return "", fmt.Errorf("failed to marshal workout history: %w", err)
			}

			return string(rawResp), nil
		},
	}
}

func (s *Service) newGetExerciseHistoryTool() agentTool {
	schema := shared.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"muscle_group_ids": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string", "description": "UUID of the muscle group"},
			},
			"exercise_limit": map[string]any{
				"type":        "integer",
				"minimum":     1,
				"maximum":     20,
				"description": "Maximum number of exercises to include. Defaults to 5.",
			},
			"logs_per_exercise": map[string]any{
				"type":        "integer",
				"minimum":     1,
				"maximum":     20,
				"description": "Maximum number of recent logs per exercise. Defaults to 3.",
			},
			"exclude_exercise_ids": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string", "description": "UUID of exercises to exclude"},
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "get_exercise_history",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("get_exercise_history"),
				Description: openai.String("Return recent performance history for exercises filtered by muscle groups."),
				Parameters:  openai.F(schema),
				Strict:      openai.Bool(true),
			}),
		},
		handler: func(ctx context.Context, toolCtx agentToolContext, raw json.RawMessage) (string, error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.get_exercise_history")
			defer span.Finish()
			annotateAgentToolSpan(span, toolCtx, "get_exercise_history")
			if len(raw) > 0 {
				span.SetTag("arguments.present", true)
			}

			if toolCtx.userID == (domain.ID{}) {
				return "", fmt.Errorf("user context is required for get_exercise_history")
			}

			args := getExerciseHistoryArgs{}
			if len(raw) > 0 {
				if err := json.Unmarshal(raw, &args); err != nil {
					return "", fmt.Errorf("invalid arguments for get_exercise_history: %w", err)
				}
			}

			exerciseLimit := 5
			if args.ExerciseLimit != nil && *args.ExerciseLimit > 0 {
				exerciseLimit = *args.ExerciseLimit
			}

			logsLimit := 3
			if args.LogsPerExercise != nil && *args.LogsPerExercise > 0 {
				logsLimit = *args.LogsPerExercise
			}

			toIDs := func(values []string) ([]domain.ID, error) {
				ids := make([]domain.ID, 0, len(values))
				for _, v := range values {
					if v == "" {
						continue
					}
					id, err := domain.ParseID(v)
					if err != nil {
						return nil, fmt.Errorf("failed to parse id %s: %w", v, err)
					}
					ids = append(ids, id)
				}
				return ids, nil
			}

			muscleGroupIDs, err := toIDs(args.MuscleGroupIDs)
			if err != nil {
				return "", err
			}

			excludedExerciseIDs, err := toIDs(args.ExcludeExerciseIDs)
			if err != nil {
				return "", err
			}

			exercises, err := s.repository.GetExercises(ctx, muscleGroupIDs, excludedExerciseIDs)
			if err != nil {
				return "", fmt.Errorf("failed to get exercises: %w", err)
			}

			if len(exercises) > exerciseLimit {
				exercises = exercises[:exerciseLimit]
			}

			type setLogPayload struct {
				ID      string  `json:"id"`
				Reps    int     `json:"reps"`
				Weight  float32 `json:"weight"`
				TimeSec int     `json:"time_seconds"`
			}

			type logPayload struct {
				ID        string          `json:"id"`
				WorkoutID string          `json:"workout_id"`
				CreatedAt time.Time       `json:"created_at"`
				Notes     string          `json:"notes"`
				SetLogs   []setLogPayload `json:"set_logs"`
			}

			type exercisePayload struct {
				ID                 string               `json:"id"`
				Name               string               `json:"name"`
				Description        string               `json:"description"`
				TargetMuscleGroups []domain.MuscleGroup `json:"target_muscle_groups"`
				Logs               []logPayload         `json:"logs"`
			}

			response := struct {
				Exercises []exercisePayload `json:"exercises"`
			}{Exercises: make([]exercisePayload, 0, len(exercises))}

			for _, exercise := range exercises {
				logs, err := s.repository.GetExerciseLogsByExerciseIDAndUserID(ctx, exercise.ID, toolCtx.userID, 0, logsLimit)
				if err != nil {
					return "", fmt.Errorf("failed to load exercise logs for exercise %s: %w", exercise.ID, err)
				}

				payload := exercisePayload{
					ID:                 exercise.ID.String(),
					Name:               exercise.Name,
					Description:        exercise.Description,
					TargetMuscleGroups: exercise.TargetMuscleGroups,
					Logs:               make([]logPayload, 0, len(logs)),
				}

				for _, logEntry := range logs {
					setLogs, err := s.repository.GetSetLogsByExerciseLogID(ctx, logEntry.ID)
					if err != nil {
						return "", fmt.Errorf("failed to load set logs for exercise log %s: %w", logEntry.ID, err)
					}

					logPayloadItem := logPayload{
						ID:        logEntry.ID.String(),
						WorkoutID: logEntry.WorkoutID.String(),
						CreatedAt: logEntry.CreatedAt,
						Notes:     logEntry.Notes,
						SetLogs:   make([]setLogPayload, 0, len(setLogs)),
					}

					for _, setLog := range setLogs {
						logPayloadItem.SetLogs = append(logPayloadItem.SetLogs, setLogPayload{
							ID:      setLog.ID.String(),
							Reps:    setLog.Reps,
							Weight:  setLog.Weight,
							TimeSec: int(setLog.Time.Seconds()),
						})
					}

					payload.Logs = append(payload.Logs, logPayloadItem)
				}

				response.Exercises = append(response.Exercises, payload)
			}

			rawResp, err := json.Marshal(response)
			if err != nil {
				return "", fmt.Errorf("failed to marshal exercise history: %w", err)
			}

			return string(rawResp), nil
		},
	}
}

type workoutPlanExercise struct {
	ExerciseLogID string `json:"exercise_log_id"`
	Exercise      struct {
		ID          string               `json:"id"`
		Name        string               `json:"name"`
		Description string               `json:"description"`
		Targets     []domain.MuscleGroup `json:"target_muscle_groups"`
	} `json:"exercise"`
	ExpectedSets []struct {
		ID         string  `json:"id"`
		SetType    string  `json:"set_type"`
		Reps       int     `json:"reps"`
		Weight     float32 `json:"weight"`
		TimeSecond int     `json:"time_seconds"`
	} `json:"expected_sets"`
	SetLogs []struct {
		ID         string  `json:"id"`
		Reps       int     `json:"reps"`
		Weight     float32 `json:"weight"`
		TimeSecond int     `json:"time_seconds"`
	} `json:"set_logs"`
}

func (s *Service) newGetWorkoutPlanTool() agentTool {
	return agentTool{
		name: "get_workout_plan",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("get_workout_plan"),
				Description: openai.String("Return the current state of the workout including exercises, expected sets, and logged performance."),
				Parameters: openai.F(shared.FunctionParameters{
					"type":                 "object",
					"additionalProperties": false,
				}),
				Strict: openai.Bool(true),
			}),
		},
		handler: func(ctx context.Context, toolCtx agentToolContext, _ json.RawMessage) (string, error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.get_workout_plan")
			defer span.Finish()
			annotateAgentToolSpan(span, toolCtx, "get_workout_plan")

			if !toolCtx.workoutID.IsValid {
				return "", fmt.Errorf("workout context is required for get_workout_plan")
			}

			exerciseLogs, err := s.repository.GetExerciseLogsByWorkoutID(ctx, toolCtx.workoutID.V)
			if err != nil {
				return "", fmt.Errorf("failed to load exercise logs: %w", err)
			}

			plan := struct {
				Exercises []workoutPlanExercise `json:"exercises"`
			}{Exercises: make([]workoutPlanExercise, 0, len(exerciseLogs))}

			for _, logEntry := range exerciseLogs {
				exercise, err := s.repository.GetExerciseByID(ctx, logEntry.ExerciseID)
				if err != nil {
					return "", fmt.Errorf("failed to load exercise %s: %w", logEntry.ExerciseID, err)
				}

				expectedSets, err := s.repository.GetExpectedSetsByExerciseLogID(ctx, logEntry.ID)
				if err != nil {
					return "", fmt.Errorf("failed to load expected sets for exercise log %s: %w", logEntry.ID, err)
				}

				setLogs, err := s.repository.GetSetLogsByExerciseLogID(ctx, logEntry.ID)
				if err != nil {
					return "", fmt.Errorf("failed to load set logs for exercise log %s: %w", logEntry.ID, err)
				}

				item := workoutPlanExercise{
					ExerciseLogID: logEntry.ID.String(),
					ExpectedSets: make([]struct {
						ID         string  `json:"id"`
						SetType    string  `json:"set_type"`
						Reps       int     `json:"reps"`
						Weight     float32 `json:"weight"`
						TimeSecond int     `json:"time_seconds"`
					}, 0, len(expectedSets)),
					SetLogs: make([]struct {
						ID         string  `json:"id"`
						Reps       int     `json:"reps"`
						Weight     float32 `json:"weight"`
						TimeSecond int     `json:"time_seconds"`
					}, 0, len(setLogs)),
				}
				item.Exercise.ID = exercise.ID.String()
				item.Exercise.Name = exercise.Name
				item.Exercise.Description = exercise.Description
				item.Exercise.Targets = exercise.TargetMuscleGroups

				for _, set := range expectedSets {
					seconds := int(set.Time / time.Second)
					item.ExpectedSets = append(item.ExpectedSets, struct {
						ID         string  `json:"id"`
						SetType    string  `json:"set_type"`
						Reps       int     `json:"reps"`
						Weight     float32 `json:"weight"`
						TimeSecond int     `json:"time_seconds"`
					}{
						ID:         set.ID.String(),
						SetType:    set.SetType.String(),
						Reps:       set.Reps,
						Weight:     set.Weight,
						TimeSecond: seconds,
					})
				}

				for _, setLog := range setLogs {
					seconds := int(setLog.Time / time.Second)
					item.SetLogs = append(item.SetLogs, struct {
						ID         string  `json:"id"`
						Reps       int     `json:"reps"`
						Weight     float32 `json:"weight"`
						TimeSecond int     `json:"time_seconds"`
					}{
						ID:         setLog.ID.String(),
						Reps:       setLog.Reps,
						Weight:     setLog.Weight,
						TimeSecond: seconds,
					})
				}

				plan.Exercises = append(plan.Exercises, item)
			}

			raw, err := json.Marshal(plan)
			if err != nil {
				return "", fmt.Errorf("failed to marshal workout plan: %w", err)
			}

			return string(raw), nil
		},
	}
}

type addExercisesToWorkoutArgs struct {
	ExerciseIDs []string `json:"exercise_ids"`
}

type removeExerciseLogsArgs struct {
	ExerciseLogIDs []string `json:"exercise_log_ids"`
}

func (s *Service) newAddExercisesToWorkoutTool() agentTool {
	schema := shared.FunctionParameters{
		"type":     "object",
		"required": []string{"exercise_ids"},
		"properties": map[string]any{
			"exercise_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string", "description": "UUID of exercise to add"},
				"minItems":    1,
				"description": "List of exercise UUIDs to add to the workout.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "add_exercises_to_workout",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("add_exercises_to_workout"),
				Description: openai.String("Add one or more exercises to the current workout. Creates exercise logs for each exercise."),
				Parameters:  openai.F(schema),
				Strict:      openai.Bool(true),
			}),
		},
		handler: func(ctx context.Context, toolCtx agentToolContext, raw json.RawMessage) (string, error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.add_exercises_to_workout")
			defer span.Finish()
			annotateAgentToolSpan(span, toolCtx, "add_exercises_to_workout")
			if len(raw) > 0 {
				span.SetTag("arguments.present", true)
			}
			if !toolCtx.workoutID.IsValid {
				return "", fmt.Errorf("workout context is required for add_exercises_to_workout")
			}

			var args addExercisesToWorkoutArgs
			if err := json.Unmarshal(raw, &args); err != nil {
				return "", fmt.Errorf("invalid arguments for add_exercises_to_workout: %w", err)
			}
			if len(args.ExerciseIDs) == 0 {
				return "", fmt.Errorf("exercise_ids is required and must not be empty")
			}

			span.SetTag("exercise_count", len(args.ExerciseIDs))

			type exerciseLogPayload struct {
				ExerciseLogID string `json:"exercise_log_id"`
				Exercise      struct {
					ID          string               `json:"id"`
					Name        string               `json:"name"`
					Description string               `json:"description"`
					Targets     []domain.MuscleGroup `json:"target_muscle_groups"`
				} `json:"exercise"`
			}

			payload := struct {
				AddedExercises []exerciseLogPayload `json:"added_exercises"`
			}{
				AddedExercises: make([]exerciseLogPayload, 0, len(args.ExerciseIDs)),
			}

			for _, exerciseIDStr := range args.ExerciseIDs {
				exerciseID, err := domain.ParseID(exerciseIDStr)
				if err != nil {
					return "", fmt.Errorf("invalid exercise_id %q: %w", exerciseIDStr, err)
				}

				exercise, err := s.repository.GetExerciseByID(ctx, exerciseID)
				if err != nil {
					return "", fmt.Errorf("failed to load exercise %s: %w", exerciseID, err)
				}

				logEntry, err := s.LogExercise(ctx, toolCtx.userID, toolCtx.workoutID.V, exerciseID)
				if err != nil {
					return "", fmt.Errorf("failed to log exercise %s: %w", exercise.Name, err)
				}

				item := exerciseLogPayload{
					ExerciseLogID: logEntry.ID.String(),
				}
				item.Exercise.ID = exercise.ID.String()
				item.Exercise.Name = exercise.Name
				item.Exercise.Description = exercise.Description
				item.Exercise.Targets = exercise.TargetMuscleGroups

				payload.AddedExercises = append(payload.AddedExercises, item)
			}

			rawResp, err := json.Marshal(payload)
			if err != nil {
				return "", fmt.Errorf("failed to marshal add_exercises_to_workout result: %w", err)
			}

			return string(rawResp), nil
		},
	}
}

func (s *Service) newRemoveExerciseLogsTool() agentTool {
	schema := shared.FunctionParameters{
		"type":     "object",
		"required": []string{"exercise_log_ids"},
		"properties": map[string]any{
			"exercise_log_ids": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string", "description": "UUID of exercise log to remove"},
				"minItems":    1,
				"description": "List of exercise log UUIDs to remove from the workout.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "remove_exercise_logs_from_workout",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("remove_exercise_logs_from_workout"),
				Description: openai.String("Remove one or more exercise logs from the current workout."),
				Parameters:  openai.F(schema),
				Strict:      openai.Bool(true),
			}),
		},
		handler: func(ctx context.Context, toolCtx agentToolContext, raw json.RawMessage) (string, error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.remove_exercise_logs_from_workout")
			defer span.Finish()
			annotateAgentToolSpan(span, toolCtx, "remove_exercise_logs_from_workout")
			if len(raw) > 0 {
				span.SetTag("arguments.present", true)
			}

			if toolCtx.userID == (domain.ID{}) {
				return "", fmt.Errorf("user context is required for remove_exercise_logs_from_workout")
			}

			if !toolCtx.workoutID.IsValid {
				return "", fmt.Errorf("workout context is required for remove_exercise_logs_from_workout")
			}

			var args removeExerciseLogsArgs
			if err := json.Unmarshal(raw, &args); err != nil {
				return "", fmt.Errorf("invalid arguments for remove_exercise_logs_from_workout: %w", err)
			}
			if len(args.ExerciseLogIDs) == 0 {
				return "", fmt.Errorf("exercise_log_ids must contain at least one id")
			}

			uniqueLogIDs := make([]domain.ID, 0, len(args.ExerciseLogIDs))
			seen := make(map[domain.ID]struct{}, len(args.ExerciseLogIDs))
			for idx, idStr := range args.ExerciseLogIDs {
				if idStr == "" {
					return "", fmt.Errorf("exercise_log_ids[%d] must not be empty", idx)
				}

				logID, err := domain.ParseID(idStr)
				if err != nil {
					return "", fmt.Errorf("invalid exercise_log_id at index %d: %w", idx, err)
				}

				if _, exists := seen[logID]; exists {
					continue
				}
				seen[logID] = struct{}{}
				uniqueLogIDs = append(uniqueLogIDs, logID)
			}

			if len(uniqueLogIDs) == 0 {
				return "", fmt.Errorf("no valid exercise_log_ids provided")
			}

			span.SetTag("exercise_log_count", len(uniqueLogIDs))

			type removedExerciseLogPayload struct {
				ExerciseLogID string `json:"exercise_log_id"`
				Exercise      struct {
					ID          string               `json:"id"`
					Name        string               `json:"name"`
					Description string               `json:"description"`
					Targets     []domain.MuscleGroup `json:"target_muscle_groups"`
				} `json:"exercise"`
			}

			removed := make([]removedExerciseLogPayload, 0, len(uniqueLogIDs))

			for _, logID := range uniqueLogIDs {
				exerciseLog, err := s.repository.GetExerciseLogByID(ctx, logID)
				if err != nil {
					return "", fmt.Errorf("failed to load exercise log %s: %w", logID.String(), err)
				}

				if exerciseLog.WorkoutID != toolCtx.workoutID.V {
					return "", fmt.Errorf("exercise log %s does not belong to the current workout", logID.String())
				}

				exercise, err := s.repository.GetExerciseByID(ctx, exerciseLog.ExerciseID)
				if err != nil {
					return "", fmt.Errorf("failed to load exercise %s: %w", exerciseLog.ExerciseID.String(), err)
				}

				item := removedExerciseLogPayload{
					ExerciseLogID: logID.String(),
				}
				item.Exercise.ID = exercise.ID.String()
				item.Exercise.Name = exercise.Name
				item.Exercise.Description = exercise.Description
				item.Exercise.Targets = exercise.TargetMuscleGroups

				removed = append(removed, item)
			}

			for _, logID := range uniqueLogIDs {
				if err := s.DeleteExerciseLog(ctx, toolCtx.userID, toolCtx.workoutID.V, logID); err != nil {
					return "", fmt.Errorf("failed to delete exercise log %s: %w", logID.String(), err)
				}
			}

			payload := struct {
				RemovedExerciseLogs []removedExerciseLogPayload `json:"removed_exercise_logs"`
				RemovedCount        int                         `json:"removed_count"`
			}{
				RemovedExerciseLogs: removed,
				RemovedCount:        len(removed),
			}

			rawResp, err := json.Marshal(payload)
			if err != nil {
				return "", fmt.Errorf("failed to marshal remove_exercise_logs_from_workout result: %w", err)
			}

			return string(rawResp), nil
		},
	}
}

type replaceExpectedSetsArgs struct {
	ExerciseLogID string `json:"exercise_log_id"`
	Sets          []struct {
		SetType     string   `json:"set_type"`
		Reps        *int     `json:"reps"`
		Weight      *float64 `json:"weight"`
		TimeSeconds *int     `json:"time_seconds"`
	} `json:"sets"`
}

func (s *Service) newReplaceExpectedSetsTool() agentTool {
	schema := shared.FunctionParameters{
		"type":     "object",
		"required": []string{"exercise_log_id", "sets"},
		"properties": map[string]any{
			"exercise_log_id": map[string]any{
				"type":        "string",
				"description": "UUID of the exercise log to update.",
			},
			"sets": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":     "object",
					"required": []string{"set_type"},
					"properties": map[string]any{
						"set_type": map[string]any{
							"type": "string",
							"enum": []string{"reps", "weight", "time"},
						},
						"reps": map[string]any{
							"type":        "integer",
							"description": "Target repetitions for the set.",
						},
						"weight": map[string]any{
							"type":        "number",
							"description": "Target weight in the workout's unit.",
						},
						"time_seconds": map[string]any{
							"type":        "integer",
							"description": "Target duration in seconds for time-based sets.",
						},
					},
					"additionalProperties": false,
				},
				"minItems": 1,
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "replace_expected_sets",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("replace_expected_sets"),
				Description: openai.String("Replace the planned sets for an exercise log in the current workout."),
				Parameters:  openai.F(schema),
				Strict:      openai.Bool(true),
			}),
		},
		handler: func(ctx context.Context, toolCtx agentToolContext, raw json.RawMessage) (string, error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.replace_expected_sets")
			defer span.Finish()
			annotateAgentToolSpan(span, toolCtx, "replace_expected_sets")
			if len(raw) > 0 {
				span.SetTag("arguments.present", true)
			}

			if !toolCtx.workoutID.IsValid {
				return "", fmt.Errorf("workout context is required for replace_expected_sets")
			}

			var args replaceExpectedSetsArgs
			if err := json.Unmarshal(raw, &args); err != nil {
				return "", fmt.Errorf("invalid arguments for replace_expected_sets: %w", err)
			}

			if args.ExerciseLogID == "" {
				return "", fmt.Errorf("exercise_log_id is required")
			}

			exerciseLogID, err := domain.ParseID(args.ExerciseLogID)
			if err != nil {
				return "", fmt.Errorf("invalid exercise_log_id: %w", err)
			}

			inputs := make([]dto.ExpectedSetInput, 0, len(args.Sets))
			for idx, set := range args.Sets {
				setType, err := domain.NewSetType(set.SetType)
				if err != nil {
					return "", fmt.Errorf("invalid set_type at index %d: %w", idx, err)
				}

				reps := 0
				if set.Reps != nil {
					reps = *set.Reps
				}

				weight := float32(0)
				if set.Weight != nil {
					weight = float32(*set.Weight)
				}

				var duration time.Duration
				if set.TimeSeconds != nil {
					duration = time.Duration(*set.TimeSeconds) * time.Second
				}

				inputs = append(inputs, dto.ExpectedSetInput{
					SetType: setType,
					Reps:    reps,
					Weight:  weight,
					Time:    duration,
				})
			}

			if err := s.ReplaceExpectedSets(ctx, toolCtx.userID, toolCtx.workoutID.V, exerciseLogID, inputs); err != nil {
				return "", fmt.Errorf("failed to replace expected sets: %w", err)
			}

			return `{"status":"success"}`, nil
		},
	}
}

type updateSetLogArgs struct {
	ExerciseLogID string   `json:"exercise_log_id"`
	SetLogID      string   `json:"set_log_id"`
	Reps          *int     `json:"reps"`
	Weight        *float64 `json:"weight"`
}

func (s *Service) newUpdateSetLogTool() agentTool {
	schema := shared.FunctionParameters{
		"type":     "object",
		"required": []string{"exercise_log_id", "set_log_id"},
		"properties": map[string]any{
			"exercise_log_id": map[string]any{
				"type":        "string",
				"description": "UUID of the exercise log that holds the set log.",
			},
			"set_log_id": map[string]any{
				"type":        "string",
				"description": "UUID of the set log to update.",
			},
			"reps": map[string]any{
				"type":        "integer",
				"description": "New repetitions count for the set.",
			},
			"weight": map[string]any{
				"type":        "number",
				"description": "New weight value for the set in the workout's unit.",
			},
		},
		"additionalProperties": false,
	}

	return agentTool{
		name: "update_set_log",
		definition: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.String("update_set_log"),
				Description: openai.String("Update actual performance data (reps or weight) for a specific logged set."),
				Parameters:  openai.F(schema),
				Strict:      openai.Bool(true),
			}),
		},
		handler: func(ctx context.Context, toolCtx agentToolContext, raw json.RawMessage) (string, error) {
			span, ctx := opentracing.StartSpanFromContext(ctx, "agentTool.update_set_log")
			defer span.Finish()
			annotateAgentToolSpan(span, toolCtx, "update_set_log")
			if len(raw) > 0 {
				span.SetTag("arguments.present", true)
			}

			if !toolCtx.workoutID.IsValid {
				return "", fmt.Errorf("workout context is required for update_set_log")
			}

			var args updateSetLogArgs
			if err := json.Unmarshal(raw, &args); err != nil {
				return "", fmt.Errorf("invalid arguments for update_set_log: %w", err)
			}

			exerciseLogID, err := domain.ParseID(args.ExerciseLogID)
			if err != nil {
				return "", fmt.Errorf("invalid exercise_log_id: %w", err)
			}

			setLogID, err := domain.ParseID(args.SetLogID)
			if err != nil {
				return "", fmt.Errorf("invalid set_log_id: %w", err)
			}

			dtoInput := dto.UpdateSetLogDTO{}
			if args.Reps != nil {
				dtoInput.Reps = utils.NewNullable(*args.Reps, true)
			}
			if args.Weight != nil {
				dtoInput.Weight = utils.NewNullable(float32(*args.Weight), true)
			}

			if !dtoInput.Reps.IsValid && !dtoInput.Weight.IsValid {
				return "", fmt.Errorf("at least one of reps or weight must be provided")
			}

			if _, err := s.UpdateSetLog(ctx, toolCtx.userID, toolCtx.workoutID.V, exerciseLogID, setLogID, dtoInput); err != nil {
				return "", fmt.Errorf("failed to update set log: %w", err)
			}

			return `{"status":"success"}`, nil
		},
	}
}

func (s *Service) executeTool(ctx context.Context, toolCtx agentToolContext, name string, arguments string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.executeTool")
	defer span.Finish()
	annotateAgentToolSpan(span, toolCtx, name)
	span.SetTag("arguments.present", len(arguments) > 0)

	s.ensureChatTools()

	tool, ok := s.chatTools[name]
	if !ok {
		span.SetTag("error", true)
		span.LogKV("event", "unknown_tool", "tool", name)
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	result, err := tool.handler(ctx, toolCtx, json.RawMessage(arguments))
	if err != nil {
		span.SetTag("error", true)
		span.LogKV("event", "tool_error", "tool", name, "error.object", err)
	}
	return result, err
}

func (s *Service) newChatCompletionParams(messages []openai.ChatCompletionMessageParamUnion, tools []openai.ChatCompletionToolParam, model string, stream bool) openai.ChatCompletionNewParams {
	params := openai.ChatCompletionNewParams{
		Model:    openai.String(model),
		Messages: openai.F(messages),
	}

	if len(tools) > 0 {
		params.Tools = openai.F(tools)
		params.ToolChoice = openai.F[openai.ChatCompletionToolChoiceOptionUnionParam](openai.ChatCompletionToolChoiceOptionAuto(openai.ChatCompletionToolChoiceOptionAutoAuto))
	}

	if stream {
		params.StreamOptions = openai.F(openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)})
	}

	return params
}

var _ openai_client.ChatClient = (*openai_client.Client)(nil)

type ChatAgentContext struct {
	UserID    domain.ID
	WorkoutID utils.Nullable[domain.ID]
}

func (s *Service) ChatAgentToolDefinitions() []openai.ChatCompletionToolParam {
	return s.chatToolDefinitions()
}

func (s *Service) ExecuteChatAgentTool(ctx context.Context, ctxData ChatAgentContext, name string, arguments string) (string, error) {
	toolCtx := agentToolContext{userID: ctxData.UserID, workoutID: ctxData.WorkoutID}
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.ExecuteChatAgentTool")
	defer span.Finish()
	annotateAgentToolSpan(span, toolCtx, name)
	span.SetTag("arguments.present", len(arguments) > 0)

	return s.executeTool(ctx, toolCtx, name, arguments)
}

func (s *Service) NewChatCompletionParams(messages []openai.ChatCompletionMessageParamUnion, tools []openai.ChatCompletionToolParam, model string, stream bool) openai.ChatCompletionNewParams {
	return s.newChatCompletionParams(messages, tools, model, stream)
}
