package repository

import (
	"fitness-trainer/internal/db"
)

type PGXRepository struct {
	contextManager *db.ContextManager
}

func NewPGXRepository(ctxManager *db.ContextManager) *PGXRepository {
	return &PGXRepository{
		contextManager: ctxManager,
	}
}
