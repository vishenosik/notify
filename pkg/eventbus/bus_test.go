package eventbus

import (
	"context"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
)

// type DomainEvent interface {
// 	EventID() uuid.UUID
// 	EventName() string
// 	OccurredAt() time.Time
// 	// AggregateId identifies the aggregate the event belongs to.
// 	AggregateID() uuid.UUID
// }

type testEvent struct {
	id         uuid.UUID
	aggregate  uuid.UUID
	occurredAt time.Time
}

func (e testEvent) EventID() uuid.UUID     { return e.id }
func (e testEvent) EventName() string      { return "service.updated" }
func (e testEvent) OccurredAt() time.Time  { return e.occurredAt }
func (e testEvent) AggregateID() uuid.UUID { return e.aggregate }

func newTestEvent(aggregateId uuid.UUID) testEvent {
	return testEvent{
		id:         uuid.NewV7(),
		aggregate:  aggregateId,
		occurredAt: time.Now(),
	}
}

func TestPublishDeliversToSubscriber(t *testing.T) {
	t.Parallel()

	bus := startBus(t)
	s := bus.Sub()
	event := newTestEvent(uuid.NewV7())

	require.NoError(t, bus.Publish(event))
	got := receive(t, s.Listen())
	require.Equal(t, event.EventID(), got.EventID())
	require.Equal(t, event.AggregateID(), got.AggregateID())
	require.Equal(t, "service.updated", got.EventName())
}

func TestPublishFansOutToAllSubscribers(t *testing.T) {
	t.Parallel()

	bus := startBus(t)
	first := bus.Sub()
	second := bus.Sub()
	event := newTestEvent(uuid.NewV7())

	require.NoError(t, bus.Publish(event))

	require.Equal(t, event.EventID(), receive(t, first.Listen()).EventID())
	require.Equal(t, event.EventID(), receive(t, second.Listen()).EventID())
}

func TestPublishDeliversAllEvents(t *testing.T) {
	t.Parallel()

	bus := startBus(t)
	s := bus.Sub()

	want := map[uuid.UUID]struct{}{
		uuid.NewV7(): {},
		uuid.NewV7(): {},
		uuid.NewV7(): {},
	}
	for id := range want {
		require.NoError(t, bus.Publish(newTestEvent(id)))
	}

	got := make(map[uuid.UUID]struct{}, len(want))
	for range want {
		got[receive(t, s.Listen()).AggregateID()] = struct{}{}
	}
	require.Equal(t, want, got)
}

func TestSequentialPublishPreservesOrder(t *testing.T) {
	t.Parallel()

	bus := startBus(t)
	s := bus.Sub()

	ids := []uuid.UUID{
		uuid.NewV7(),
		uuid.NewV7(),
		uuid.NewV7(),
	}
	for _, id := range ids {
		require.NoError(t, bus.Publish(newTestEvent(id)))
		require.Equal(t, id, receive(t, s.Listen()).AggregateID())
	}
}

func TestPublishDoesNotBlock(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	done := make(chan struct{})
	go func() {
		defer close(done)
		require.NoError(t, bus.Publish(newTestEvent(uuid.NewV7())))
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked")
	}
}

func TestCloseStopsRunAndClosesSubscribers(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	s := bus.Sub()

	done := make(chan struct{})
	go func() {
		defer close(done)
		bus.Run(context.Background())
	}()

	require.NoError(t, bus.Close())

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not exit after Close")
	}

	_, ok := <-s.Listen()
	require.False(t, ok)
}

func TestCloseWithNoSubscribers(t *testing.T) {
	t.Parallel()

	require.NoError(t, NewEventBus().Close())
}

func TestCanceledContextSkipsBlockedSubscriber(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go bus.Run(ctx)

	s := bus.Sub()
	for range defaultBufferSize {
		require.NoError(t, bus.Publish(newTestEvent(uuid.NewV7())))
	}
	require.Eventually(t, func() bool {
		return len(s.ch) == defaultBufferSize
	}, time.Second, time.Millisecond)

	cancel()
	require.NoError(t, bus.Publish(newTestEvent(uuid.NewV7())))
	require.Eventually(t, func() bool {
		return len(s.ch) == defaultBufferSize && len(bus.ch) == 0
	}, time.Second, time.Millisecond)
}

func startBus(t *testing.T) *EventBus {
	t.Helper()

	bus := NewEventBus()
	go bus.Run(context.Background())
	return bus
}

func receive(t *testing.T, ch <-chan DomainEvent) DomainEvent {
	t.Helper()

	select {
	case ev, ok := <-ch:
		require.True(t, ok, "subscriber channel closed")
		return ev
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
		return nil
	}
}
