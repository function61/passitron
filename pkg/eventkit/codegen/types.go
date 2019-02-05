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
