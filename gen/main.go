package main

import (
	"github.com/function61/pi-security-module/pkg/codegen"
	"github.com/function61/pi-security-module/pkg/version/versioncodegen"
	"os"
)

//go:generate go run main.go commands.go

func main() {
	// normalize to root of the project
	panicIfError(os.Chdir(".."))

	file, err := codegen.DeserializeDomainFile("pkg/domain/domain.json")
	panicIfError(err)

	panicIfError(versioncodegen.Generate())

	panicIfError(generateCommands())

	eventDefs, eventStructsAsGoCode := codegen.ProcessEvents(file)

	tplData := &codegen.TplData{
		GoPackage:            "domain",
		StringEnums:          codegen.ProcessStringEnums(file),
		StringConsts:         codegen.ProcessStringConsts(file),
		EventDefs:            eventDefs,
		EventStructsAsGoCode: eventStructsAsGoCode,
	}

	panicIfError(codegen.WriteTemplateFile("pkg/domain/events.go", tplData, codegen.EventsTemplateGo))

	panicIfError(codegen.GenerateEnumsAndConsts(tplData))
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
