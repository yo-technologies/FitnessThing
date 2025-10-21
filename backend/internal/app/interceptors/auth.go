package interceptors

import (
	"context"
	"fmt"
	"strings"

	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/domain/dto"
	"fitness-trainer/internal/logger"
	"fitness-trainer/internal/utils"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	userIDKey contextKey = "auth-interceptor.user-id"
)

type UserService interface {
	GetOrCreateUser(ctx context.Context, dto dto.CreateUserDTO) (domain.User, error)
}

type TelegramTokenParser interface {
	Parse(token string) (domain.TelegramTokenData, error)
}

func TelegramAuthInterceptor(
	userService UserService,
	tokenParser TelegramTokenParser,
) func(context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, err := authenticateFromMetadata(ctx, userService, tokenParser, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func TelegramAuthStreamInterceptor(
	userService UserService,
	tokenParser TelegramTokenParser,
) func(any, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler) error {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := authenticateFromMetadata(ss.Context(), userService, tokenParser, info.FullMethod)
		if err != nil {
			return err
		}

		wrapped := &wrappedServerStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

func EnrichContextWithUserID(ctx context.Context, userID domain.ID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) (domain.ID, bool) {
	value, ok := ctx.Value(userIDKey).(domain.ID)
	return value, ok
}

func authenticateFromMetadata(
	ctx context.Context,
	userService UserService,
	tokenParser TelegramTokenParser,
	_ string,
) (context.Context, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "interceptors.Auth")
	defer span.Finish()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Errorf("metadata is not provided")
		return ctx, fmt.Errorf("metadata is not provided: %w", domain.ErrUnauthorized)
	}

	tokens, ok := md["authorization"]
	if !ok {
		logger.Errorf("authorization token is not provided")
		return ctx, fmt.Errorf("authorization token is not provided: %w", domain.ErrUnauthorized)
	}

	if len(tokens) != 1 {
		logger.Errorf("invalid token format")
		return ctx, fmt.Errorf("invalid token format: %w", domain.ErrUnauthorized)
	}

	token := strings.TrimSpace(strings.TrimPrefix(tokens[0], "tma "))

	parsedData, err := tokenParser.Parse(token)
	if err != nil {
		logger.Errorf("failed to parse token: %v", err)
		return ctx, fmt.Errorf("failed to parse token: %w", domain.ErrUnauthorized)
	}

	var dto dto.CreateUserDTO
	{
		dto.TelegramID = parsedData.User.ID
		dto.TelegramUsername = utils.NewNullable(parsedData.User.Username, parsedData.User.Username != "")
		dto.FirstName = utils.NewNullable(parsedData.User.FirstName, parsedData.User.FirstName != "")
		dto.LastName = utils.NewNullable(parsedData.User.LastName, parsedData.User.LastName != "")
		dto.ProfilePicURL = utils.NewNullable(parsedData.User.PhotoURL, parsedData.User.PhotoURL != "")
	}

	user, err := userService.GetOrCreateUser(ctx, dto)
	if err != nil {
		logger.Errorf("failed to get or create user: %v", err)
		return ctx, fmt.Errorf("failed to get or create user: %w", domain.ErrInternal)
	}

	ctx = context.WithValue(ctx, userIDKey, user.ID)
	span.SetTag("user_id", user.ID.String())

	return ctx, nil
}
