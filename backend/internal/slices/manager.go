package slices

import (
	"fitness-trainer/internal/service"
	"fitness-trainer/internal/shared/db"
	
	// Import vertical slices (features)
	userFeature "fitness-trainer/internal/features/user"
	
	// Import existing structure for compatibility
	"fitness-trainer/internal/repository"
	
	"fitness-trainer/internal/shared/clients/ratelimiter"
	s3_client "fitness-trainer/internal/shared/clients/s3"
	workout_generator_service "fitness-trainer/internal/service/workout_generator"
)

// VerticalSlicesManager organizes the application into feature slices
type VerticalSlicesManager struct {
	// Feature slices
	userSlice *userFeature.UserFeature
	
	// Shared components
	contextManager *db.ContextManager
	
	// Combined service for backward compatibility
	combinedService *service.Service
}

// NewVerticalSlicesManager creates a new manager that organizes features into vertical slices
func NewVerticalSlicesManager(
	contextManager *db.ContextManager,
	s3Client s3_client.S3Client,
	workoutGenerator workout_generator_service.WorkoutGenerator,
	workoutGenerationRateLimiter ratelimiter.RateLimiter,
	repo repository.Repository,
) *VerticalSlicesManager {
	// Initialize individual feature slices
	userSlice := userFeature.NewUserFeature(contextManager, contextManager)
	
	// Create combined service for backward compatibility
	// This allows existing code to continue working while we transition to vertical slices
	combinedService := service.New(
		contextManager,
		s3Client,
		workoutGenerator,
		workoutGenerationRateLimiter,
		repo,
	)
	
	return &VerticalSlicesManager{
		userSlice:       userSlice,
		contextManager:  contextManager,
		combinedService: combinedService,
	}
}

// GetCombinedService returns the service that combines all features
// This maintains backward compatibility during the transition
func (m *VerticalSlicesManager) GetCombinedService() *service.Service {
	return m.combinedService
}

// GetUserSlice returns the user feature slice
func (m *VerticalSlicesManager) GetUserSlice() *userFeature.UserFeature {
	return m.userSlice
}

// Future: GetWorkoutSlice, GetExerciseSlice, etc. would be added here
// as we complete the migration of each feature to vertical slices

// Example of how a fully migrated application would work:
// func (m *VerticalSlicesManager) RegisterAllSlices(server *grpc.Server) {
//     m.userSlice.RegisterGRPC(server)
//     m.workoutSlice.RegisterGRPC(server)
//     m.exerciseSlice.RegisterGRPC(server)
//     m.routineSlice.RegisterGRPC(server)
//     m.fileSlice.RegisterGRPC(server)
// }