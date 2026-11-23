package middleware

import (
	"context"
	"testing"
	"uuid"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
	ifacemock "github.com/vishenosik/notify/internal/app/interface/mock"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

func TestUnaryAuthWebPushMissingKey(t *testing.T) {
	t.Parallel()

	interceptor := UnaryAuth("admin-secret", ifacemock.NewMockAPIKeyUseCase(t))
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: notifyv1.WebPush_Subscribe_FullMethodName,
	}, func(context.Context, any) (any, error) {
		t.Fatal("handler must not run")
		return nil, nil
	})
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestUnaryAuthWebPushInvalidKey(t *testing.T) {
	t.Parallel()

	keys := ifacemock.NewMockAPIKeyUseCase(t)
	keys.EXPECT().Authenticate(mock.Anything, "ntf_bad").Return(uuid.Nil(), xerrors.ErrAuth).Once()
	interceptor := UnaryAuth("admin-secret", keys)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(MetadataAPIKey, "ntf_bad"))
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: notifyv1.WebPush_GetVapidPublicKey_FullMethodName,
	}, func(context.Context, any) (any, error) {
		t.Fatal("handler must not run")
		return nil, nil
	})
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestUnaryAuthWebPushSuccess(t *testing.T) {
	t.Parallel()

	clientID := uuid.NewV7()
	keys := ifacemock.NewMockAPIKeyUseCase(t)
	keys.EXPECT().Authenticate(mock.Anything, "ntf_good").Return(clientID, nil).Once()
	interceptor := UnaryAuth("admin-secret", keys)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(MetadataAPIKey, "ntf_good"))
	called := false
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: notifyv1.WebPush_Unsubscribe_FullMethodName,
	}, func(ctx context.Context, _ any) (any, error) {
		called = true
		got, ok := ClientIDFromContext(ctx)
		require.True(t, ok)
		require.Equal(t, clientID, got)
		return "ok", nil
	})
	require.NoError(t, err)
	require.True(t, called)
}

func TestUnaryAuthClientRegistry(t *testing.T) {
	t.Parallel()

	info := &grpc.UnaryServerInfo{FullMethod: notifyv1.ClientRegistry_CreateClient_FullMethodName}

	t.Run("missing admin", func(t *testing.T) {
		t.Parallel()
		interceptor := UnaryAuth("admin-secret", ifacemock.NewMockAPIKeyUseCase(t))
		_, err := interceptor(context.Background(), nil, info, nopHandler(t))
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("wrong admin", func(t *testing.T) {
		t.Parallel()
		interceptor := UnaryAuth("admin-secret", ifacemock.NewMockAPIKeyUseCase(t))
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(MetadataAPIKey, "nope"))
		_, err := interceptor(ctx, nil, info, nopHandler(t))
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("caller key without admin is permission denied", func(t *testing.T) {
		t.Parallel()
		interceptor := UnaryAuth("admin-secret", ifacemock.NewMockAPIKeyUseCase(t))
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(MetadataAPIKey, "ntf_svc"))
		_, err := interceptor(ctx, nil, info, nopHandler(t))
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("caller key with wrong admin is permission denied", func(t *testing.T) {
		t.Parallel()
		interceptor := UnaryAuth("admin-secret", ifacemock.NewMockAPIKeyUseCase(t))
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			MetadataAPIKey, "wrong",
		))
		_, err := interceptor(ctx, nil, info, nopHandler(t))
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("valid admin", func(t *testing.T) {
		t.Parallel()
		interceptor := UnaryAuth("admin-secret", ifacemock.NewMockAPIKeyUseCase(t))
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(MetadataAPIKey, "admin-secret"))
		called := false
		_, err := interceptor(ctx, nil, info, func(context.Context, any) (any, error) {
			called = true
			return "ok", nil
		})
		require.NoError(t, err)
		require.True(t, called)
	})
}

func TestToStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "nil", err: nil, code: codes.OK},
		{name: "validation", err: xerrors.ErrValidation, code: codes.InvalidArgument},
		{name: "not found", err: xerrors.ErrNotFound, code: codes.NotFound},
		{name: "auth", err: xerrors.ErrAuth, code: codes.Unauthenticated},
		{name: "permission", err: xerrors.ErrPermission, code: codes.PermissionDenied},
		{name: "conflict", err: xerrors.ErrConflict, code: codes.AlreadyExists},
		{name: "internal", err: xerrors.ErrInternal, code: codes.Internal},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := toStatus(tc.err)
			if tc.err == nil {
				require.NoError(t, got)
				return
			}
			require.Equal(t, tc.code, status.Code(got))
		})
	}
}

func nopHandler(t *testing.T) grpc.UnaryHandler {
	t.Helper()
	return func(context.Context, any) (any, error) {
		t.Fatal("handler must not run")
		return nil, nil
	}
}
