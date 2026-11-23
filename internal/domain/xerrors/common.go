package xerrors

import "github.com/vishenosik/xerrors"

var (
	ErrValidation = xerrors.ErrValidation
	ErrNotFound   = xerrors.ErrNotFound
	ErrAuth       = xerrors.ErrAuth
	ErrPermission = xerrors.ErrPermission
	ErrConflict   = xerrors.ErrConflict
	ErrDependency = xerrors.ErrDependency
	ErrRateLimit  = xerrors.ErrRateLimit
	ErrTimeout    = xerrors.ErrTimeout
	ErrInternal   = xerrors.ErrInternal
)

var (
	InvalidIDError               = xerrors.InvalidIDError
	CreatedAtRequiredError       = xerrors.CreatedAtRequiredError
	UpdatedAtRequiredError       = xerrors.UpdatedAtRequiredError
	CreatedAtAfterUpdatedAtError = xerrors.CreatedAtAfterUpdatedAtError
	InvalidNameError             = xerrors.InvalidNameError
)
