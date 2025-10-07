package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

type contextKey string

// wrappedServerStream wraps grpc.ServerStream to allow context overriding
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
