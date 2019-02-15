package codegen

import (
	"sort"
	"strings"
)

type DatatypeDef struct {
	NameRaw  string                  `json:"_"` // "Type" | "module.Type" if referring to another module (such as "domain")
	Nullable bool                    `json:"nullable"`
	Of       *DatatypeDef            `json:"of"`     // only used if Name==list
	Fields   map[string]*DatatypeDef `json:"fields"` // only used if Name==object
}

type DatatypeDefField struct {
	Key  string
	Type *DatatypeDef
}

func (d *DatatypeDef) FieldsSorted() []DatatypeDefField {
	ret := make([]DatatypeDefField, len(d.Fields))
	i := 0
	for key, field := range d.Fields {
		ret[i] = DatatypeDefField{key, field}
		i++
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Key < ret[j].Key
	})

	return ret
}

func (d *DatatypeDef) Name() string {
	// no module specified => return as -is
	if !strings.Contains(d.NameRaw, ".") {
		return d.NameRaw
	}

	return strings.Split(d.NameRaw, ".")[1]
}

func (d *DatatypeDef) ModuleId() string {
	// no module specified
	if !strings.Contains(d.NameRaw, ".") {
		return ""
	}

	return strings.Split(d.NameRaw, ".")[0]
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
			memberCopy := member // stupid implicit pointers on looping
			flattenDatatypeInternal(memberCopy, all)
		}
	}
}

func isUppercase(input string) bool {
	return strings.ToLower(input) != input
}

func (d *DatatypeDef) isCustomType() bool {
	return isUppercase(d.Name()[0:1])
}

func uniqueModuleIdsFromDatatypes(dts []*DatatypeDef) []string {
	uniqueModuleIds := map[string]bool{}

	for _, dt := range dts {
		moduleId := dt.ModuleId()
		if moduleId == "" {
			continue
		}

		uniqueModuleIds[moduleId] = true
	}

	return stringBoolMapKeysSorted(uniqueModuleIds)
}

func stringBoolMapKeysSorted(sbm map[string]bool) []string {
	keys := make([]string, len(sbm))

	i := 0
	for key, _ := range sbm {
		keys[i] = key
		i++
	}

	sort.Sort(sort.StringSlice(keys))

	return keys
}
