package entity

import (
	"time"
	"uuid"

	"github.com/hashicorp/go-multierror"

	"github.com/vishenosik/notify/internal/domain/xerrors"
)

type base struct {
	id        uuid.UUID
	createdAt time.Time
	updatedAt time.Time
}

type Base struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func newBase(opts ...func(*base)) base {
	now := time.Now().UTC()
	b := base{
		id:        uuid.NewV7(),
		createdAt: now,
		updatedAt: now,
	}
	for _, opt := range opts {
		opt(&b)
	}
	return b
}

func created(variants Base) func(*base) {
	return func(b *base) {
		b.id = variants.ID
		b.createdAt = variants.CreatedAt
		b.updatedAt = variants.UpdatedAt
	}
}

func (b *base) validate() error {
	var err *multierror.Error

	if b.ID() == uuid.Nil() {
		err = multierror.Append(err, xerrors.InvalidIDError())
	}

	if b.createdAt.IsZero() {
		err = multierror.Append(err, xerrors.CreatedAtRequiredError())
	}

	if b.updatedAt.IsZero() {
		err = multierror.Append(err, xerrors.UpdatedAtRequiredError())
	} else if b.createdAt.After(b.updatedAt) {
		err = multierror.Append(err, xerrors.CreatedAtAfterUpdatedAtError())
	}

	return err.ErrorOrNil()
}

func (b *base) update() {
	b.updatedAt = time.Now().UTC()
}

func (b *base) ID() uuid.UUID        { return b.id }
func (b *base) CreatedAt() time.Time { return b.createdAt }
func (b *base) UpdatedAt() time.Time { return b.updatedAt }
