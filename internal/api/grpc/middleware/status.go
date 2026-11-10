package middleware

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vishenosik/notify/internal/domain/xerrors"
)

func UnaryStatus() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		return resp, toStatus(err)
	}
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}

	if st, ok := status.FromError(err); ok && st.Code() != codes.Unknown {
		return err
	}

	if merr, ok := errors.AsType[*multierror.Error](err); ok && len(merr.Errors) > 0 {
		err = merr.Errors[0]
	}

	switch {
	case errors.Is(err, xerrors.ErrValidation):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, xerrors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, xerrors.ErrAuth):
		return status.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, xerrors.ErrPermission):
		return status.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, xerrors.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
