package entity

import (
	"slices"
	"strings"
	"uuid"

	"github.com/hashicorp/go-multierror"

	"github.com/vishenosik/notify/internal/domain/xerrors"
)

const (
	APIKeyPrefixLen   = 8
	APIKeyLast4Len    = 4
	APIKeyDefaultName = "default"
)

type APIKey struct {
	base
	clientID uuid.UUID
	name     string
	keyHash  []byte
	prefix   string
	last4    string
}

func NewAPIKey(variants APIKeyVariants, opts ...func(*APIKey)) *APIKey {
	k := &APIKey{
		base:     newBase(),
		clientID: variants.ClientID,
		name:     strings.TrimSpace(variants.Name),
		keyHash:  slices.Clone(variants.KeyHash),
		prefix:   variants.Prefix,
		last4:    variants.Last4,
	}
	for _, opt := range opts {
		opt(k)
	}
	return k
}

func CreatedAPIKey(variants Base) func(*APIKey) {
	return func(k *APIKey) {
		k.base = newBase(created(variants))
	}
}

func (k *APIKey) ClientID() uuid.UUID { return k.clientID }

func (k *APIKey) Name() string { return k.name }

func (k *APIKey) KeyHash() []byte { return slices.Clone(k.keyHash) }

func (k *APIKey) Prefix() string { return k.prefix }

func (k *APIKey) Last4() string { return k.last4 }

type ValidatedAPIKey = *Validated[*APIKey]

type APIKeyVariants struct {
	ClientID uuid.UUID
	Name     string
	KeyHash  []byte
	Prefix   string
	Last4    string
}

func (k APIKey) validate() error {
	var err *multierror.Error

	if k.clientID == uuid.Nil() {
		err = multierror.Append(err, xerrors.InvalidIDError())
	}
	if k.name == "" {
		err = multierror.Append(err, xerrors.InvalidNameError("api_key", k.name))
	}
	if len(k.keyHash) == 0 {
		err = multierror.Append(err, xerrors.APIKeyHashRequiredError())
	}
	if len(k.prefix) != APIKeyPrefixLen {
		err = multierror.Append(err, xerrors.APIKeyPrefixInvalidError())
	}
	if len(k.last4) != APIKeyLast4Len {
		err = multierror.Append(err, xerrors.APIKeyLast4InvalidError())
	}

	if er := k.base.validate(); er != nil {
		err = multierror.Append(err, er)
	}

	return err.ErrorOrNil()
}
