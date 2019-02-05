package codegen

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type ApplicationTypesDefinition struct {
	Structs   []StructDefinition   `json:"types"`
	Endpoints []EndpointDefinition `json:"endpoints"`
}

type EndpointDefinition struct {
	Path            string       `json:"path"`
	HttpMethod      string       `json:"method"`
	Name            string       `json:"name"`
	MiddlewareChain string       `json:"chain"`
	Description     string       `json:"description"`
	Produces        *DatatypeDef `json:"produces"` // optional
	Consumes        *DatatypeDef `json:"consumes"` // optional
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
	Name string      `json:"name"`
	Type DatatypeDef `json:"type"`
}

func (s *StructDefinition) AsTypeScriptCode() string {
	fieldsSerialized := []string{}

	for fieldKey, fieldType := range s.Type.Fields {
		fieldsSerialized = append(fieldsSerialized, fieldKey+": "+fieldType.AsTypeScriptType()+";")
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

	for fieldKey, fieldType := range s.Type.Fields {
		fields = append(fields, GoStructField{
			Name: fieldKey,
			Type: AsGoTypeWithInlineSupport(&fieldType, fieldKey, visitor),
			Tags: "json:\"" + fieldKey + "\"",
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

	if d.isCustomType() {
		tsType = d.Name()
	} else {
		switch d.Name() {
		case "string":
			tsType = "string"
		case "boolean":
			tsType = "boolean"
		case "datetime":
			tsType = "datetimeRFC3339"
		case "list":
			tsType = d.Of.AsTypeScriptType() + "[]"
		default:
			panic("unknown type for TypeScript: " + d.Name())
		}
	}

	if d.Nullable {
		tsType = tsType + " | null"
	}

	return tsType
}

func (a *ApplicationTypesDefinition) EndpointsProducesAndConsumesTypescriptTypes() []string {
	dedupe := map[string]bool{}

	processOneDt := func(dt *DatatypeDef) {
		if dt == nil {
			return
		}

		// look inside arrays and objects
		for _, flattenedItem := range flattenDatatype(dt) {
			if !flattenedItem.isCustomType() {
				continue
			}

			dedupe[flattenedItem.Name()] = true
		}
	}

	for _, endpoint := range a.Endpoints {
		processOneDt(endpoint.Consumes)
		processOneDt(endpoint.Produces)
	}

	uniques := make([]string, len(dedupe))

	i := 0
	for name, _ := range dedupe {
		uniques[i] = name
		i++
	}

	sort.Sort(sort.StringSlice(uniques))

	return uniques
}
