package common

import (
	"time"
	"uuid"
)

type ClientResult struct {
	ID           uuid.UUID
	Name         string
	Description  string
	MaskedAPIKey string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreatedClientResult struct {
	ClientResult
	Key string
}
