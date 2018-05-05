package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

type CommandSpecFile []*CommandSpec

func (c *CommandSpecFile) Validate() error {
	for _, spec := range *c {
		if err := spec.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type CommandSpec struct {
	Command  string              `json:"command"`
	Title    string              `json:"title"`
	CtorArgs []string            `json:"ctor"`
	Fields   []*CommandFieldSpec `json:"fields"`
}

func (c *CommandSpec) AsGoStructName() string {
	// "user.Create" => "userCreate"
	dotRemoved := strings.Replace(c.Command, ".", "", -1)

	// "userCreate" => "UserCreate"
	titleCased := strings.Title(dotRemoved)

	return titleCased
}

func (c *CommandSpec) Validate() error {
	for _, field := range c.Fields {
		if err := field.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type CommandFieldSpec struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`
}

func (c *CommandFieldSpec) AsGoField() string {
	return fmt.Sprintf("%s %s `json:\"%s\"`", c.Key, c.AsGoType(), c.Key)
}

func (c *CommandFieldSpec) AsValidationSnippet() string {
	goType := c.AsGoType()

	if goType == "string" || goType == "password" {
		return fmt.Sprintf(
			`if x.%s == "" {
		return fieldEmptyValidationError("%s")
	}`,
			c.Key,
			c.Key)
	} else if goType == "bool" || goType == "int" {
		// presence check not possible for these types
		return ""
	} else {
		panic(errors.New("validation not supported for type: " + goType))
	}
}

func (c *CommandFieldSpec) AsGoType() string {
	goType := ""
	if c.Type == "text" {
		goType = "string"
	}
	if c.Type == "multiline" {
		goType = "string"
	}
	if c.Type == "password" {
		goType = "string"
	}
	if c.Type == "checkbox" {
		goType = "bool"
	}
	if c.Type == "integer" {
		goType = "int"
	}
	return goType
}

func (c *CommandFieldSpec) Validate() error {
	if c.Type == "" {
		c.Type = "text"
	}

	if c.AsGoType() == "" {
		return errors.New("field " + c.Key + " has invalid type: " + c.Type)
	}

	return nil
}

func makeStruct(spec *CommandSpec) string {
	template := `type %s struct {
	%s
}

func (x *%s) Validate() error {
	%s

	return nil
}
`

	fieldLines := []string{}
	validationSnippets := []string{}

	for _, field := range spec.Fields {
		fieldLines = append(fieldLines, field.AsGoField())

		if !field.Optional {
			validationSnippets = append(validationSnippets, field.AsValidationSnippet())
		}
	}

	return fmt.Sprintf(
		template,
		spec.AsGoStructName(),
		strings.Join(fieldLines, "\n\t"),
		spec.AsGoStructName(),
		strings.Join(validationSnippets, "\n\t"))
}

func writeCommandsMap(file *CommandSpecFile) error {
	template := `package commandhandlers

// WARNING: generated file

import (
	"errors"
	"github.com/function61/pi-security-module/pkg/command"
)

func fieldEmptyValidationError(fieldName string) error {
	return errors.New("field " + fieldName + " cannot be empty")
}

%s

var StructBuilders = map[string]func() command.Command{
%s
}


`
	structs := []string{}

	handlerLines := []string{}

	for _, commandSpec := range *file {
		structs = append(structs, makeStruct(commandSpec))

		handlerLine := fmt.Sprintf(
			`	"%s": func() command.Command { return &%s{} },`,
			commandSpec.Command,
			commandSpec.AsGoStructName())

		handlerLines = append(handlerLines, handlerLine)
	}

	commandsGenJsContent := fmt.Sprintf(
		template,
		strings.Join(structs, "\n\n"),
		strings.Join(handlerLines, "\n"))

	if writeErr := ioutil.WriteFile("pkg/commandhandlers/generated.go", []byte(commandsGenJsContent), 0777); writeErr != nil {
		return writeErr
	}

	return nil
}

func generateTypescript(file *CommandSpecFile) error {
	template := `import {CommandDefinition, CommandFieldKind} from 'types';

// WARNING: generated file

%s
`

	fnTemplate := `export function %s(%s): CommandDefinition {
	return {
		key: '%s',
		title: '%s',
		fields: [
%s
		],
	};
}`

	emptyString := `''`

	fns := []string{}

	for _, commandSpec := range *file {
		fields := []string{}

		ctorArgs := []string{}

		for _, ctorArg := range commandSpec.CtorArgs {
			ctorArgs = append(ctorArgs, ctorArg+": string")
		}

		for _, fieldSpec := range commandSpec.Fields {
			fieldSerialized := ""

			defVal := emptyString
			for _, ctorArg := range commandSpec.CtorArgs {
				if ctorArg == fieldSpec.Key {
					defVal = fieldSpec.Key
					break
				}
			}

			switch fieldSpec.Type {
			case "text":
				fieldSerialized = fmt.Sprintf(
					`{ Key: '%s', Kind: CommandFieldKind.Text, DefaultValueString: %s },`,
					fieldSpec.Key,
					defVal)
			case "multiline":
				fieldSerialized = fmt.Sprintf(
					`{ Key: '%s', Kind: CommandFieldKind.Multiline, DefaultValueString: %s },`,
					fieldSpec.Key,
					defVal)
			case "password":
				fieldSerialized = fmt.Sprintf(
					`{ Key: '%s', Kind: CommandFieldKind.Password, DefaultValueString: %s },`,
					fieldSpec.Key,
					defVal)
			case "checkbox":
				fieldSerialized = fmt.Sprintf(
					`{ Key: '%s', Kind: CommandFieldKind.Checkbox },`,
					fieldSpec.Key)
			case "integer":
				fieldSerialized = fmt.Sprintf(
					`{ Key: '%s', Kind: CommandFieldKind.Integer },`,
					fieldSpec.Key)
			default:
				return fmt.Errorf("Unsupported field type for UI: %s", fieldSpec.Type)
			}

			fields = append(fields, fieldSerialized)
		}

		fn := fmt.Sprintf(
			fnTemplate,
			commandSpec.AsGoStructName(),
			strings.Join(ctorArgs, ", "),
			commandSpec.Command,
			commandSpec.Title,
			strings.Join(fields, "\n\t\t\t"))

		fns = append(fns, fn)
	}

	fnsSerialized := fmt.Sprintf(
		template,
		strings.Join(fns, "\n\n"))

	if writeErr := ioutil.WriteFile("frontend/generated/commanddefinitions.ts", []byte(fnsSerialized), 0777); writeErr != nil {
		return writeErr
	}

	return nil
}

func generateCommands() error {
	contents, readErr := ioutil.ReadFile("commands.json")
	if readErr != nil {
		return readErr
	}

	var file CommandSpecFile
	if jsonErr := json.Unmarshal(contents, &file); jsonErr != nil {
		return jsonErr
	}

	if validationErr := file.Validate(); validationErr != nil {
		return validationErr
	}

	if err := writeCommandsMap(&file); err != nil {
		return err
	}

	if err := generateTypescript(&file); err != nil {
		return err
	}

	return nil
}
