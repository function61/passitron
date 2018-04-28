package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type EventSpecFile []*EventSpec

func (e *EventSpecFile) Validate() error {
	for _, event := range *e {
		for _, field := range event.Fields {
			if field.AsGoType() == "" {
				return fmt.Errorf("Event %s field %s invalid type: %s", event.Event, field.Key, field.Type)
			}
		}
	}

	return nil
}

type EventSpec struct {
	Event    string            `json:"event"`
	CtorArgs []string          `json:"ctor"`
	Fields   []*EventFieldSpec `json:"fields"`
}

func (e *EventSpec) AsGoStructName() string {
	// "user.Created" => "userCreated"
	dotRemoved := strings.Replace(e.Event, ".", "", -1)

	// "userCreated" => "UserCreated"
	titleCased := strings.Title(dotRemoved)

	return titleCased
}

type EventFieldSpec struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

func (e *EventFieldSpec) AsGoType() string {
	switch e.Type {
	case "string":
		return "string"
	case "float32":
		return "float32"
	case "uint":
		return "uint"
	default:
		return "" // unrecognized type
	}
}

func generateEvents() error {
	contents, readErr := ioutil.ReadFile("domain/events.json")
	if readErr != nil {
		return readErr
	}

	var file EventSpecFile
	if jsonErr := json.Unmarshal(contents, &file); jsonErr != nil {
		return jsonErr
	}

	if validationErr := file.Validate(); validationErr != nil {
		return validationErr
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

	structTemplate := `type %s struct {
	meta *EventMeta
	%s
}`

	ctorTemplate := `func New%s(%s) *%s {
	return &%s{
		meta: &meta,
		%s
	}
}`

	structs := []string{}
	constructors := []string{}
	metaGetters := []string{}
	typeGetters := []string{}
	serializes := []string{}
	builderLines := []string{}

	for _, eventSpec := range file {
		goStructName := eventSpec.AsGoStructName()

		ctorArgs := []string{}
		ctorAssignments := []string{}

		metaGetters = append(metaGetters, fmt.Sprintf("func (e *%s) Meta() *EventMeta { return e.meta }", goStructName))
		typeGetters = append(typeGetters, fmt.Sprintf("func (e *%s) MetaType() string { return `%s` }", goStructName, eventSpec.Event))
		serializes = append(serializes, fmt.Sprintf("func (e *%s) Serialize() string { return e.meta.Serialize(e) }", goStructName))

		fields := []string{}

		for _, fieldSpec := range eventSpec.Fields {
			fieldAsGoCode := fmt.Sprintf(
				"%s %s `json:\"%s\"`",
				fieldSpec.Key,
				fieldSpec.AsGoType(),
				fieldSpec.Key)

			fields = append(fields, fieldAsGoCode)
		}

		for _, ctorArg := range eventSpec.CtorArgs {
			ctorArgs = append(ctorArgs, ctorArg+" string")

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

		eventStruct := fmt.Sprintf(structTemplate, goStructName, strings.Join(fields, "\n\t"))

		structs = append(structs, eventStruct)

		builderLine := fmt.Sprintf(
			`"%s": func() Event { return &%s{meta: &EventMeta{}} },`,
			eventSpec.Event,
			goStructName)

		builderLines = append(builderLines, builderLine)
	}

	content := fmt.Sprintf(
		template,
		strings.Join(builderLines, "\n\t"),
		strings.Join(structs, "\n\n\n"),
		strings.Join(constructors, "\n\n"),
		strings.Join(metaGetters, "\n"),
		strings.Join(typeGetters, "\n"),
		strings.Join(serializes, "\n"))

	if writeErr := ioutil.WriteFile("domain/events.go", []byte(content), 0777); writeErr != nil {
		panic(writeErr)
	}

	return nil
}
