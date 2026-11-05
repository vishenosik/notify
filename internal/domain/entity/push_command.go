package entity

import (
	"errors"
	"fmt"
	"strings"
	"uuid"

	"github.com/hashicorp/go-multierror"
)

var (
	ErrPushSubscriptionIDRequired = errors.New("push subscription id is required")
	ErrPushAPIKeyRequired         = errors.New("push api key is required")
)

// PushCommand is a broker message that asks notify to send one web push.
type PushCommand struct {
	subscriptionID uuid.UUID
	apiKey         apiKey
	message        PushMessage
}

// apiKey is a secret credential. It never appears in fmt output.
type apiKey string

func (k apiKey) String() string { return "[redacted]" }

func (k apiKey) GoString() string { return "[redacted]" }

func NewPushCommand(subscriptionID uuid.UUID, key string, message PushMessage) *PushCommand {
	return &PushCommand{
		subscriptionID: subscriptionID,
		apiKey:         apiKey(strings.TrimSpace(key)),
		message:        NewPushMessage(message.Title(), message.Body(), message.Data()),
	}
}

func (c *PushCommand) SubscriptionID() uuid.UUID { return c.subscriptionID }

func (c *PushCommand) APIKey() string { return string(c.apiKey) }

func (c *PushCommand) Message() PushMessage {
	return NewPushMessage(c.message.Title(), c.message.Body(), c.message.Data())
}

func (c *PushCommand) Validate() error {
	var err *multierror.Error

	if c.subscriptionID == uuid.Nil() {
		err = multierror.Append(err, ErrPushSubscriptionIDRequired)
	}
	if c.apiKey == "" {
		err = multierror.Append(err, ErrPushAPIKeyRequired)
	}
	if er := c.message.Validate(); er != nil {
		err = multierror.Append(err, er)
	}

	return err.ErrorOrNil()
}

func (c PushCommand) String() string {
	return fmt.Sprintf("PushCommand{subscriptionID:%s}", c.subscriptionID)
}

func (c PushCommand) GoString() string {
	return c.String()
}
