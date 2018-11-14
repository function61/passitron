package main

import (
	"github.com/function61/pi-security-module/pkg/codegen"
	"github.com/function61/pi-security-module/pkg/version/versionresolver"
	"io/ioutil"
	"os"
)

//go:generate go run codegenerate.go

// name the code generation part to GUTS (Grand Unified Type System, i.e. a whack at GRUB)

func main() {
	// normalize to root of the project
	panicIfError(os.Chdir(".."))

	domainSpecs := &codegen.DomainFile{}
	panicIfError(codegen.DeserializeJsonFile("pkg/domain/domain.json", domainSpecs))

	applicationTypes := &codegen.ApplicationTypesDefinition{}
	panicIfError(codegen.DeserializeJsonFile("pkg/apitypes/apitypes.json", applicationTypes))

	commandSpecs := &codegen.CommandSpecFile{}
	panicIfError(codegen.DeserializeJsonFile("pkg/commandhandlers/commands.json", commandSpecs))
	panicIfError(commandSpecs.Validate())

	eventDefs, eventStructsAsGoCode := codegen.ProcessEvents(domainSpecs)

	data := &codegen.TplData{
		Version:              versionresolver.ResolveVersion(),
		DomainSpecs:          domainSpecs,
		CommandSpecs:         commandSpecs,
		ApplicationTypes:     applicationTypes,
		StringEnums:          codegen.ProcessStringEnums(domainSpecs),
		EventDefs:            eventDefs,
		EventStructsAsGoCode: eventStructsAsGoCode,
	}

	// for each of these files their corresponding .template file will be rendered which
	// will end up as the filename below
	files := []string{
		"docs/application_model/commands.md",
		"docs/application_model/datatypes.md",
		"docs/application_model/rest_endpoints.md",
		"docs/domain_model/consts.md",
		"docs/domain_model/events.md",
		"frontend/generated/apitypes.ts",
		"frontend/generated/commanddefinitions.ts",
		"frontend/generated/domain.ts",
		"frontend/generated/restapi.ts",
		"frontend/generated/version.ts",
		"pkg/apitypes/apitypes.go",
		"pkg/commandhandlers/generated.go",
		"pkg/domain/domain.go",
		"pkg/domain/events.go",
		"pkg/version/version.go",
	}

	for _, file := range files {
		panicIfError(renderTemplateFile(file, data))
	}
}

func renderTemplateFile(generatedPath string, data *codegen.TplData) error {
	templatePath := generatedPath + ".template"

	templateContent, readErr := ioutil.ReadFile(templatePath)
	if readErr != nil {
		return readErr
	}

	return codegen.WriteTemplateFile(generatedPath, data, string(templateContent))
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
