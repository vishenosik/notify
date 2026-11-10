package xerrors

import "github.com/vishenosik/xerrors"

func APIKeyNotFoundError() error {
	return xerrors.NotFound("api_key.not_found", nil)
}

func APIKeyGenerationError() error {
	return xerrors.Internal("api_key.generation_error", nil)
}

func APIKeyHashRequiredError() error {
	return xerrors.Validation("api_key.hash_required", nil)
}

func APIKeyPrefixInvalidError() error {
	return xerrors.Validation("api_key.prefix_invalid", nil)
}

func APIKeyLast4InvalidError() error {
	return xerrors.Validation("api_key.last4_invalid", nil)
}
