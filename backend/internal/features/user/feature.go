package user

import (
	"fitness-trainer/internal/features/user/handlers"
	"fitness-trainer/internal/features/user/repository"
	"fitness-trainer/internal/features/user/service"
	"fitness-trainer/internal/shared/db"
	desc "fitness-trainer/pkg/workouts"
	"google.golang.org/grpc"
)

type UserFeature struct {
	userService *service.UserService
	handlers    *handlers.Implementation
}

func NewUserFeature(contextManager *db.ContextManager, unitOfWork db.UnitOfWork) *UserFeature {
	// Repositories
	userRepo := repository.NewUserRepository(contextManager)
	generationSettingsRepo := repository.NewGenerationSettingsRepository(contextManager)

	// Service
	userService := service.NewUserService(userRepo, generationSettingsRepo, unitOfWork)

	// Handlers
	handlers := handlers.New(userService)

	return &UserFeature{
		userService: userService,
		handlers:    handlers,
	}
}

func (f *UserFeature) RegisterGRPC(server *grpc.Server) {
	desc.RegisterUserServiceServer(server, f.handlers)
}

func (f *UserFeature) GetService() *service.UserService {
	return f.userService
}