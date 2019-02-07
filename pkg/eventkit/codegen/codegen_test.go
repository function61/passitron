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

func TestFlattenDatatype(t *testing.T) {
	person := &DatatypeDef{
		NameRaw: "object",
		Fields: map[string]*DatatypeDef{
			"Name": &DatatypeDef{NameRaw: "string"},
			"Age":  &DatatypeDef{NameRaw: "boolean"},
		},
	}

	flattened := flattenDatatype(person)
	assert.Assert(t, len(flattened) == 3)
	assert.EqualString(t, flattened[0].NameRaw, "object")
	assert.EqualString(t, flattened[1].NameRaw, "string")
	assert.EqualString(t, flattened[2].NameRaw, "boolean")
}
