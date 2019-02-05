package codegen

import (
	"strings"
)

type DatatypeDef struct {
	NameRaw  string                 `json:"_"` // "Type" | "module.Type" if referring to another module (such as "domain")
	Nullable bool                   `json:"nullable"`
	Of       *DatatypeDef           `json:"of"`     // only used if Name==list
	Fields   map[string]DatatypeDef `json:"fields"` // only used if Name==object
}

func (d *DatatypeDef) Name() string {
	// no module specified => return as -is
	if !strings.Contains(d.NameRaw, ".") {
		return d.NameRaw
	}

	comps := strings.Split(d.NameRaw, ".")

	return comps[1]
}

type DatatypeDefObjectField struct {
	Key  string       `json:"key"`
	Type *DatatypeDef `json:"type"`
}

// flattens datatype recursively, needed for example to know custom types inside arrays and objects
// def{Name:"object", Keys:{{Name: "foo", Type: Def{Name: "Foo"}}, {Name: "bar", Type: Def{Name: "Bar"}}}}
// => [def{"object"}, def{"Foo"}], def{"Bar"}
func flattenDatatype(def *DatatypeDef) []*DatatypeDef {
	all := []*DatatypeDef{}
	flattenDatatypeInternal(def, &all)
	return all
}

func flattenDatatypeInternal(def *DatatypeDef, all *[]*DatatypeDef) {
	*all = append(*all, def)

	if def.Name() == "list" {
		flattenDatatypeInternal(def.Of, all)
	} else if def.Name() == "object" {
		for _, member := range def.Fields {
			flattenDatatypeInternal(&member, all)
		}
	}
}

func isUppercase(input string) bool {
	return strings.ToLower(input) != input
}

func (d *DatatypeDef) isCustomType() bool {
	return isUppercase(d.Name()[0:1])
}
