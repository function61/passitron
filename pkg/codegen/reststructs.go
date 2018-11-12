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
	Path     string       `json:"path"`
	Name     string       `json:"name"`
	Produces DatatypeDef  `json:"produces"`
	Consumes *DatatypeDef `json:"consumes"`
}

// "/users/{id}" => "/users/${encodeURIComponent(id)}"
// "/search?q={query}" => "/search?query=${encodeURIComponent(query)}"
func (e *EndpointDefinition) TypescriptPath() string {
	replacements := []string{}

	for _, item := range routePlaceholderParseRe.FindAllStringSubmatch(e.Path, -1) {
		replacements = append(replacements,
			item[0],
			"${encodeURIComponent("+item[1]+")}")
	}

	return strings.NewReplacer(replacements...).Replace(e.Path)
}

var routePlaceholderParseRe = regexp.MustCompile("\\{([a-zA-Z0-9]+)\\}")

// "/users/{id}/addresses/{idx}" => "id: string, idx: string"
func (e *EndpointDefinition) TypescriptArgs() string {
	parsed := routePlaceholderParseRe.FindAllStringSubmatch(e.Path, -1)
	typescriptedArgs := []string{}

	for _, item := range parsed {
		typescriptedArgs = append(typescriptedArgs, item[1]+": string")
	}

	if e.Consumes != nil {
		typescriptedArgs = append(typescriptedArgs, "body: "+e.Consumes.AsTypeScriptType())
	}

	return strings.Join(typescriptedArgs, ", ")
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

func (s *StructDefinition) AsToGoCode() string {
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
