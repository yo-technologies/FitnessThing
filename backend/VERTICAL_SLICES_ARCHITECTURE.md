# Vertical Slices Architecture Migration

This document describes the migration from a layered architecture to a vertical slices architecture for the FitnessThing backend.

## What is Vertical Slices Architecture?

Vertical Slices Architecture organizes code by feature/use case rather than by technical layers. Each slice contains all the components needed to fulfill a specific business capability.

### Before (Layered Architecture)
```
internal/
├── app/                   # Presentation layer
├── service/              # Business logic layer  
├── repository/           # Data access layer
└── domain/               # Domain models
```

### After (Vertical Slices)
```
internal/
├── features/             # Feature-based organization
│   ├── user/            # User management slice
│   │   ├── handlers/    # API handlers
│   │   ├── service/     # Business logic
│   │   ├── repository/  # Data access
│   │   └── models/      # Feature-specific models
│   ├── workout/         # Workout management slice  
│   ├── exercise/        # Exercise management slice
│   └── routine/         # Routine management slice
└── shared/              # Shared components
    ├── domain/          # Common domain types
    ├── db/              # Database utilities
    └── clients/         # External service clients
```

## Benefits of Vertical Slices

1. **Better Cohesion**: All code for a feature is co-located
2. **Reduced Coupling**: Features don't share business logic layers
3. **Easier to Modify**: Changes to a feature are contained within its slice
4. **Parallel Development**: Teams can work on different slices independently
5. **Clear Boundaries**: Feature boundaries are explicit in the code structure

## Migration Strategy

The migration is being done incrementally to minimize risk:

### Phase 1: Infrastructure Setup ✅
- [x] Create new directory structure for vertical slices
- [x] Move shared components to `/internal/shared/`
- [x] Create user feature slice as example

### Phase 2: Feature Migration (In Progress)
- [ ] Complete user feature slice
- [ ] Migrate workout feature
- [ ] Migrate exercise feature  
- [ ] Migrate routine feature
- [ ] Migrate file feature

### Phase 3: Integration
- [ ] Update main application to use feature slices
- [ ] Remove old layered structure
- [ ] Update documentation and tests

## Current Implementation

### User Feature Slice (Example)
```
internal/features/user/
├── feature.go           # Feature assembly and registration
├── handlers/            # gRPC/HTTP handlers
│   ├── service.go
│   ├── get.go
│   └── update.go
├── service/             # Business logic
│   └── user_service.go
├── repository/          # Data access
│   ├── user_repository.go
│   └── generation_settings_repository.go
└── models/              # Feature-specific models (future)
```

### Shared Components
```
internal/shared/
├── domain/              # Common domain types (ID, errors, etc.)
├── db/                  # Database connection management
├── clients/             # External service clients (S3, AI services)
├── logger/              # Logging utilities
├── tracer/              # Tracing utilities
└── interceptors/        # gRPC interceptors
```

## Feature Boundaries

The application has been divided into these feature slices:

1. **User Management**: User profiles, authentication, settings
2. **Workout Generation**: AI-powered workout creation and management  
3. **Exercise Management**: Exercise catalog, alternatives, muscle groups
4. **Routine Management**: Workout templates and routine CRUD operations
5. **Workout Tracking**: Active workouts, exercise logging, set logging
6. **File Management**: File uploads and presigned URLs

## Integration Points

Each feature slice exposes:
- **Service Interface**: For use by other features
- **gRPC Handlers**: For external API access
- **Feature Assembly**: For dependency injection and setup

## Testing Strategy

With vertical slices, testing becomes more focused:
- **Unit Tests**: Test individual components within a slice
- **Integration Tests**: Test the entire slice end-to-end
- **Contract Tests**: Test interfaces between slices

## Development Workflow

1. Choose a feature to work on
2. Navigate to the feature slice directory
3. All related code (handlers, business logic, data access) is in one place
4. Make changes within the slice boundary
5. Test the entire feature slice

## Migration Progress

- [x] **Infrastructure**: Set up directory structure and shared components
- [x] **User Slice**: Created initial structure (handlers, service, repository)
- [ ] **Workout Slice**: Migrate workout-related functionality
- [ ] **Exercise Slice**: Migrate exercise-related functionality  
- [ ] **Routine Slice**: Migrate routine-related functionality
- [ ] **File Slice**: Migrate file-related functionality
- [ ] **Integration**: Update main application and remove old structure

## Benefits Realized

By adopting vertical slices, we expect to see:
- Faster feature development
- Easier onboarding for new developers  
- Clearer feature ownership
- Reduced merge conflicts
- Better testability
- Improved maintainability