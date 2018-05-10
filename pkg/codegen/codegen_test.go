package codegen

import (
	"github.com/function61/eventhorizon/util/ass"
	"testing"
)

func TestBeginsWithUppercaseLetter(t *testing.T) {
	ass.True(t, isCustomType("Foo"))
	ass.False(t, isCustomType("foo"))

	ass.False(t, isCustomType("!perkele"))
}
