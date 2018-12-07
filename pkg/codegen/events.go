package codegen

import (
	"fmt"
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

func VisitForGoStructs(e *EventSpec, visitor *Visitor) *GoStruct {
	eventFields := []GoStructField{
		GoStructField{Name: "meta", Type: "*event.EventMeta"},
	}

	for _, fieldSpec := range e.Fields {
		eventFields = append(eventFields, GoStructField{
			Name: fieldSpec.Key,
			Type: AsGoTypeWithInlineSupport(&fieldSpec.Type, EventNameAsGoStructName(e)+fieldSpec.Key, visitor),
			Tags: "json:\"" + fieldSpec.Key + "\"",
		})
	}

	eventStruct := GoStruct{
		Name:   EventNameAsGoStructName(e),
		Fields: eventFields,
	}

	visitor.AppendStruct(eventStruct)

	return &eventStruct
}

func EventNameAsGoStructName(e *EventSpec) string {
	// "user.Created" => "userCreated"
	dotRemoved := strings.Replace(e.Event, ".", "", -1)

	// "userCreated" => "UserCreated"
	titleCased := strings.Title(dotRemoved)

	return titleCased
}

func asGoTypeInternal(e *DatatypeDef, parentGoName string, visitor *Visitor) string {
	if isCustomType(e.Name) {
		return e.Name
	}

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
				Type: AsGoTypeWithInlineSupport(objectKeyDefinition.Type, supportStructDef.Name, visitor),
				Tags: "json:\"" + objectKeyDefinition.Key + "\"",
			}

			supportStructDef.Fields = append(supportStructDef.Fields, field)
		}

		visitor.AppendStruct(supportStructDef)

		return supportStructDef.Name
	case "integer":
		return "int"
	case "string":
		return "string"
	case "boolean":
		return "bool"
	case "datetime":
		return "time.Time"
	case "list":
		return "[]" + AsGoTypeWithInlineSupport(e.Of, parentGoName, visitor)
	default:
		panic("unsupported type: " + e.Name)
	}
}

func (e *DatatypeDef) AsGoType() string {
	return AsGoTypeWithInlineSupport(e, "", &Visitor{})
}

func AsGoTypeWithInlineSupport(e *DatatypeDef, parentGoName string, visitor *Visitor) string {
	typ := asGoTypeInternal(e, parentGoName, visitor)

	if e.Nullable {
		typ = "*" + typ
	}

	return typ
}

func ProcessEvents(file *DomainFile) ([]EventDefForTpl, string) {
	structsVisitor := &Visitor{}

	eventDefs := []EventDefForTpl{}

	for _, eventSpec := range file.Events {
		structForEvent := VisitForGoStructs(eventSpec, structsVisitor)

		ctorArgs := []string{}
		ctorAssignments := []string{}

		for _, ctorArg := range eventSpec.CtorArgs {
			ctorArgs = append(ctorArgs, ctorArg+" "+structForEvent.Field(ctorArg).Type)

			ctorAssignments = append(ctorAssignments, ctorArg+": "+ctorArg+",")
		}

		ctorArgs = append(ctorArgs, "meta event.EventMeta")

		eventDefs = append(eventDefs, EventDefForTpl{
			EventKey:        eventSpec.Event,
			GoStructName:    EventNameAsGoStructName(eventSpec),
			CtorArgs:        strings.Join(ctorArgs, ", "),
			CtorAssignments: strings.Join(ctorAssignments, "\n\t\t"),
		})
	}

	return eventDefs, structsVisitor.AsGoCode()
}
