package handlers

import (
	"context"
	desc "fitness-trainer/pkg/workouts"
)

type Service interface {
	GenerateUploadURL(ctx context.Context, key string) (uploadURL string, getURL string, err error)
}

type Implementation struct {
	service Service
	desc.UnimplementedFileServiceServer
}

func New(service Service) *Implementation {
	return &Implementation{
		service: service,
	}
}
