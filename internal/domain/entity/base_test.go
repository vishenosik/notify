package entity

import (
	"testing"
	"time"
	"uuid"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"

	xerrors "github.com/vishenosik/notify/internal/domain/xerrors"
)

type testStruct struct {
	base
	name string
}

type testStructVariants struct {
	Name string
}

func newTestStruct(name string, opts ...func(*testStruct)) *testStruct {
	s := &testStruct{
		base: newBase(),
		name: name,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (ts *testStruct) validate() error {
	var errs *multierror.Error

	if ts.ID().Compare(uuid.Nil()) == 0 {
		errs = multierror.Append(errs, xerrors.InvalidIDError())
	}

	if ts.name == "" {
		errs = multierror.Append(errs, xerrors.InvalidNameError("test", "name"))
	}

	if er := ts.base.validate(); er != nil {
		errs = multierror.Append(errs, er)
	}

	return errs.ErrorOrNil()
}

func createdTestStruct(variants Base) func(*testStruct) {
	return func(s *testStruct) {
		s.base = newBase(created(variants))
	}
}

func (s *testStruct) Update(variants testStructVariants) error {
	s.name = variants.Name
	s.update()
	return s.validate()
}

func TestNewEntityAssignsUUIDv7AndTimestamps(t *testing.T) {
	t.Parallel()

	before := time.Now().UTC().Add(-time.Second)
	s := newTestStruct("test")
	after := time.Now().UTC().Add(time.Second)

	require.NotEqual(t, uuid.Nil(), s.ID())
	require.False(t, s.CreatedAt().Before(before))
	require.False(t, s.CreatedAt().After(after))
	require.Equal(t, s.CreatedAt(), s.UpdatedAt())
}

func TestUpdateAdvancesUpdatedAt(t *testing.T) {
	t.Parallel()

	past := time.Now().UTC().Add(-time.Hour)
	s := newTestStruct("test", createdTestStruct(Base{
		ID:        uuid.NewV7(),
		CreatedAt: past,
		UpdatedAt: past,
	}))

	require.NoError(t, s.Update(testStructVariants{
		Name: "test",
	}))
	require.True(t, s.UpdatedAt().After(past))
	require.Equal(t, past, s.CreatedAt())
}

func TestIdentityValidation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	id := uuid.NewV7()

	tests := []struct {
		name     string
		base     Base
		wantErrs []error
	}{
		{
			name: "valid",
			base: Base{ID: id, CreatedAt: now, UpdatedAt: now},
		},
		{
			name: "nil id",
			base: Base{CreatedAt: now, UpdatedAt: now},
			wantErrs: []error{
				xerrors.ErrValidation,
				xerrors.InvalidIDError(),
			},
		},
		{
			name: "zero created at",
			base: Base{ID: id, UpdatedAt: now},
			wantErrs: []error{
				xerrors.ErrValidation,
				xerrors.CreatedAtRequiredError(),
			},
		},
		{
			name: "zero updated at",
			base: Base{ID: id, CreatedAt: now},
			wantErrs: []error{
				xerrors.ErrValidation,
				xerrors.UpdatedAtRequiredError(),
			},
		},
		{
			name: "created after updated",
			base: Base{
				ID:        id,
				CreatedAt: now.Add(time.Second),
				UpdatedAt: now,
			},
			wantErrs: []error{
				xerrors.ErrValidation,
				xerrors.CreatedAtAfterUpdatedAtError(),
			},
		},
		{
			name: "fully invalid",
			base: Base{},
			wantErrs: []error{
				xerrors.ErrValidation,
				xerrors.InvalidIDError(),
				xerrors.CreatedAtRequiredError(),
				xerrors.UpdatedAtRequiredError(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := newTestStruct("test", createdTestStruct(tc.base))

			got, err := NewValidated(s)
			if len(tc.wantErrs) == 0 {
				require.NoError(t, err)
				require.True(t, got.IsValid())
				return
			}

			for _, wantErr := range tc.wantErrs {
				require.ErrorIs(t, err, wantErr)
			}
			require.Nil(t, got)
		})
	}
}
