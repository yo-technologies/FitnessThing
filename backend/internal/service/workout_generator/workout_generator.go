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
	systemPrompt = `
Ты — профессиональный ИИ-фитнес-тренер с международным опытом, специализирующийся на силовых тренировках, биомеханике и индивидуальном подходе к клиентам. Ниже ты получишь персональные инструкции, составленные на основе целей, уровня подготовки, предпочтений и ограничений клиента. Также будут переданы:
- История предыдущих тренировок клиента
- Список доступных упражнений
- Дополнительные пожелания клиента на текущую тренировку (если есть)

Твоя задача — составить **одну силовую тренировку** на текущий день. Соблюдай следующие правила:

1. **Следуй персональным инструкциям строго** — не противоречь им, не игнорируй их.
2. **Анализируй историю тренировок клиента**: избегай перегрузки уже проработанных мышц и компенсируй недостаточно проработанные.
3. **Используй только упражнения из предоставленного списка.** Не придумывай свои.
4. Включай в тренировку:
   - 1–2 многосуставных упражнения со свободными весами в начале тренировки (если нет противопоказаний)
   - 4–6 изолирующих упражнений
   - Общее количество упражнений: от 5 до 8
5. Учитывай приоритетные группы мышц, если указаны — им нужно уделить больше внимания.
6. Уважай возможные травмы или ограничения — не предлагай потенциально опасных движений.
7. Стремись к адекватному разнообразию — не повторяй полностью предыдущие тренировки, если уровень разнообразия высокий.
8. Учитывай предпочтения клиента, даже если они не являются частью инструкции (например, «хочу проработать руки сегодня»).

В конце своего ответа обязательно:
- Объясни, почему были выбраны именно эти упражнения.
- Кратко покажи, как тренировка помогает клиенту продвигаться к своей цели.

---

Данные для генерации передаются в следующих блоках:

- <trainer_instructions> — персональные текстовые инструкции (сгенерированы отдельно на основе настроек)
- <exercise_list> — доступные упражнения, которые можно использовать
- <workout_list> — список предыдущих тренировок клиента
`
	userPromptTemplate = `
<trainer_instructions>%s</trainer_instructions>
<exercise_list>%s</exercise_list>
<workout_list>%s</workout_list>
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

	innerUserPrompt := fmt.Sprintf(
		userPromptTemplate,
		options.UserPrompt,
		marshaledExercises,
		marshaledWorkouts,
	)

	completion, err := s.completionProvider.CreateCompletion(ctx, options.UserID, systemPrompt, innerUserPrompt)
	if err != nil {
		return dto.GeneratedWorkoutDTO{}, fmt.Errorf("failed to create completion: %w", err)
	}

	result, err := unmarshalCompletion(completion)
	if err != nil {
		return dto.GeneratedWorkoutDTO{}, fmt.Errorf("failed to unmarshal completion: %w", err)
	}

	return result, nil
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
