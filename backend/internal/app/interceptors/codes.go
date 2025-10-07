package interceptors

import (
	"context"
	"errors"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapDomainErrorToGRPCStatus(err error, method string) error {
	if errors.Is(err, domain.ErrNotFound) {
		return status.Errorf(codes.NotFound, "%s", err.Error())
	}
	if errors.Is(err, domain.ErrAlreadyExists) {
		return status.Errorf(codes.AlreadyExists, "%s", err.Error())
	}
	if errors.Is(err, domain.ErrInvalidArgument) {
		return status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		return status.Errorf(codes.Unauthenticated, "%s", err.Error())
	}
	if errors.Is(err, domain.ErrForbidden) {
		return status.Errorf(codes.PermissionDenied, "%s", err.Error())
	}
	if errors.Is(err, domain.ErrTooManyRequests) {
		return status.Errorf(codes.ResourceExhausted, "%s", err.Error())
	}

	logger.Errorf("[interceptor.Error] method: %s; error: %s", method, err.Error())
	return status.Error(codes.Internal, "internal server error")
}

func ErrCodesInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, mapDomainErrorToGRPCStatus(err, info.FullMethod)
	}

	return resp, err
}

func ErrCodesStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, ss)
	if err != nil {
		return mapDomainErrorToGRPCStatus(err, info.FullMethod)
	}

	return err
}
