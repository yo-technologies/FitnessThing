package interceptors

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

func TracingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, info.FullMethod)
	defer span.Finish()

	result, err := handler(ctx, req)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err)
	}

	return result, err
}

func TracingStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	span, ctx := opentracing.StartSpanFromContext(ss.Context(), info.FullMethod)
	defer span.Finish()

	wrapped := &wrappedServerStream{ServerStream: ss, ctx: ctx}
	err := handler(srv, wrapped)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err)
	}

	return err
}
