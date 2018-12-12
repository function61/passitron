package codegen

import (
	"github.com/function61/gokit/assert"
	"testing"
)

func TestBeginsWithUppercaseLetter(t *testing.T) {
	assert.True(t, isCustomType("Foo"))
	assert.True(t, !isCustomType("foo"))

	assert.True(t, !isCustomType("!perkele"))
}
