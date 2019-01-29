package codegen

type ProcessedStringEnumMember struct {
	Key     string
	GoKey   string
	GoValue string
}

type ProcessedStringEnum struct {
	Name          string
	MembersDigest string
	Members       []ProcessedStringEnumMember
}

type EnumDef struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	StringMembers []string `json:"stringMembers"`
}

type StringConstDef struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DomainFile struct {
	StringConsts []StringConstDef `json:"stringConsts"`
	Enums        []EnumDef        `json:"enums"`
	Events       []*EventSpec     `json:"events"`
}

type EventDefForTpl struct {
	EventKey        string
	CtorArgs        string
	CtorAssignments string
	GoStructName    string
}

type TplData struct {
	Version              string
	DomainSpecs          *DomainFile
	CommandSpecs         *CommandSpecFile
	ApplicationTypes     *ApplicationTypesDefinition
	StringEnums          []ProcessedStringEnum
	EventStructsAsGoCode string
	EventDefs            []EventDefForTpl
}

type EventSpec struct {
	Event     string            `json:"event"`
	CtorArgs  []string          `json:"ctor"`
	Changelog []string          `json:"changelog"`
	Fields    []*EventFieldSpec `json:"fields"`
}

type EventFieldSpec struct {
	Key   string      `json:"key"`
	Type  DatatypeDef `json:"type"`
	Notes string      `json:"notes"`
}

type DatatypeDef struct {
	Name     string                   `json:"_"`
	Nullable bool                     `json:"nullable"`
	Of       *DatatypeDef             `json:"of"`
	Keys     []DatatypeDefObjectField `json:"keys"`
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

	if def.Name == "list" {
		flattenDatatypeInternal(def.Of, all)
	} else if def.Name == "object" {
		for _, member := range def.Keys {
			flattenDatatypeInternal(member.Type, all)
		}
	}
}
