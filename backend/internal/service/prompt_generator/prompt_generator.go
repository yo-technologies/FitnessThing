package prompt_generator_service

import (
	"context"
	"encoding/json"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/llm"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

type completionProvider interface {
	CreateCompletion(ctx context.Context, llmParams llm.ChatParams) (string, llm.Usage, error)
}

type muscleGroupRepository interface {
	GetMuscleGroupsByIDs(ctx context.Context, ids []domain.ID) ([]dto.MuscleGroupDTO, error)
}

type quotaService interface {
	Reserve(ctx context.Context, userID domain.ID, n int) (bool, error)
	Confirm(ctx context.Context, userID domain.ID, reserved int, actual int) error
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
- Стиль — деловой, но дружелюбный. Инструкции должны быть подробны и четко передавать пожелания клиента
- В ответе укажи только текст инструкции, без дополнительных пояснений или форматирования
`

type Service struct {
	completionProvider    completionProvider
	muscleGroupRepository muscleGroupRepository
	quotaService          quotaService
}

func New(completionProvider completionProvider, muscleGroupRepository muscleGroupRepository, quotaSvc quotaService) *Service {
	return &Service{
		completionProvider:    completionProvider,
		muscleGroupRepository: muscleGroupRepository,
		quotaService:          quotaSvc,
	}
}

type generationSettings struct {
	BasePrompt               string
	VarietyLevel             int
	PrimaryGoal              string
	SecondaryGoals           []string
	ExperienceLevel          string
	DaysPerWeek              int
	SessionDurationMinutes   int
	Injuries                 string
	PriorityMuscleGroupNames []string
	WorkoutPlanType          string
}

func (s *Service) newGenerationSettings(ctx context.Context, settings domain.GenerationSettings) (generationSettings, error) {
	// Получаем названия групп мышц по их ID
	muscleGroupNames, err := s.getMuscleGroupNamesByIDs(ctx, settings.PriorityMuscleGroupsIDs)
	if err != nil {
		return generationSettings{}, fmt.Errorf("failed to get muscle group names: %w", err)
	}

	return generationSettings{
		BasePrompt:               settings.BasePrompt.V,
		VarietyLevel:             settings.VarietyLevel.V,
		PrimaryGoal:              settings.PrimaryGoal.String(),
		SecondaryGoals:           settings.SecondaryGoals,
		ExperienceLevel:          settings.ExperienceLevel.String(),
		DaysPerWeek:              settings.DaysPerWeek.V,
		SessionDurationMinutes:   settings.SessionDurationMinutes.V,
		Injuries:                 settings.Injuries.V,
		PriorityMuscleGroupNames: muscleGroupNames,
		WorkoutPlanType:          settings.WorkoutPlanType.String(),
	}, nil
}

func (s *Service) getMuscleGroupNamesByIDs(ctx context.Context, ids []domain.ID) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}

	muscleGroups, err := s.muscleGroupRepository.GetMuscleGroupsByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get muscle groups: %w", err)
	}

	names := make([]string, len(muscleGroups))
	for i, group := range muscleGroups {
		names[i] = group.Name
	}

	return names, nil
}

func (s *Service) GeneratePrompt(ctx context.Context, settings domain.GenerationSettings) (domain.Prompt, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "prompt_generator_service.GeneratePrompt")
	defer span.Finish()

	settingsDTO, err := s.newGenerationSettings(ctx, settings)
	if err != nil {
		return domain.Prompt{}, fmt.Errorf("failed to prepare generation settings: %w", err)
	}

	bytes, err := json.Marshal(settingsDTO)
	if err != nil {
		return domain.Prompt{}, fmt.Errorf("failed to marshal generation settings: %w", err)
	}

	// Reserve 1 token upfront
	allowed, err := s.quotaService.Reserve(ctx, settings.UserID, 1)
	if err != nil {
		return domain.Prompt{}, err
	}
	if !allowed {
		return domain.Prompt{}, domain.ErrTooManyRequests
	}

	prompt, usage, err := s.completionProvider.CreateCompletion(ctx, llm.ChatParams{
		Messages: []llm.MessageParam{
			{Role: llm.RoleSystem, Content: systemPrompt},
			{Role: llm.RoleUser, Content: string(bytes)},
		},
	})
	if err != nil {
		return domain.Prompt{}, err
	}

	err = s.quotaService.Confirm(ctx, settings.UserID, 1, usage.TotalTokens)
	if err != nil {
		return domain.Prompt{}, fmt.Errorf("error confirming quota: %w", err)
	}

	return domain.NewPrompt(settings.UserID, prompt, settings.Hash), nil
}
