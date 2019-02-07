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

func NewModule(id string, typesFile string, eventsSpecFile string, commandSpecFile string) *Module {
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

	if mod.EventsSpecFile != "" {
		if err := jsonfile.Read(mod.EventsSpecFile, mod.EventsSpec, true); err != nil {
			return err
		}
	}

	if mod.TypesFile != "" {
		if err := jsonfile.Read(mod.TypesFile, mod.Types, true); err != nil {
			return err
		}
	}

	if mod.CommandsSpecFile != "" {
		if err := jsonfile.Read(mod.CommandsSpecFile, mod.Commands, true); err != nil {
			return err
		}
		if err := mod.Commands.Validate(); err != nil {
			return err
		}
	}

	// preprocessing
	eventDefs, eventStructsAsGoCode := ProcessEvents(mod.EventsSpec)

	uniqueTypes := mod.Types.UniqueDatatypesFlattened()
	typeDependencyModuleIds := uniqueModuleIdsFromDatatypes(uniqueTypes)
	typesDependOnTime := false

	for _, datatype := range uniqueTypes {
		if datatype.NameRaw == "datetime" {
			typesDependOnTime = true
		}
	}

	backendPath := func(file string) string {
		return "pkg/" + mod.Id + "/" + file
	}

	frontendPath := func(file string) string {
		return "frontend/generated/" + mod.Id + "_" + file
	}

	data := &TplData{
		ModuleId:                mod.Id,
		Opts:                    opts,
		TypesDependOnTime:       typesDependOnTime,
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

		return renderOneTemplate(Inline(path, template), data)
	}

	hasRestEndpoints := len(mod.Types.Endpoints) > 0

	return allOk([]error{
		maybeRenderOne(mod.CommandsSpecFile != "", backendPath("commanddefinitions.gen.go"), codegentemplates.BackendCommandsDefinitions),
		maybeRenderOne(mod.EventsSpecFile != "", backendPath("events.gen.go"), codegentemplates.BackendEventDefinitions),
		maybeRenderOne(hasRestEndpoints, backendPath("restendpoints.gen.go"), codegentemplates.BackendRestEndpoints),
		maybeRenderOne(mod.TypesFile != "", backendPath("types.gen.go"), codegentemplates.BackendTypes),
		maybeRenderOne(mod.TypesFile != "", frontendPath("types.ts"), codegentemplates.FrontendDatatypes),
		maybeRenderOne(hasRestEndpoints, frontendPath("endpoints.ts"), codegentemplates.FrontendRestEndpoints),
		maybeRenderOne(mod.CommandsSpecFile != "", frontendPath("commands.ts"), codegentemplates.FrontendCommandDefinitions),
	})
}

type Opts struct {
	BackendModulePrefix  string // "github.com/myorg/myproject/pkg/"
	FrontendModulePrefix string // "generated/"
}

func ProcessModules(modules []*Module, opts Opts) error {
	for _, mod := range modules {
		if err := processModule(mod, opts); err != nil {
			return err
		}
	}

	return nil
}

func ProcessFile(file FileToGenerate, data interface{}) error {
	return renderOneTemplate(file, data)
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

func renderOneTemplate(target FileToGenerate, data interface{}) error {
	templateContent, err := target.obtainTemplate()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target.targetPath), 0755); err != nil {
		return err
	}

	return WriteTemplateFile(target.targetPath, data, templateContent)
}
