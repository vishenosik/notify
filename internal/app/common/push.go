package common

import "uuid"

type PushSubscriptionResult struct {
	ID uuid.UUID
}

type VAPIDPublicKeyResult struct {
	PublicKey string
}
