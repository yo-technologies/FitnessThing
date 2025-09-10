package handlers

import (
	"context"
	"fmt"

	"fitness-trainer/internal/shared/domain"
	"fitness-trainer/internal/logger"

	desc "fitness-trainer/pkg/workouts"

	"github.com/opentracing/opentracing-go"
)

func (i *Implementation) PresignUpload(ctx context.Context, in *desc.PresignUploadRequest) (*desc.PresignUploadResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.file.PresignUpload")
	defer span.Finish()

	if err := in.ValidateAll(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidArgument, err)
	}

	uploadURL, getURL, err := i.service.GenerateUploadURL(
		ctx,
		in.Filename,
	)
	if err != nil {
		logger.Errorf("error generating presigned URL: %v", err)
		return nil, err
	}

	return &desc.PresignUploadResponse{
		UploadUrl: uploadURL,
		GetUrl:    getURL,
	}, nil
}
