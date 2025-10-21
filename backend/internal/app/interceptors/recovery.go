package interceptors

import (
	"context"
	"fitness-trainer/internal/logger"
	"runtime/debug"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
)

func handlePanic(ctx context.Context, method string) {
	if r := recover(); r != nil {
		logger.Errorf("[interceptor.Recovery] method: %s; error: %v\n%s", method, r, debug.Stack())
		span := opentracing.SpanFromContext(ctx)
		if span == nil {
			return
		}
		ext.Error.Set(span, true)
		span.SetTag("error.message", r)
	}
}

func RecoveryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer handlePanic(ctx, info.FullMethod)
	return handler(ctx, req)
}

func RecoveryStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	defer handlePanic(ss.Context(), info.FullMethod)
	return handler(srv, ss)
}
