package codegen

type ProcessedStringEnumMember struct {
	Key     string
	GoKey   string
	GoValue string
}

type ProcessedStringConst struct {
	Key   string
	Value string
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
	GoPackage            string
	StringEnums          []ProcessedStringEnum
	StringConsts         []ProcessedStringConst
	EventStructsAsGoCode string
	EventDefs            []EventDefForTpl
}

type EventSpec struct {
	Event    string            `json:"event"`
	CtorArgs []string          `json:"ctor"`
	Fields   []*EventFieldSpec `json:"fields"`
}

type EventFieldSpec struct {
	Key  string            `json:"key"`
	Type EventFieldTypeDef `json:"type"`
}

type EventFieldTypeDef struct {
	Name string                         `json:"_"`
	Of   *EventFieldTypeDef             `json:"of"`
	Keys []EventFieldObjectFieldTypeDef `json:"keys"`
}

type EventFieldObjectFieldTypeDef struct {
	Key  string             `json:"key"`
	Type *EventFieldTypeDef `json:"type"`
}
