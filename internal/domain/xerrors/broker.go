package xerrors

import (
	"github.com/vishenosik/xerrors"

	"github.com/vishenosik/notify/pkg/rabbitmq"
)

func DropMessageError() error {
	return xerrors.NewLocalizableError(rabbitmq.ErrDropMessage, "broker.drop_message", nil)
}
