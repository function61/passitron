package codegen

import (
	"errors"
	"fmt"
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
	Command                string              `json:"command"`
	Title                  string              `json:"title"`
	CrudNature             string              `json:"crudNature"`
	AdditionalConfirmation string              `json:"additional_confirmation"`
	MiddlewareChain        string              `json:"chain"`
	CtorArgs               []string            `json:"ctor"`
	Fields                 []*CommandFieldSpec `json:"fields"`
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
	Key                string `json:"key"`
	Type               string `json:"type"`
	Optional           bool   `json:"optional"`
	HideIfDefaultValue bool   `json:"hideIfDefaultValue"`
	Help               string `json:"help"`
}

func (c *CommandFieldSpec) AsGoField() string {
	return fmt.Sprintf("%s %s `json:\"%s\"`", c.Key, c.AsGoType(), c.Key)
}

func (c *CommandFieldSpec) AsValidationSnippet() string {
	goType := c.AsGoType()

	if goType == "string" || goType == "password" {
		maxLen := 128

		if c.Type == "multiline" {
			maxLen = 4 * 1024
		}

		return fmt.Sprintf(
			`if x.%s == "" {
		return fieldEmptyValidationError("%s")
	}
	if len(x.%s) > %d {
		return fieldLengthValidationError("%s", %d)
	}`,
			c.Key,
			c.Key,
			c.Key,
			maxLen,
			c.Key,
			maxLen)
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

func (c *CommandSpec) MakeStruct() string {
	template := `type %s struct {
	%s
}`

	fieldLines := []string{}

	for _, field := range c.Fields {
		fieldLines = append(fieldLines, field.AsGoField())
	}

	return fmt.Sprintf(
		template,
		c.AsGoStructName(),
		strings.Join(fieldLines, "\n\t"))
}

func (c *CommandSpec) MakeValidation() string {
	validationSnippets := []string{}

	for _, field := range c.Fields {
		if !field.Optional {
			validationSnippets = append(validationSnippets, field.AsValidationSnippet())
		}
	}

	return strings.Join(validationSnippets, "\n\t")
}

func (c *CommandSpec) FieldsForTypeScript() string {
	fields := []string{}

	emptyString := `''`

	for _, fieldSpec := range c.Fields {
		fieldSerialized := ""

		fieldToTypescript := func(fieldSpec *CommandFieldSpec, tsKind string) string {
			defVal := emptyString
			for _, ctorArg := range c.CtorArgs {
				if ctorArg == fieldSpec.Key {
					defVal = fieldSpec.Key
					break
				}
			}

			return fmt.Sprintf(
				`{ Key: '%s', Required: %v, HideIfDefaultValue: %v, Kind: CommandFieldKind.%s, DefaultValueString: %s, Help: '%s' },`,
				fieldSpec.Key,
				!fieldSpec.Optional,
				fieldSpec.HideIfDefaultValue,
				tsKind,
				defVal,
				fieldSpec.Help)
		}

		switch fieldSpec.Type {
		case "text":
			fieldSerialized = fieldToTypescript(fieldSpec, "Text")
		case "multiline":
			fieldSerialized = fieldToTypescript(fieldSpec, "Multiline")
		case "password":
			fieldSerialized = fieldToTypescript(fieldSpec, "Password")
		case "checkbox":
			fieldSerialized = fieldToTypescript(fieldSpec, "Checkbox")
		case "integer":
			fieldSerialized = fieldToTypescript(fieldSpec, "Integer")
		default:
			panic(fmt.Errorf("Unsupported field type for UI: %s", fieldSpec.Type))
		}

		fields = append(fields, fieldSerialized)
	}

	return strings.Join(fields, "\n\t\t\t")
}

func (c *CommandSpec) CtorArgsForTypeScript() string {
	ctorArgs := []string{}

	for _, ctorArg := range c.CtorArgs {
		ctorArgs = append(ctorArgs, ctorArg+": string")
	}

	return strings.Join(ctorArgs, ", ")
}
