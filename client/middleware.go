package client

import (
	"context"
	"strings"

	"github.com/vishenosik/notify/internal/api/grpc/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type apiKeyContextKey struct{}

// WithAPIKey overrides the client API key for this call (gRPC metadata and broker messages).
func WithAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, apiKeyContextKey{}, apiKey)
}

func apiKeyFromContext(ctx context.Context, fallback string) string {
	if v, ok := ctx.Value(apiKeyContextKey{}).(string); ok {
		return strings.TrimSpace(v)
	}
	return strings.TrimSpace(fallback)
}

func unaryAuthInterceptor(apikey string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if key := apiKeyFromContext(ctx, apikey); key != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, middleware.MetadataAPIKey, key)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
