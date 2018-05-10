package codegen

import (
	"fmt"
	"strings"
)

type RestStructsFile struct {
	Structs []StructDefinition `json:"structs"`
}

type StructDefinition struct {
	Name   string                   `json:"name"`
	Fields []DatatypeDefObjectField `json:"fields"`
}

func StructToGoCode(s *StructDefinition) string {
	fields := []GoStructField{}

	visitor := &Visitor{}

	for _, field := range s.Fields {
		fields = append(fields, GoStructField{
			Name: field.Key,
			Type: AsGoType(field.Type, field.Key, visitor),
			Tags: "json:\"" + field.Key + "\"",
		})
	}

	structProcessed := GoStruct{
		Name:   s.Name,
		Fields: fields,
	}

	return structProcessed.AsGoCode()
}

func asTypeScriptType(d *DatatypeDef) string {
	tsType := ""

	if d.Name == "string" {
		tsType = "string"
	} else if d.Name == "datetime" {
		tsType = "string"
	} else if d.Name == "boolean" {
		tsType = "boolean"
	} else if d.Name == "list" {
		tsType = asTypeScriptType(d.Of) + "[]"
	} else if isCustomType(d.Name) {
		tsType = d.Name
	} else {
		panic("unknown type for TypeScript: " + d.Name)
	}

	if d.Nullable {
		tsType = tsType + " | null"
	}

	return tsType
}

func StructToTypeScriptCode(s *StructDefinition) string {
	fieldsSerialized := []string{}

	for _, field := range s.Fields {
		fieldsSerialized = append(fieldsSerialized, field.Key+": "+asTypeScriptType(field.Type)+";")
	}

	return fmt.Sprintf(`export interface %s {
	%s
}`,
		s.Name,
		strings.Join(fieldsSerialized, "\n\t"))
}

func ProcessRestStructsAsGoCode(file *RestStructsFile) []string {
	ret := []string{}

	for _, struct_ := range file.Structs {
		ret = append(ret, StructToGoCode(&struct_))
	}

	return ret
}

func ProcessRestStructsAsTypeScriptCode(file *RestStructsFile) []string {
	ret := []string{}

	for _, struct_ := range file.Structs {
		ret = append(ret, StructToTypeScriptCode(&struct_))
	}

	return ret
}
