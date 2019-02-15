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
	Events []*EventSpec `json:"events"`
}

type EventDefForTpl struct {
	EventKey        string
	CtorArgs        string
	CtorAssignments string
	GoStructName    string
}

// this is passed as data to each template that we'll render
type TplData struct {
	ModuleId                string
	Opts                    Opts
	TypesDependOnTime       bool
	TypesDependOnBinary     bool
	TypeDependencyModuleIds []string
	DomainSpecs             *DomainFile
	CommandSpecs            *CommandSpecFile
	ApplicationTypes        *ApplicationTypesDefinition
	StringEnums             []ProcessedStringEnum
	EventStructsAsGoCode    string
	EventDefs               []EventDefForTpl
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
