package prompt_generator_service

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

type completionProvider interface {
	CreateCompletion(ctx context.Context, userID domain.ID, systemPrompt, prompt string) (string, error)
}

const systemPrompt = `
Ты — генератор инструкций для ИИ-фитнес-тренера.

На входе ты получаешь настройки клиента в виде структурированных данных. Преобразуй их в краткие, логичные и чёткие текстовые инструкции для ИИ, который будет генерировать персональные силовые тренировки. Инструкции должны отразить цели клиента, уровень подготовки, режим тренировок, стиль тренировок, приоритеты, ограничения и желаемый уровень разнообразия тренировок.

ВНИМАНИЕ:
- Не пересказывай поля напрямую (не нужно писать "primary_goal: GOAL_STRENGTH")
- Формулируй описание так, как это сделал бы опытный тренер
- Пиши от третьего лица
- Не выдумывай ничего сверх полученных данных
- Если какие-то данные отсутствуют, просто не упоминай их
- В инструкции учти все данные, которые ты получил
- Учитывай уровень разнообразия (VarietyLevel) как фактор — чем выше значение, тем больше клиенту важно, чтобы тренировки не повторялись и содержали вариативность подходов, упражнений и стимулов
- Не добавляй оценочных суждений или мотивационных фраз
- Будь лаконичным и конкретным. Стиль — деловой, но дружелюбный
- В ответе укажи только текст инструкции, без дополнительных пояснений или форматирования
`

type Service struct {
	completionProvider completionProvider
}

func New(completionProvider completionProvider) *Service {
	return &Service{
		completionProvider: completionProvider,
	}
}

type generationSettings struct {
	BasePrompt              string
	VarietyLevel            int
	PrimaryGoal             string
	SecondaryGoals          []string
	ExperienceLevel         string
	DaysPerWeek             int
	SessionDurationMinutes  int
	Injuries                string
	PriorityMuscleGroupsIDs []domain.ID // TODO: конвертировать в строку
	WorkoutPlanType         string
}

func newGenerationSettings(settings domain.GenerationSettings) generationSettings {
	return generationSettings{
		BasePrompt:              settings.BasePrompt.V,
		VarietyLevel:            settings.VarietyLevel.V,
		PrimaryGoal:             settings.PrimaryGoal.String(),
		SecondaryGoals:          settings.SecondaryGoals,
		ExperienceLevel:         settings.ExperienceLevel.String(),
		DaysPerWeek:             settings.DaysPerWeek.V,
		SessionDurationMinutes:  settings.SessionDurationMinutes.V,
		Injuries:                settings.Injuries.V,
		PriorityMuscleGroupsIDs: settings.PriorityMuscleGroupsIDs,
		WorkoutPlanType:         settings.WorkoutPlanType.String(),
	}
}

func (s *Service) GeneratePrompt(ctx context.Context, settings domain.GenerationSettings) (domain.Prompt, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "prompt_generator_service.GeneratePrompt")
	defer span.Finish()

	settingsDTO := newGenerationSettings(settings)

	bytes, err := json.Marshal(settingsDTO)
	if err != nil {
		return domain.Prompt{}, fmt.Errorf("failed to marshal generation settings: %w", err)
	}

	prompt, err := s.completionProvider.CreateCompletion(
		ctx,
		settings.UserID,
		systemPrompt,
		string(bytes),
	)
	if err != nil {
		return domain.Prompt{}, err
	}

	return domain.NewPrompt(settings.UserID, prompt, settings.Hash), nil
}
