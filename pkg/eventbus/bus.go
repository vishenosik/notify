package eventbus

import (
	"context"
	"log/slog"
	"sync"
	"time"
	"uuid"
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

const defaultBufferSize = 100

type EventBus struct {
	ch chan DomainEvent

	mu   sync.RWMutex
	subs []*sub

	log *slog.Logger
}

type sub struct {
	ch chan DomainEvent
}

func NewEventBus() *EventBus {
	return &EventBus{
		ch:   make(chan DomainEvent, defaultBufferSize),
		subs: make([]*sub, 0, 1),
		log:  slog.Default(),
	}
}

func (b *EventBus) Run(ctx context.Context) {
	for event := range b.ch {
		b.publish(ctx, event)
	}
}

func (b *EventBus) publish(ctx context.Context, event DomainEvent) {
	b.log.Debug("new event received",
		slog.String("event_name", event.EventName()),
		slog.String("event_id", event.EventID().String()),
		slog.Time("event_occurred_at", event.OccurredAt()),
	)

	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, sub := range b.subs {
		select {
		case sub.ch <- event:
		case <-ctx.Done():
			return
		}
	}
}

func (b *EventBus) Close() error {
	close(b.ch)
	for _, sub := range b.subs {
		close(sub.ch)
	}
	return nil
}

func newSub() *sub {
	return &sub{
		ch: make(chan DomainEvent, defaultBufferSize),
	}
}

func (b *EventBus) Publish(event DomainEvent) error {
	go func() {
		b.ch <- event
	}()
	return nil
}

func (b *EventBus) Sub() *sub {
	sub := newSub()

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subs = append(b.subs, sub)
	return sub
}

func (s *sub) Listen() <-chan DomainEvent {
	return s.ch
}
