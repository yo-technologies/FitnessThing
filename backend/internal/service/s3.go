package service

import (
	"context"
	"strings"
)

func (s *Service) GenerateUploadURL(ctx context.Context, key string) (uploadURL string, getURL string, err error) {
	uploadURL, err = s.s3Client.GeneratePutPresignedURL(ctx, key)
	if err != nil {
		return "", "", err
	}

	getURL = strings.SplitN(uploadURL, "?", 2)[0]

	return uploadURL, getURL, nil
}
