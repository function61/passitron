package codegen

const DomainFileTemplateTypeScript = `
// WARNING: generated file

{{range .StringEnums}}
export enum {{.Name}} {
{{range .Members}}
	{{.Key}} = '{{.GoValue}}',{{end}}
}
{{end}}

{{range .StringConsts}}
export const {{.Key}} = '{{.Value}}';
{{end}}

`

const DomainFileTemplateGo = `package {{.GoPackage}}

// WARNING: generated file

{{range .StringEnums}}
const (
{{range .Members}}
	{{.GoKey}} = "{{.GoValue}}"{{end}}
)

// digest in name because there's no easy way to make exhaustive Enum pattern matching
// in Go, so we hack around it by calling this generated function everywhere we want
// to do the pattern match, and when enum members change the digest changes and thus
// it forces you to manually review and fix each place
func {{.Name}}Exhaustive{{.MembersDigest}}(in string) string {
	return in
}
{{end}}

{{range .StringConsts}}
const {{.Key}} = "{{.Value}}";
{{end}}

`

const EventsTemplateGo = `package domain

// WARNING: generated file

var eventBuilders = map[string]func() Event{
{{range .EventDefs}}
	"{{.EventKey}}": func() Event { return &{{.GoStructName}}{meta: &EventMeta{}} },{{end}}
}


{{.EventStructsAsGoCode}}


// constructors

{{range .EventDefs}}
func New{{.GoStructName}}({{.CtorArgs}}) *{{.GoStructName}} {
	return &{{.GoStructName}}{
		meta: &meta,
		{{.CtorAssignments}}
	}
}
{{end}}

{{range .EventDefs}}
func (e *{{.GoStructName}}) Meta() *EventMeta { return e.meta }{{end}}

{{range .EventDefs}}
func (e *{{.GoStructName}}) MetaType() string { return "{{.EventKey}}" }{{end}}

{{range .EventDefs}}
func (e *{{.GoStructName}}) Serialize() string { return e.meta.Serialize(e) }{{end}}

// interface

type EventListener interface { {{range .EventDefs}}
	Apply{{.GoStructName}}(*{{.GoStructName}}) error{{end}}

	HandleUnknownEvent(event Event) error
}

func DispatchEvent(event Event, listener EventListener) error {
	switch e := event.(type) { {{range .EventDefs}}
	case *{{.GoStructName}}:
		return listener.Apply{{.GoStructName}}(e){{end}}
	default:
		return listener.HandleUnknownEvent(event)
	}
}

`

const RestStructsTemplateGo = `package apitypes

import (
	"time"
)

type SecretKind string

{{range .RestStructsAsGoCode}}

{{.}}

{{end}}
`

const RestStructsTemplateTypeScript = `import {SecretKind} from 'generated/domain';

{{range .RestStructsAsTypeScriptCode}}

{{.}}

{{end}}
`
