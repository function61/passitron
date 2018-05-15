package codegen

import (
	"fmt"
	"regexp"
	"strings"
)

type ApplicationTypesDefinition struct {
	Structs   []StructDefinition   `json:"structs"`
	Endpoints []EndpointDefinition `json:"endpoints"`
}

type EndpointDefinition struct {
	Path     string      `json:"path"`
	Name     string      `json:"name"`
	Produces DatatypeDef `json:"produces"`
}

// "/users/{id}" => "/users/${id}"
func (e *EndpointDefinition) TypescriptPath() string {
	return strings.Replace(e.Path, "{", "${", -1)
}

var routePlaceholderParseRe = regexp.MustCompile("\\{([a-zA-Z0-9]+)\\}")

// "/users/{id}/addresses/{idx}" => "id: string, idx: string"
func (e *EndpointDefinition) TypescriptArgs() string {
	parsed := routePlaceholderParseRe.FindAllStringSubmatch(e.Path, -1)
	typescripted := []string{}

	for _, item := range parsed {
		typescripted = append(typescripted, item[1]+": string")
	}

	return strings.Join(typescripted, ", ")
}

type StructDefinition struct {
	Name   string                   `json:"name"`
	Fields []DatatypeDefObjectField `json:"fields"`
}

func (s *StructDefinition) AsTypeScriptCode() string {
	fieldsSerialized := []string{}

	for _, field := range s.Fields {
		fieldsSerialized = append(fieldsSerialized, field.Key+": "+field.Type.AsTypeScriptType()+";")
	}

	return fmt.Sprintf(`export interface %s {
	%s
}`,
		s.Name,
		strings.Join(fieldsSerialized, "\n\t"))
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

func (d *DatatypeDef) AsTypeScriptType() string {
	tsType := ""

	if d.Name == "string" {
		tsType = "string"
	} else if d.Name == "datetime" {
		tsType = "string"
	} else if d.Name == "boolean" {
		tsType = "boolean"
	} else if d.Name == "list" {
		tsType = d.Of.AsTypeScriptType() + "[]"
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

func ProcessRestStructsAsGoCode(file *ApplicationTypesDefinition) []string {
	ret := []string{}

	for _, struct_ := range file.Structs {
		ret = append(ret, StructToGoCode(&struct_))
	}

	return ret
}
