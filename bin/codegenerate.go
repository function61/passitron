package main

import (
	"github.com/function61/pi-security-module/pkg/codegen"
	"github.com/function61/pi-security-module/pkg/version/versioncodegen"
	"io/ioutil"
	"os"
)

//go:generate go run codegenerate.go

func main() {
	// normalize to root of the project
	panicIfError(os.Chdir(".."))

	domainSpecs := &codegen.DomainFile{}
	panicIfError(codegen.DeserializeJsonFile("pkg/domain/domain.json", domainSpecs))

	applicationTypes := &codegen.ApplicationTypesDefinition{}
	panicIfError(codegen.DeserializeJsonFile("pkg/apitypes/apitypes.json", applicationTypes))

	panicIfError(versioncodegen.Generate())

	panicIfError(codegen.GenerateCommands())

	eventDefs, eventStructsAsGoCode := codegen.ProcessEvents(domainSpecs)

	data := &codegen.TplData{
		GoPackage:                   "domain",
		DomainSpecs:                 domainSpecs,
		ApplicationTypes:            applicationTypes,
		StringEnums:                 codegen.ProcessStringEnums(domainSpecs),
		StringConsts:                codegen.ProcessStringConsts(domainSpecs),
		EventDefs:                   eventDefs,
		EventStructsAsGoCode:        eventStructsAsGoCode,
		RestStructsAsGoCode:         codegen.ProcessRestStructsAsGoCode(applicationTypes),
		RestStructsAsTypeScriptCode: codegen.ProcessRestStructsAsTypeScriptCode(applicationTypes),
	}

	panicIfError(renderTemplateFile("pkg/domain/events.go", data))

	panicIfError(renderTemplateFile("pkg/domain/domain.go", data))

	panicIfError(renderTemplateFile("frontend/generated/domain.ts", data))

	panicIfError(renderTemplateFile("pkg/apitypes/apitypes.go", data))

	panicIfError(renderTemplateFile("frontend/generated/apitypes.ts", data))

	panicIfError(renderTemplateFile("frontend/generated/restapi.ts", data))
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
