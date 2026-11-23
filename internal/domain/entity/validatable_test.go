package entity

import (
	"testing"

	"github.com/stretchr/testify/require"

	xerrors "github.com/vishenosik/notify/internal/domain/xerrors"
)

func TestNewValidated(t *testing.T) {
	t.Parallel()

	t.Run("valid value", func(t *testing.T) {
		t.Parallel()
		value := newTestStruct("test")
		got, err := NewValidated(value)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.True(t, got.IsValid())
		require.Equal(t, value, got.Value)
	})

	t.Run("invalid value", func(t *testing.T) {
		t.Parallel()
		got, err := NewValidated(newTestStruct(""))
		require.ErrorIs(t, err, xerrors.ErrValidation)
		require.Nil(t, got)
	})
}

func TestZeroValueValidatedIsInvalid(t *testing.T) {
	t.Parallel()

	var v Validated[*testStruct]
	require.False(t, v.IsValid())
}
