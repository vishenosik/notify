package xerrors

import "github.com/vishenosik/xerrors"

func PushEndpointRequiredError() error {
	return xerrors.Validation("push.endpoint_required", nil)
}

func PushP256dhRequiredError() error {
	return xerrors.Validation("push.p256dh_required", nil)
}

func PushAuthRequiredError() error {
	return xerrors.Validation("push.auth_required", nil)
}

func VAPIDNotConfiguredError() error {
	return xerrors.Validation("push.vapid_not_configured", nil)
}
