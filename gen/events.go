package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

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

func generateEvents() error {
	contents, readErr := ioutil.ReadFile("../pkg/domain/events.json")
	if readErr != nil {
		return readErr
	}

	var file EventSpecFile
	if jsonErr := json.Unmarshal(contents, &file); jsonErr != nil {
		return jsonErr
	}

	template := `package domain

// WARNING: generated file

// builder map

var eventBuilders = map[string]func() Event{
	%s
}

// structs
	
%s

// constructors

%s

// boilerplate functions

%s

%s

%s
`

	ctorTemplate := `func New%s(%s) *%s {
	return &%s{
		meta: &meta,
		%s
	}
}`

	constructors := []string{}
	metaGetters := []string{}
	typeGetters := []string{}
	serializes := []string{}
	builderLines := []string{}

	structsVisitor := &Visitor{}

	for _, eventSpec := range file {
		goStructName := eventSpec.AsGoStructName()

		structForEvent := eventSpec.VisitForGoStructs(structsVisitor)

		ctorArgs := []string{}
		ctorAssignments := []string{}

		metaGetters = append(metaGetters, fmt.Sprintf("func (e *%s) Meta() *EventMeta { return e.meta }", goStructName))
		typeGetters = append(typeGetters, fmt.Sprintf("func (e *%s) MetaType() string { return `%s` }", goStructName, eventSpec.Event))
		serializes = append(serializes, fmt.Sprintf("func (e *%s) Serialize() string { return e.meta.Serialize(e) }", goStructName))

		for _, ctorArg := range eventSpec.CtorArgs {
			ctorArgs = append(ctorArgs, ctorArg+" "+structForEvent.Field(ctorArg).Type)

			ctorAssignments = append(ctorAssignments, ctorArg+": "+ctorArg+",")
		}

		ctorArgs = append(ctorArgs, "meta EventMeta")

		constructors = append(constructors, fmt.Sprintf(
			ctorTemplate,
			goStructName,
			strings.Join(ctorArgs, ", "),
			goStructName,
			goStructName,
			strings.Join(ctorAssignments, "\n\t\t")))

		builderLine := fmt.Sprintf(
			`"%s": func() Event { return &%s{meta: &EventMeta{}} },`,
			eventSpec.Event,
			goStructName)

		builderLines = append(builderLines, builderLine)
	}

	content := fmt.Sprintf(
		template,
		strings.Join(builderLines, "\n\t"),
		structsVisitor.AsGoCode(),
		strings.Join(constructors, "\n\n"),
		strings.Join(metaGetters, "\n"),
		strings.Join(typeGetters, "\n"),
		strings.Join(serializes, "\n"))

	if writeErr := ioutil.WriteFile("../pkg/domain/events.go", []byte(content), 0777); writeErr != nil {
		panic(writeErr)
	}

	return nil
}
