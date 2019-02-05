package codegen

import (
	"github.com/function61/gokit/assert"
	"testing"
)

func TestBeginsWithUppercaseLetter(t *testing.T) {
	mkDatatype := func(name string) *DatatypeDef {
		return &DatatypeDef{NameRaw: name}
	}

	assert.Assert(t, mkDatatype("Foo").isCustomType())
	assert.Assert(t, !mkDatatype("foo").isCustomType())

	assert.Assert(t, !mkDatatype("!perkele").isCustomType())
}
