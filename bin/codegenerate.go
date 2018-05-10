package main

import (
	"github.com/function61/pi-security-module/pkg/codegen"
	"github.com/function61/pi-security-module/pkg/version/versioncodegen"
	"os"
)

//go:generate go run codegenerate.go

func main() {
	// normalize to root of the project
	panicIfError(os.Chdir(".."))

	file := &codegen.DomainFile{}
	panicIfError(codegen.DeserializeJsonFile("pkg/domain/domain.json", file))

	restStructsFile := &codegen.RestStructsFile{}
	panicIfError(codegen.DeserializeJsonFile("pkg/apitypes/apitypes.json", restStructsFile))

	panicIfError(versioncodegen.Generate())

	panicIfError(codegen.GenerateCommands())

	eventDefs, eventStructsAsGoCode := codegen.ProcessEvents(file)

	tplData := &codegen.TplData{
		GoPackage:                   "domain",
		StringEnums:                 codegen.ProcessStringEnums(file),
		StringConsts:                codegen.ProcessStringConsts(file),
		EventDefs:                   eventDefs,
		EventStructsAsGoCode:        eventStructsAsGoCode,
		RestStructsAsGoCode:         codegen.ProcessRestStructsAsGoCode(restStructsFile),
		RestStructsAsTypeScriptCode: codegen.ProcessRestStructsAsTypeScriptCode(restStructsFile),
	}

	panicIfError(codegen.WriteTemplateFile("pkg/domain/events.go", tplData, codegen.EventsTemplateGo))

	panicIfError(codegen.GenerateEnumsAndConsts(tplData))

	panicIfError(codegen.WriteTemplateFile("pkg/apitypes/apitypes.go", tplData, codegen.RestStructsTemplateGo))

	panicIfError(codegen.WriteTemplateFile("frontend/generated/apitypes.ts", tplData, codegen.RestStructsTemplateTypeScript))
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
