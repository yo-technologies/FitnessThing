package main

import (
	_ "embed"

	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"fitness-trainer/internal/app"
	"fitness-trainer/internal/clients/ratelimiter"
	"fitness-trainer/internal/db"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/repository"
	"fitness-trainer/internal/service"
	"fitness-trainer/internal/service/background"
	"fitness-trainer/internal/service/tools"
	"fitness-trainer/internal/telegram/token_parser"
	"fitness-trainer/internal/tracer"

	genai_client "fitness-trainer/internal/clients/gemini"
	openai_client "fitness-trainer/internal/clients/openai"
	s3_client "fitness-trainer/internal/clients/s3"
	prompt_generator_service "fitness-trainer/internal/service/prompt_generator"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/generative-ai-go/genai"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"

	apiOpts "google.golang.org/api/option"
)

func init() {
	logger.Init()
	godotenv.Load()
	log.SetOutput(io.Discard)
}

func loadPostgresURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_SSL_MODE"),
	)
}

func Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracer.MustSetup(
		ctx,
		tracer.WithServiceName("fitness-trainer"),
		tracer.WithCollectorEndpoint(os.Getenv("JAEGER_COLLECTOR_ENDPOINT")),
	)

	postgresURL := loadPostgresURL()

	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal(err.Error())
	}

	endpoint := os.Getenv("AWS_ENDPOINT")
	bucket := os.Getenv("AWS_S3_BUCKET")

	awsConfig := getAWSConfig(ctx)
	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
	s3ClientWrapper := s3_client.New(s3Client, bucket)

	contextManager := db.NewContextManager(pool)

	repo := repository.NewPGXRepository(contextManager)

	genaiClient, err := newGeminiClient(ctx)
	if err != nil {
		return err
	}

	openAIClient := newOpenAIClient()
	openAIClientWrapper := openai_client.New(openAIClient)
	openAIModel := os.Getenv("OPENAI_MODEL")
	if openAIModel == "" {
		return fmt.Errorf("OPENAI_MODEL environment variable is not set")
	}

	service := service.New(
		contextManager,
		s3ClientWrapper,
		repo,
		openAIClientWrapper,
		openAIModel,
		tools.New(nil), // tools will set service later to avoid circular dependency
	)

	telegramTokenParser := newTelegramTokenParser()

	promptGenerationDebounce := time.Second * 60
	promptGenerationPeriod := time.Second * 10

	promptGenerationQuota := throttled.RateQuota{
		MaxRate:  throttled.PerHour(5),
		MaxBurst: 5,
	}

	promptGenerationInmemmoryStore, err := memstore.NewCtx(65536)
	if err != nil {
		return fmt.Errorf("failed to create in memory store: %w", err)
	}

	promptGenerationRateLimiter, err := throttled.NewGCRARateLimiterCtx(
		promptGenerationInmemmoryStore,
		promptGenerationQuota,
	)
	if err != nil {
		return fmt.Errorf("failed to create rate limiter: %w", err)
	}

	promptGenerationRateLimiterWrapper := ratelimiter.New(promptGenerationRateLimiter)

	promptsClientWrapper := genai_client.New(
		genaiClient,
		os.Getenv("GENAI_MODEL_NAME"),
		nil,
	)

	promptGenerationService := prompt_generator_service.New(
		promptsClientWrapper,
		repo,
	)

	backgroundService := background.New(
		promptGenerationDebounce,
		repo,
		repo,
		promptGenerationService,
		promptGenerationRateLimiterWrapper,
	)

	scheduler, err := gocron.NewScheduler(
		gocron.WithLocation(time.UTC),
	)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	scheduler.NewJob(
		gocron.DurationJob(promptGenerationPeriod),
		gocron.NewTask(backgroundService.GeneratePrompts),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.JobOption(gocron.WithStartImmediately()),
	)

	scheduler.Start()

	app := app.New(
		service,
		telegramTokenParser,
		app.WithHTTPPathPrefix("/api"),
	)

	if err := app.Run(ctx); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

func newTelegramTokenParser() token_parser.TelegramTokenParser {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		logger.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	expireInStr := os.Getenv("TELEGRAM_TOKEN_EXPIRE_IN")
	if expireInStr == "" {
		logger.Fatal("TELEGRAM_TOKEN_EXPIRE_IN environment variable is not set")
	}
	expireIn, err := time.ParseDuration(expireInStr)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to parse TELEGRAM_TOKEN_EXPIRE_IN: %v", err))
	}

	return token_parser.NewTelegramTokenParser(
		botToken,
		expireIn,
	)
}

func getAWSConfig(ctx context.Context) aws.Config {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	customRegion := os.Getenv("AWS_REGION")

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(customRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)

	if err != nil {
		log.Fatal("Unable to load AWS config:", err)
	}

	return cfg
}

type ProxyRoundTripper struct {
	proxy  *url.URL
	apiKey string
}

func (t *ProxyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if t.proxy != nil {
		transport.Proxy = http.ProxyURL(t.proxy)
	}

	newReq := req.Clone(req.Context())
	q := newReq.URL.Query()
	q.Add("key", t.apiKey)
	newReq.URL.RawQuery = q.Encode()

	return transport.RoundTrip(newReq)
}

func loadProxyData() *url.URL {
	proxyURL := os.Getenv("PROXY_URL")
	proxyUser := os.Getenv("PROXY_USER")
	proxyPassword := os.Getenv("PROXY_PASSWORD")

	if proxyURL == "" {
		return nil
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		logger.Fatal(err.Error())
	}

	if proxyUser != "" && proxyPassword != "" {
		parsedURL.User = url.UserPassword(proxyUser, proxyPassword)
	}

	return parsedURL
}

func newHTTPClient(proxyURL *url.URL, apiKey string) *http.Client {
	return &http.Client{
		Transport: &ProxyRoundTripper{
			apiKey: apiKey,
			proxy:  proxyURL,
		},
		Timeout: 30 * time.Second,
	}
}

func newGeminiClient(ctx context.Context) (*genai.Client, error) {
	proxy := loadProxyData()

	return genai.NewClient(
		ctx,
		apiOpts.WithHTTPClient(newHTTPClient(proxy, os.Getenv("GENAI_API_KEY"))),
		apiOpts.WithAPIKey(os.Getenv("GENAI_API_KEY")),
	)
}

func newOpenAIClient() *openai.Client {
	proxyURL := loadProxyData()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		logger.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")

	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(&http.Client{
			Transport: &ProxyRoundTripper{
				proxy: proxyURL,
			},
			Timeout: 30 * time.Second,
		}),
	}

	if baseURL != "" {
		options = append(options, option.WithBaseURL(baseURL))
	}

	return openai.NewClient(options...)
}
