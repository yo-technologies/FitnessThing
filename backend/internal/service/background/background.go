package background

import (
	"context"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	"time"

	"github.com/opentracing/opentracing-go"
)

type generationSettingsRepository interface {
	ListGenerationSettingsToProcess(ctx context.Context, debounceTime time.Duration) ([]domain.GenerationSettings, error)
}

type promptRepository interface {
	CreatePrompt(ctx context.Context, prompt domain.Prompt) (domain.Prompt, error)
}

type promptGenerator interface {
	GeneratePrompt(ctx context.Context, settings domain.GenerationSettings) (domain.Prompt, error)
}

type rateLimiter interface {
	Allow(ctx context.Context, userID domain.ID) (bool, error)
}

// Service is responsible for generating prompts based on generation settings.
type Service struct {
	debounceTime                 time.Duration
	generationSettingsRepository generationSettingsRepository
	promptRepository             promptRepository
	promptGenerator              promptGenerator
	rateLimiter                  rateLimiter
}

// New creates a new instance of the prompt generator service.
func New(
	debounceTime time.Duration,
	generationSettingsRepository generationSettingsRepository,
	promptRepository promptRepository,
	promptGenerator promptGenerator,
	rateLimiter rateLimiter,
) *Service {
	return &Service{
		debounceTime:                 debounceTime,
		generationSettingsRepository: generationSettingsRepository,
		promptRepository:             promptRepository,
		promptGenerator:              promptGenerator,
		rateLimiter:                  rateLimiter,
	}
}

// GeneratePrompts generates prompts based on the generation settings that need to be processed.
func (s *Service) GeneratePrompts() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	span, ctx := opentracing.StartSpanFromContext(ctx, "background.GeneratePrompts")
	defer span.Finish()

	logger.Debug("starting prompt generation")

	settingsList, err := s.generationSettingsRepository.ListGenerationSettingsToProcess(ctx, s.debounceTime)
	if err != nil {
		logger.Errorf("failed to list generation settings to process: %v", err)
		return
	}

	logger.Debugf("found %d generation settings to process", len(settingsList))

	for _, settings := range settingsList {
		logger.Debugf("processing generation settings for user %s", settings.UserID)

		allow, err := s.rateLimiter.Allow(ctx, settings.UserID)
		if err != nil {
			logger.Errorf("failed to check rate limit for user %s: %v", settings.UserID, err)
			continue
		}
		if !allow {
			logger.Warnf("rate limit exceeded for user %s, skipping prompt generation", settings.UserID)
			continue
		}

		logger.Debugf("generating prompt for settings %s", settings.ID)

		prompt, err := s.promptGenerator.GeneratePrompt(ctx, settings)
		if err != nil {
			logger.Errorf("failed to generate prompt for settings %s: %v", settings.ID, err)
			continue
		}

		logger.Debugf("generated prompt for settings %s: %s", settings.ID, prompt.PromptText)

		logger.Debugf("saving prompt for user %s", settings.UserID)

		if _, err := s.promptRepository.CreatePrompt(ctx, prompt); err != nil {
			logger.Errorf("failed to create prompt for settings %s: %v", settings.ID, err)
			continue
		}

		logger.Infof("successfully created prompt for user %s", settings.UserID)
	}
}
