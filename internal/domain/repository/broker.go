package repository

import (
	"context"
	"errors"

	"github.com/vishenosik/notify/internal/domain/entity"
)

// ErrDrop tells a broker consumer to reject the current message without requeue.
var ErrDrop = errors.New("drop message")

// Handler processes one push-send command.
// Return nil to acknowledge the message.
// Return an error that matches ErrDrop to reject without requeue.
// Any other error may be retried by the broker.
type Handler func(ctx context.Context, command *entity.PushCommand) error

// Publisher enqueues one push-send command.
// The destination (exchange, queue, topic) is adapter configuration, not an argument.
type Publisher interface {
	Publish(ctx context.Context, command *entity.PushCommand) error
}

// Consumer delivers push-send commands until ctx is done or the broker fails.
// Consume blocks. Cancellation of ctx is a clean stop; Consume returns ctx.Err().
type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
}

// Broker is the full port. Adapters may implement both publishing and consuming.
type Broker interface {
	Publisher
	Consumer
}
