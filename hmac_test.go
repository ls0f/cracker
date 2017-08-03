package cracker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenHMACSHA1(t *testing.T) {
	assert.Equal(t, GenHMACSHA1("123456", "123456"), "74b55b6ab2b8e438ac810435e369e3047b3951d0")
}

func TestVerifyHMACSHA1(t *testing.T) {
	assert.True(t, VerifyHMACSHA1("123456", "123456", "74b55b6ab2b8e438ac810435e369e3047b3951d0"))
}

func TestVerifyHMACSHA12(t *testing.T) {

	assert.False(t, VerifyHMACSHA1("123456", "123456", "74b55b6ab2b8e438ac810435e369e304"))
}
