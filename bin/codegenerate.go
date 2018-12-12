package main

import (
	"github.com/function61/gokit/dynversion/precompilationversion"
	"github.com/function61/pi-security-module/pkg/eventkit/codegen"
	"os"
)

//go:generate go run codegenerate.go

// name the code generation part to GUTS (Grand Unified Type System, i.e. a whack at GRUB)

func main() {
	if err := mainInternal(); err != nil {
		panic(err)
	}
}

func mainInternal() error {
	// normalize to root of the project
	if err := os.Chdir(".."); err != nil {
		return err
	}

	// companion file means that for each of these files their corresponding .template file
	// exists and will be rendered which will end up as the filename given
	files := []codegen.FileToGenerate{
		codegen.CompanionFile("docs/application_model/commands.md"),
		codegen.CompanionFile("docs/application_model/rest_endpoints.md"),
		codegen.CompanionFile("docs/domain_model/consts-and-enums.md"),
		codegen.CompanionFile("docs/domain_model/events.md"),
		codegen.CompanionFile("frontend/generated/apitypes.ts"),
		codegen.CompanionFile("frontend/generated/commanddefinitions.ts"),
		codegen.CompanionFile("frontend/generated/domain.ts"),
		codegen.CompanionFile("frontend/generated/restapi.ts"),
		codegen.CompanionFile("frontend/generated/version.ts"),
		codegen.Inline("pkg/apitypes/apitypes.go", codegen.ApitypesTemplate),
		codegen.Inline("pkg/apitypes/restendpoints.go", codegen.RestEndpointsTemplate),
		codegen.Inline("pkg/commandhandlers/commanddefinitions.go", codegen.CommandsDefinitionsTemplate),
		codegen.Inline("pkg/domain/consts-and-enums.go", codegen.ConstsAndEnumsTemplate),
		codegen.Inline("pkg/domain/events.go", codegen.EventDefinitionsTemplate),
	}

	if err := codegen.Run(
		"pkg/domain/domain.json",
		"pkg/apitypes/apitypes.json",
		"pkg/commandhandlers/commands.json",
		// code generation doesn't have access to version via regular method
		precompilationversion.PreCompilationVersion(),
		files); err != nil {
		return err
	}

	return nil
}
