package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

const (
	MetadataAPIKey = "x-api-key"

	webPushServicePrefix        = "/notify.v1.WebPush/"
	clientRegistryServicePrefix = "/notify.v1.ClientRegistry/"
)

func UnaryAuth(adminKey string, keys interfaces.APIKeyUseCase) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		switch {
		case strings.HasPrefix(info.FullMethod, webPushServicePrefix):
			return authenticateCaller(ctx, req, keys, handler)

		case strings.HasPrefix(info.FullMethod, clientRegistryServicePrefix):
			return authenticateAdmin(ctx, req, adminKey, handler)

		default:
			return handler(ctx, req)
		}
	}
}

func authenticateCaller(
	ctx context.Context,
	req any,
	keys interfaces.APIKeyUseCase,
	handler grpc.UnaryHandler,
) (any, error) {
	apiKey := metadataValue(ctx, MetadataAPIKey)
	if apiKey == "" {
		return nil, status.Error(codes.Unauthenticated, "forbidden")
	}

	clientID, err := keys.Authenticate(ctx, apiKey)
	if err != nil {
		if errors.Is(err, xerrors.ErrAuth) {
			return nil, status.Error(codes.Unauthenticated, "forbidden")
		}
		return nil, toStatus(err)
	}

	return handler(WithClientID(ctx, clientID), req)
}

func authenticateAdmin(
	ctx context.Context,
	req any,
	adminKey string,
	handler grpc.UnaryHandler,
) (any, error) {
	gotAdmin := metadataValue(ctx, MetadataAPIKey)

	if gotAdmin == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if !secureEqual(gotAdmin, adminKey) || strings.TrimSpace(adminKey) == "" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	return handler(ctx, req)
}

func secureEqual(got, want string) bool {
	if subtle.ConstantTimeEq(int32(len(got)), int32(len(want))) != 1 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
