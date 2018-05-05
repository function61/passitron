package main

import (
	"github.com/function61/pi-security-module/pkg/codegen"
)

//go:generate go run main.go version.go commands.go

func main() {
	file, err := codegen.DeserializeDomainFile("../pkg/domain/domain.json")
	panicIfError(err)
	if err != nil {
		panic(err)
	}

	panicIfError(genVersionFile())

	panicIfError(generateCommands())

	eventDefs, eventStructsAsGoCode := codegen.ProcessEvents(file)

	tplData := &codegen.TplData{
		GoPackage:            "domain",
		StringEnums:          codegen.ProcessStringEnums(file),
		StringConsts:         codegen.ProcessStringConsts(file),
		EventDefs:            eventDefs,
		EventStructsAsGoCode: eventStructsAsGoCode,
	}

	panicIfError(codegen.WriteTemplateFile("../pkg/domain/events.go", tplData, codegen.EventsTemplateGo))

	panicIfError(codegen.GenerateEnumsAndConsts(tplData))
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
