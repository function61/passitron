package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const fileTemplate = `package domain

// WARNING: generated file

var eventBuilders = map[string]func() Event{
{{range .EventDefs}}
	"{{.EventKey}}": func() Event { return &{{.GoStructName}}{meta: &EventMeta{}} },{{end}}
}


{{.StructsAsGoCode}}


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

type GoStructField struct {
	Name string
	Type string
	Tags string
}

func (g *GoStructField) AsGoCode() string {
	return fmt.Sprintf(
		"%s %s `%s`",
		g.Name,
		g.Type,
		g.Tags)
}

type GoStruct struct {
	Name   string
	Fields []GoStructField
}

func (g *GoStruct) Field(name string) *GoStructField {
	for _, field := range g.Fields {
		if field.Name == name {
			return &field
		}
	}

	panic("field " + name + " not found")
}

func (g *GoStruct) AsGoCode() string {
	fieldsAsGoCode := []string{}

	template := `type %s struct {
	%s
}`

	for _, g := range g.Fields {
		fieldAsGoCode := g.AsGoCode()

		fieldsAsGoCode = append(fieldsAsGoCode, fieldAsGoCode)
	}

	return fmt.Sprintf(
		template,
		g.Name,
		strings.Join(fieldsAsGoCode, "\n\t"))
}

type Visitor struct {
	Structs []GoStruct
}

func (v *Visitor) AppendStruct(item GoStruct) {
	v.Structs = append(v.Structs, item)
}

func (v *Visitor) AsGoCode() string {
	structs := []string{}

	for _, item := range v.Structs {
		structs = append(structs, item.AsGoCode())
	}

	return strings.Join(structs, "\n\n")
}

type EventSpecFile []*EventSpec

type EventSpec struct {
	Event    string            `json:"event"`
	CtorArgs []string          `json:"ctor"`
	Fields   []*EventFieldSpec `json:"fields"`
}

func (e *EventSpec) VisitForGoStructs(visitor *Visitor) *GoStruct {
	eventFields := []GoStructField{
		GoStructField{Name: "meta", Type: "*EventMeta"},
	}

	for _, fieldSpec := range e.Fields {
		eventFields = append(eventFields, GoStructField{
			Name: fieldSpec.Key,
			Type: fieldSpec.Type.AsGoType(e.AsGoStructName()+fieldSpec.Key, visitor),
			Tags: "json:\"" + fieldSpec.Key + "\"",
		})
	}

	eventStruct := GoStruct{
		Name:   e.AsGoStructName(),
		Fields: eventFields,
	}

	visitor.AppendStruct(eventStruct)

	return &eventStruct
}

func (e *EventSpec) AsGoStructName() string {
	// "user.Created" => "userCreated"
	dotRemoved := strings.Replace(e.Event, ".", "", -1)

	// "userCreated" => "UserCreated"
	titleCased := strings.Title(dotRemoved)

	return titleCased
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

func (e *EventFieldTypeDef) AsGoType(parentGoName string, visitor *Visitor) string {
	switch e.Name {
	case "object":
		// create supporting structure to represent item
		supportStructDef := GoStruct{
			Name:   parentGoName + "Item",
			Fields: nil,
		}

		for _, objectKeyDefinition := range e.Keys {
			field := GoStructField{
				Name: objectKeyDefinition.Key,
				Type: objectKeyDefinition.Type.AsGoType(supportStructDef.Name, visitor),
				Tags: "json:\"" + objectKeyDefinition.Key + "\"",
			}

			supportStructDef.Fields = append(supportStructDef.Fields, field)
		}

		visitor.AppendStruct(supportStructDef)

		return supportStructDef.Name
	case "string":
		return "string"
	case "list":
		return "[]" + e.Of.AsGoType(parentGoName, visitor)
	default:
		panic("unsupported type: " + e.Name)
	}
}

type EventFieldObjectFieldTypeDef struct {
	Key  string             `json:"key"`
	Type *EventFieldTypeDef `json:"type"`
}

type EventDefForTpl struct {
	EventKey        string
	CtorArgs        string
	CtorAssignments string
	GoStructName    string
}

func generateEvents() error {
	eventsFile, openErr := os.Open("../pkg/domain/events.json")
	if openErr != nil {
		return openErr
	}

	var file EventSpecFile
	if jsonErr := json.NewDecoder(eventsFile).Decode(&file); jsonErr != nil {
		return jsonErr
	}

	structsVisitor := &Visitor{}

	eventDefs := []EventDefForTpl{}

	for _, eventSpec := range file {
		structForEvent := eventSpec.VisitForGoStructs(structsVisitor)

		ctorArgs := []string{}
		ctorAssignments := []string{}

		for _, ctorArg := range eventSpec.CtorArgs {
			ctorArgs = append(ctorArgs, ctorArg+" "+structForEvent.Field(ctorArg).Type)

			ctorAssignments = append(ctorAssignments, ctorArg+": "+ctorArg+",")
		}

		ctorArgs = append(ctorArgs, "meta EventMeta")

		eventDefs = append(eventDefs, EventDefForTpl{
			EventKey:        eventSpec.Event,
			GoStructName:    eventSpec.AsGoStructName(),
			CtorArgs:        strings.Join(ctorArgs, ", "),
			CtorAssignments: strings.Join(ctorAssignments, "\n\t\t"),
		})
	}

	eventsFileGenerated, errFile := os.Create("../pkg/domain/events.go")
	if errFile != nil {
		return errFile
	}

	defer eventsFileGenerated.Close()

	type TplData struct {
		StructsAsGoCode string
		EventDefs       []EventDefForTpl
	}

	tplParsed, _ := template.New("").Parse(fileTemplate)

	if err := tplParsed.Execute(eventsFileGenerated, TplData{
		EventDefs:       eventDefs,
		StructsAsGoCode: structsVisitor.AsGoCode(),
	}); err != nil {
		return err
	}

	return nil
}
