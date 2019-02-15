package codegen

import (
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/pi-security-module/pkg/eventkit/codegen/codegentemplates"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Module struct {
	// config
	Id string

	// input files
	EventsSpecFile   string
	CommandsSpecFile string
	TypesFile        string

	// computed state
	EventsSpec *DomainFile
	Types      *ApplicationTypesDefinition
	Commands   *CommandSpecFile
}

func NewModule(
	id string,
	typesFile string,
	eventsSpecFile string,
	commandSpecFile string,
) *Module {
	return &Module{
		Id:               id,
		EventsSpecFile:   eventsSpecFile,
		CommandsSpecFile: commandSpecFile,
		TypesFile:        typesFile,
	}
}

type FileToGenerate struct {
	targetPath     string
	obtainTemplate func() (string, error)
}

func processModule(mod *Module, opts Opts) error {
	// should be ok with nil data
	mod.EventsSpec = &DomainFile{}
	mod.Types = &ApplicationTypesDefinition{}
	mod.Commands = &CommandSpecFile{}

	hasTypes := mod.TypesFile != ""
	hasEvents := mod.EventsSpecFile != ""
	hasCommands := mod.CommandsSpecFile != ""

	if hasEvents {
		if err := jsonfile.Read(mod.EventsSpecFile, mod.EventsSpec, true); err != nil {
			return err
		}
	}

	if hasTypes {
		if err := jsonfile.Read(mod.TypesFile, mod.Types, true); err != nil {
			return err
		}
	}

	if hasCommands {
		if err := jsonfile.Read(mod.CommandsSpecFile, mod.Commands, true); err != nil {
			return err
		}
		if err := mod.Commands.Validate(); err != nil {
			return err
		}
	}

	hasRestEndpoints := len(mod.Types.Endpoints) > 0

	// preprocessing
	eventDefs, eventStructsAsGoCode := ProcessEvents(mod.EventsSpec)

	uniqueTypes := mod.Types.UniqueDatatypesFlattened()
	typeDependencyModuleIds := uniqueModuleIdsFromDatatypes(uniqueTypes)
	typesDependOnTime := false
	typesDependOnBinary := false

	for _, datatype := range uniqueTypes {
		switch datatype.NameRaw {
		case "datetime":
			typesDependOnTime = true
		case "binary":
			typesDependOnBinary = true
		}
	}

	backendPath := func(file string) string {
		return "pkg/" + mod.Id + "/" + file
	}

	frontendPath := func(file string) string {
		return "frontend/generated/" + mod.Id + "_" + file
	}

	docPath := func(file string) string {
		return "docs/" + mod.Id + "/" + file
	}

	data := &TplData{
		ModuleId:                mod.Id,
		Opts:                    opts,
		TypesDependOnTime:       typesDependOnTime,
		TypesDependOnBinary:     typesDependOnBinary,
		TypeDependencyModuleIds: typeDependencyModuleIds, // other modules whose types this module's types have dependencies to
		DomainSpecs:             mod.EventsSpec,          // backwards compat
		CommandSpecs:            mod.Commands,            // backwards compat
		ApplicationTypes:        mod.Types,               // backwards compat
		StringEnums:             ProcessStringEnums(mod.Types.Enums),
		EventDefs:               eventDefs,
		EventStructsAsGoCode:    eventStructsAsGoCode,
	}

	maybeRenderOne := func(expr bool, path string, template string) error {
		if !expr {
			return nil
		}

		return ProcessFile(Inline(path, template), data)
	}

	docs := opts.AutogenerateModuleDocs

	return allOk([]error{
		maybeRenderOne(hasCommands, backendPath("commanddefinitions.gen.go"), codegentemplates.BackendCommandsDefinitions),
		maybeRenderOne(hasCommands, frontendPath("commands.ts"), codegentemplates.FrontendCommandDefinitions),
		maybeRenderOne(hasCommands && docs, docPath("commands.md"), codegentemplates.DocsCommands),
		maybeRenderOne(hasEvents, backendPath("events.gen.go"), codegentemplates.BackendEventDefinitions),
		maybeRenderOne(hasEvents && docs, docPath("events.md"), codegentemplates.DocsEvents),
		maybeRenderOne(hasRestEndpoints, frontendPath("endpoints.ts"), codegentemplates.FrontendRestEndpoints),
		maybeRenderOne(hasRestEndpoints, backendPath("restendpoints.gen.go"), codegentemplates.BackendRestEndpoints),
		maybeRenderOne(hasRestEndpoints && docs, docPath("rest_endpoints.md"), codegentemplates.DocsRestEndpoints),
		maybeRenderOne(hasTypes, backendPath("types.gen.go"), codegentemplates.BackendTypes),
		maybeRenderOne(hasTypes, frontendPath("types.ts"), codegentemplates.FrontendDatatypes),
		maybeRenderOne(hasTypes && docs, docPath("types.md"), codegentemplates.DocsTypes),
	})
}

type Opts struct {
	BackendModulePrefix    string // "github.com/myorg/myproject/pkg/"
	FrontendModulePrefix   string // "generated/"
	AutogenerateModuleDocs bool
}

func ProcessModules(modules []*Module, opts Opts) error {
	for _, mod := range modules {
		if err := processModule(mod, opts); err != nil {
			return err
		}
	}

	return nil
}

// companion file means that for each of these files their corresponding .template file
// exists and will be rendered which will end up as the filename given
func CompanionFile(targetPath string) FileToGenerate {
	return FileToGenerate{
		targetPath: targetPath,
		obtainTemplate: func() (string, error) {
			templateContent, readErr := ioutil.ReadFile(targetPath + ".template")
			if readErr != nil {
				return "", readErr
			}

			return string(templateContent), nil
		},
	}
}

func Inline(targetPath string, inline string) FileToGenerate {
	return FileToGenerate{
		targetPath: targetPath,
		obtainTemplate: func() (string, error) {
			return inline, nil
		},
	}
}

func ProcessFile(target FileToGenerate, data interface{}) error {
	templateContent, err := target.obtainTemplate()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target.targetPath), 0755); err != nil {
		return err
	}

	return WriteTemplateFile(target.targetPath, data, templateContent)
}
