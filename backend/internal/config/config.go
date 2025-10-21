package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"fitness-trainer/internal/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// LLMConfig holds LLM-related settings
type LLMConfig struct {
	OpenAI struct {
		Model   string `mapstructure:"model"`
		BaseURL string `mapstructure:"base_url"`
		APIKey  string `mapstructure:"api_key"`
	} `mapstructure:"openai"`
	ReasoningEffort string `mapstructure:"reasoning_effort"`
	Limits          struct {
		DailyTokenLimit int `mapstructure:"daily_token_limit"`
	} `mapstructure:"limits"`
}

// PromptGenerationConfig holds prompt generation settings
type PromptGenerationConfig struct {
	Debounce  time.Duration `mapstructure:"debounce"`
	Period    time.Duration `mapstructure:"period"`
	RateLimit struct {
		PerHour int `mapstructure:"per_hour"`
		Burst   int `mapstructure:"burst"`
	} `mapstructure:"rate_limit"`
}

// Config holds all runtime-configurable settings
type Config struct {
	mu sync.RWMutex

	LLM              LLMConfig              `mapstructure:"llm"`
	PromptGeneration PromptGenerationConfig `mapstructure:"prompt_generation"`
}

var (
	instance *Config
	once     sync.Once
)

// Get returns the singleton config instance
func Get() *Config {
	once.Do(func() {
		instance = &Config{}
		instance.loadDefaults()
	})
	return instance
}

// Initialize sets up viper to watch config file
func Initialize(configPath string) error {
	cfg := Get()

	viper.SetConfigFile(configPath)

	// Enable environment variable substitution
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")

	// Load initial config
	if err := cfg.reload(); err != nil {
		return fmt.Errorf("failed to load initial config: %w", err)
	}

	// Watch for config changes
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logger.Info(fmt.Sprintf("Config file changed: %s", e.Name))
		if err := cfg.reload(); err != nil {
			logger.Error(fmt.Sprintf("Failed to reload config: %v", err))
		} else {
			logger.Info("Config reloaded successfully")
		}
	})

	return nil
}

// loadDefaults sets default values for all config fields
func (c *Config) loadDefaults() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// LLM defaults
	c.LLM.OpenAI.Model = "openai/gpt-5-mini"
	c.LLM.ReasoningEffort = "medium"
	c.LLM.Limits.DailyTokenLimit = 50000

	// Prompt generation defaults
	c.PromptGeneration.Debounce = time.Second * 60
	c.PromptGeneration.Period = time.Second * 10
	c.PromptGeneration.RateLimit.PerHour = 5
	c.PromptGeneration.RateLimit.Burst = 5
}

// reload reads the config file and updates all values
func (c *Config) reload() error {
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Unmarshal the config into the struct
	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand environment variables in sensitive fields
	c.LLM.OpenAI.APIKey = os.Expand(c.LLM.OpenAI.APIKey, os.Getenv)

	logger.Info(fmt.Sprintf("Config updated - OpenAI Model: %s, Reasoning Effort: %s",
		c.LLM.OpenAI.Model, c.LLM.ReasoningEffort))

	return nil
}

// GetOpenAIModel returns the current OpenAI model
func (c *Config) GetOpenAIModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LLM.OpenAI.Model
}

// GetReasoningEffort returns the current reasoning effort level
func (c *Config) GetReasoningEffort() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LLM.ReasoningEffort
}

// GetLLMDailyTokenLimit returns the daily token cap per user
func (c *Config) GetLLMDailyTokenLimit() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.LLM.Limits.DailyTokenLimit <= 0 {
		return 0
	}
	return c.LLM.Limits.DailyTokenLimit
}

// GetPromptGenerationDebounce returns the prompt generation debounce duration
func (c *Config) GetPromptGenerationDebounce() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PromptGeneration.Debounce
}

// GetPromptGenerationPeriod returns the prompt generation period
func (c *Config) GetPromptGenerationPeriod() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PromptGeneration.Period
}

// GetPromptGenerationRatePerHour returns the rate limit per hour
func (c *Config) GetPromptGenerationRatePerHour() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PromptGeneration.RateLimit.PerHour
}

// GetPromptGenerationRateBurst returns the rate limit burst
func (c *Config) GetPromptGenerationRateBurst() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PromptGeneration.RateLimit.Burst
}
