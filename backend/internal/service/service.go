package service

import (
	"context"
	"sync"

	openai_client "fitness-trainer/internal/clients/openai"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
)

type workoutGenerator interface {
	GenerateWorkout(ctx context.Context, options *dto.GenerateWorkoutOptions) (dto.GeneratedWorkoutDTO, error)
}

type userRepository interface {
	GetUserByID(ctx context.Context, id domain.ID) (domain.User, error)
	GetOrCreateUser(ctx context.Context, user domain.User) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
}

type exerciseRepository interface {
	GetExercises(ctx context.Context, muscleGroups, excludedExercises []domain.ID) ([]domain.Exercise, error)
	GetExerciseByID(ctx context.Context, id domain.ID) (domain.Exercise, error)
	CreateExercise(ctx context.Context, exercise domain.Exercise, miscleGroupsIDs []domain.ID) (domain.Exercise, error)
}

type routineRepository interface {
	GetRoutines(ctx context.Context, userID domain.ID) ([]domain.Routine, error)
	CreateRoutine(ctx context.Context, routine domain.Routine) (domain.Routine, error)
	GetRoutineByID(ctx context.Context, id domain.ID) (domain.Routine, error)
	DeleteRoutine(ctx context.Context, id domain.ID) error
	UpdateRoutine(ctx context.Context, id domain.ID, routine domain.Routine) (domain.Routine, error)
}

type exerciseInstanceRepository interface {
	GetExerciseInstanceByID(ctx context.Context, id domain.ID) (domain.ExerciseInstance, error)
	GetExerciseInstancesByRoutineID(ctx context.Context, routineID domain.ID) ([]domain.ExerciseInstance, error)
	CreateExerciseInstance(ctx context.Context, exerciseInstance domain.ExerciseInstance) (domain.ExerciseInstance, error)
	DeleteExerciseInstance(ctx context.Context, id domain.ID) error
	SetExerciseOrder(ctx context.Context, routineID domain.ID, exerciseInstanceIDs []domain.ID) error
}

type muscleGroupRepository interface {
	GetMuscleGroups(ctx context.Context) ([]dto.MuscleGroupDTO, error)
	GetMuscleGroupByName(ctx context.Context, name string) (dto.MuscleGroupDTO, error)
}

type workoutRepository interface {
	GetWorkouts(ctx context.Context, userID domain.ID, limit, offset int) ([]domain.Workout, error)
	CreateWorkout(ctx context.Context, workout domain.Workout) (domain.Workout, error)
	GetWorkoutByID(ctx context.Context, id domain.ID) (domain.Workout, error)
	GetActiveWorkouts(ctx context.Context, userID domain.ID) ([]domain.Workout, error)
	UpdateWorkout(ctx context.Context, id domain.ID, workout domain.Workout) (domain.Workout, error)
	DeleteWorkout(ctx context.Context, id domain.ID) error
}

type exerciseLogRepository interface {
	GetExerciseLogsByWorkoutID(ctx context.Context, workoutID domain.ID) ([]domain.ExerciseLog, error)
	GetExerciseLogByID(ctx context.Context, id domain.ID) (domain.ExerciseLog, error)
	CreateExerciseLog(ctx context.Context, exerciseLog domain.ExerciseLog) (domain.ExerciseLog, error)
	GetExerciseLogsByExerciseIDAndUserID(ctx context.Context, exerciseID, userID domain.ID, offset, limit int) ([]domain.ExerciseLog, error)
	DeleteExerciseLog(ctx context.Context, id domain.ID) error
	UpdateExerciseLog(ctx context.Context, id domain.ID, exerciseLog domain.ExerciseLog) (domain.ExerciseLog, error)
}

type setLogRepository interface {
	GetSetLogsByExerciseLogID(ctx context.Context, exerciseLogID domain.ID) ([]domain.ExerciseSetLog, error)
	CreateSetLog(ctx context.Context, setLog domain.ExerciseSetLog) (domain.ExerciseSetLog, error)
	GetSetLogByID(ctx context.Context, id domain.ID) (domain.ExerciseSetLog, error)
	DeleteSetLog(ctx context.Context, id domain.ID) error
	UpdateSetLog(ctx context.Context, id domain.ID, setLog domain.ExerciseSetLog) (domain.ExerciseSetLog, error)
}

type setRepository interface {
	GetSetsByExerciseInstanceID(ctx context.Context, exerciseInstanceID domain.ID) ([]domain.Set, error)
	UpdateSet(ctx context.Context, id domain.ID, set domain.Set) (domain.Set, error)
	CreateSet(ctx context.Context, set domain.Set) (domain.Set, error)
	GetSetByID(ctx context.Context, id domain.ID) (domain.Set, error)
	DeleteSet(ctx context.Context, id domain.ID) error
}

type expectedSetRepository interface {
	GetExpectedSetsByExerciseLogID(ctx context.Context, exerciseLogID domain.ID) ([]domain.ExpectedSet, error)
	CreateExpectedSet(ctx context.Context, set domain.ExpectedSet) (domain.ExpectedSet, error)
	DeleteExpectedSetsByExerciseLogID(ctx context.Context, exerciseLogID domain.ID) error
}

type generationSettingsRepository interface {
	CreateOrUpdateGenerationSettings(ctx context.Context, settings domain.GenerationSettings) (domain.GenerationSettings, error)
	GetGenerationSettings(ctx context.Context, userID domain.ID) (domain.GenerationSettings, error)
}

type promptRepository interface {
	GetLastPromptByUserID(ctx context.Context, userID domain.ID) (domain.Prompt, error)
}

type chatRepository interface {
	CreateChat(ctx context.Context, chat domain.Chat) (domain.Chat, error)
	GetChatByWorkoutID(ctx context.Context, workoutID domain.ID) (domain.Chat, error)
	GetChatByID(ctx context.Context, chatID domain.ID) (domain.Chat, error)
	CreateChatMessage(ctx context.Context, message domain.ChatMessage) (domain.ChatMessage, error)
	ListChatMessages(ctx context.Context, chatID domain.ID, limit, offset int) ([]domain.ChatMessage, error)
}

type repository interface {
	userRepository
	exerciseRepository
	routineRepository
	exerciseInstanceRepository
	muscleGroupRepository
	workoutRepository
	exerciseLogRepository
	setLogRepository
	setRepository
	expectedSetRepository
	generationSettingsRepository
	promptRepository
	chatRepository
}

type unitOfWork interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	InTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

type s3Client interface {
	GeneratePutPresignedURL(ctx context.Context, key string) (string, error)
}

type generateWorkoutLimiter interface {
	Allow(ctx context.Context, userID domain.ID) (bool, error)
}

type Service struct {
	s3Client               s3Client
	workoutGenerator       workoutGenerator
	generateWorkoutLimiter generateWorkoutLimiter
	unitOfWork             unitOfWork
	repository             repository
	openAIClient           openai_client.ChatClient
	openAIModel            string
	chatTools              map[string]agentTool
	chatToolsOnce          sync.Once
}

func New(
	unitOfWork unitOfWork,
	s3Client s3Client,
	workoutGenerator workoutGenerator,
	generateWorkoutLimiter generateWorkoutLimiter,
	repository repository,
	openAIClient openai_client.ChatClient,
	openAIModel string,
) *Service {
	return &Service{
		unitOfWork:             unitOfWork,
		workoutGenerator:       workoutGenerator,
		s3Client:               s3Client,
		generateWorkoutLimiter: generateWorkoutLimiter,
		repository:             repository,
		openAIClient:           openAIClient,
		openAIModel:            openAIModel,
	}
}
