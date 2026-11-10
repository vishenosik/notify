package repository

import "context"

// Pinger is a named dependency the health-check endpoint can ping.
type Pinger interface {
	Name() string
	Ping(ctx context.Context) error
}
