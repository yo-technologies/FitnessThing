package workout_generator_service

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fmt"
	"math/rand"
	"slices"

	"github.com/opentracing/opentracing-go"
)

const (
	systemPromptTemplate = `
Ты профессиональный и всемирно известный фитнес-тренер, обладающий глубокими знаниями в области физиологии, биомеханики и диетологии. Твоя задача - внимательно проанализировать последние тренировки клиента и на основе этого анализа выбрать оптимальный набор упражнений для его текущей тренировки.
Обязательные условия: Мысленно сделай список упражнений, которые делал пользователь на предыдущих тренировках. В своих суждениях используй только этот список и не придумывай дополнительные упражнения. Используй только упражнения из предоставленного списка. Обеспечь проработку всех групп мышц тела, уделяя особое внимание тем группам, которые могли быть недостаточно проработаны в предыдущих тренировках. Включай в тренировочный план как упражнения со свободными весами, так и на тренажерах, отдавая предпочтение свободным весам в начале тренировки, так как они более энергозатратны. В тренировке должно быть 1, в редких случаях 2, основных упражнения со свободными весами, которые задействуют несколько групп мышц. Остальные упражнения должны быть изолированными, направленными на проработку конкретных мышц. Количество упражнений в тренировке должно быть не менее 5 и не более 8. Учитывай любые дополнительные пожелания клиента, отдавая им более высокий приоритет при выборе упражнений. 
Variety_level определяет уровень разнообразия тренировок: 1 - минимальное разнообразие, повторяющиеся тренировки; 2 - умеренное разнообразие; 3 - максимальное разнообразие упражнений, но сбалансированное по нагрузке. Base_user_prompt содержит цель, общие пожелания и возможные противопоказания пользователя.
Стремись к разнообразию в тренировках в соответствии с уровнем variety_level, сохраняя при этом их эффективность и безопасность. При составлении тренировочного плана также учитывай общие пожелания, цели клиента и возможные противопоказания, указанные в base_user_prompt.
В ответ так же включи пояснение к итоговому результату, объясни, почему ты выбрал именно эти упражнения и как они помогут клиенту достичь его целей.
<exercise_list>%s</exercise_list>
`
	userPromptTemplate = `
<workout_list>%s</workout_list>
<variety_level>%d</variety_level>
<base_user_prompt>%s</base_user_prompt>
<user_preferences>%s</user_preferences>
`
)

type CompletionProvider interface {
	CreateCompletion(ctx context.Context, userID domain.ID, systemPrompt, prompt string) (string, error)
}

type Service struct {
	completionProvider CompletionProvider
}

func New(completionProvider CompletionProvider) *Service {
	return &Service{
		completionProvider: completionProvider,
	}
}

func (s *Service) GenerateWorkout(ctx context.Context, options *dto.GenerateWorkoutOptions) (dto.GeneratedWorkoutDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "workout_generator_service.GenerateWorkout")
	defer span.Finish()

	marshaledWorkouts, err := marshalWorkouts(options.Workouts)
	if err != nil {
		return dto.GeneratedWorkoutDTO{}, fmt.Errorf("failed to marshal workouts: %w", err)
	}

	// Shuffle exercises to ensure variety
	slices.SortFunc(options.Exercises, func(a, b dto.SlimExerciseDTO) int {
		return rand.Intn(2)*2 - 1
	})

	marshaledExercises, err := marshalExercises(options.Exercises)
	if err != nil {
		return dto.GeneratedWorkoutDTO{}, fmt.Errorf("failed to marshal exercises: %w", err)
	}

	innerUserPrompt := fmt.Sprintf(userPromptTemplate, marshaledWorkouts, options.Settings.VarietyLevel.V, options.Settings.BasePrompt.V, options.UserPrompt)
	systemPrompt := fmt.Sprintf(systemPromptTemplate, marshaledExercises)

	completion, err := s.completionProvider.CreateCompletion(ctx, options.UserID, systemPrompt, innerUserPrompt)
	if err != nil {
		return dto.GeneratedWorkoutDTO{}, fmt.Errorf("failed to create completion: %w", err)
	}

	return unmarshalCompletion(completion)
}

func marshalWorkouts(workouts []dto.SlimWorkoutDTO) (string, error) {
	type exercise struct {
		Name string `json:"name"`
	}

	type workout struct {
		ID        string     `json:"id"`
		CreatedAt string     `json:"created_at"`
		Exercises []exercise `json:"exercises"`
	}

	workoutsToMarshal := make([]workout, 0, len(workouts))
	for _, w := range workouts {
		exercises := make([]exercise, 0, len(w.ExerciseNames))
		for _, e := range w.ExerciseNames {
			exercises = append(exercises, exercise{Name: e})
		}
		workoutsToMarshal = append(workoutsToMarshal, workout{
			ID:        w.ID.String(),
			CreatedAt: w.CreatedAt.String(),
			Exercises: exercises,
		})
	}

	return marshal(workoutsToMarshal)
}

func marshalExercises(exercises []dto.SlimExerciseDTO) (string, error) {
	type exercise struct {
		ID                 string   `json:"id"`
		Name               string   `json:"name"`
		TargetMuscleGroups []string `json:"targetMuscleGroups"`
	}

	exercisesToMarshal := make([]exercise, 0, len(exercises))
	for _, e := range exercises {
		exercisesToMarshal = append(exercisesToMarshal, exercise{
			ID:                 e.ID.String(),
			Name:               e.Name,
			TargetMuscleGroups: marshalMuscleGroups(e.TargetMuscleGroups),
		})
	}

	return marshal(exercisesToMarshal)
}

func marshalMuscleGroups(muscleGroups []domain.MuscleGroup) []string {
	muscleGroupsToMarshal := make([]string, 0, len(muscleGroups))
	for _, mg := range muscleGroups {
		muscleGroupsToMarshal = append(muscleGroupsToMarshal, mg.String())
	}

	return muscleGroupsToMarshal
}

func marshal(data interface{}) (string, error) {
	marshaledData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(marshaledData), nil
}

func unmarshalCompletion(rawCompletion string) (dto.GeneratedWorkoutDTO, error) {
	type exercise struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	type completion struct {
		Exercises []exercise `json:"exercises"`
		Reasoning string     `json:"reasoning"`
	}

	var completionData completion
	err := json.Unmarshal([]byte(rawCompletion), &completionData)
	if err != nil {
		return dto.GeneratedWorkoutDTO{}, err
	}

	exerciseIDs := make([]domain.ID, 0, len(completionData.Exercises))
	for _, e := range completionData.Exercises {
		parsedID, err := domain.ParseID(e.ID)
		if err != nil {
			return dto.GeneratedWorkoutDTO{}, fmt.Errorf("failed to parse exercise ID: %w", err)
		}
		exerciseIDs = append(exerciseIDs, parsedID)
	}

	return dto.GeneratedWorkoutDTO{
		ExerciseIDs: exerciseIDs,
		Reasoning:   completionData.Reasoning,
	}, nil
}
