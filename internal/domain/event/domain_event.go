//nolint:unused
package event

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent is implemented by every event raised in the domain layer.
// Events are facts: named in past tense and immutable once created.
type DomainEvent interface {
	EventID() uuid.UUID
	EventName() string
	OccurredAt() time.Time
	// AggregateId identifies the aggregate the event belongs to.
	AggregateID() uuid.UUID
}

// BaseEvent carries the fields shared by all domain events. Embed it and
// implement EventName on the concrete event.
type base struct {
	id         uuid.UUID
	aggregate  uuid.UUID
	occurredAt time.Time
}

func newBase(aggregateId uuid.UUID) base {
	return base{
		id:         uuid.Must(uuid.NewV7()),
		aggregate:  aggregateId,
		occurredAt: time.Now(),
	}
}

func (e base) EventID() uuid.UUID     { return e.id }
func (e base) AggregateID() uuid.UUID { return e.aggregate }
func (e base) OccurredAt() time.Time  { return e.occurredAt }
