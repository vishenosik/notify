package repository

import (
	"github.com/vishenosik/notify/internal/domain/event"
)

type EventPub interface {
	Publish(event.DomainEvent) error
}

type EventSub interface {
	Listen() <-chan event.DomainEvent
}
