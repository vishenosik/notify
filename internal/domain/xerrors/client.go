package xerrors

import "github.com/vishenosik/xerrors"

func ClientNotFoundError() error {
	return xerrors.NotFound("client.not_found", nil)
}

func ClientNameRequiredError() error {
	return xerrors.Validation("client.name_required", nil)
}
