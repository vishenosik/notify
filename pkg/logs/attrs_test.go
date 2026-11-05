package logs

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_Error(t *testing.T) {
	ErrTest := errors.New("test error")
	result := Error(ErrTest)
	assert.Equal(t, AttrError, result.Key)
	assert.Equal(t, ErrTest.Error(), result.Value.String())
}
