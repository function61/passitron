package codegen

import (
	"github.com/function61/gokit/assert"
	"testing"
)

func TestBeginsWithUppercaseLetter(t *testing.T) {
	assert.Assert(t, isCustomType("Foo"))
	assert.Assert(t, !isCustomType("foo"))

	assert.Assert(t, !isCustomType("!perkele"))
}
