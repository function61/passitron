package main

import (
	"github.com/function61/gokit/dynversion/precompilationversion"
	"github.com/function61/pi-security-module/pkg/eventkit/codegen"
	"github.com/function61/pi-security-module/pkg/eventkit/codegen/codegentemplates"
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
	domainFiles := []codegen.FileToGenerate{
		codegen.Inline("docs/domain_model/types.md", codegentemplates.DocsTypes),
		codegen.Inline("docs/domain_model/events.md", codegentemplates.DocsEvents),
	}

	commandsFiles := []codegen.FileToGenerate{
		codegen.Inline("docs/application_model/commands.md", codegentemplates.DocsCommands),
	}

	apitypesFiles := []codegen.FileToGenerate{
		codegen.Inline("docs/application_model/rest_endpoints.md", codegentemplates.DocsRestEndpoints),
		codegen.Inline("docs/application_model/types.md", codegentemplates.DocsTypes),
	}

	modules := []*codegen.Module{
		codegen.NewModule("domain", "pkg/domain/types.json", "pkg/domain/events.json", "", domainFiles),
		codegen.NewModule("commands", "", "", "pkg/commands/commands.json", commandsFiles),
		codegen.NewModule("apitypes", "pkg/apitypes/types.json", "", "", apitypesFiles),
	}

	opts := codegen.Opts{
		BackendModulePrefix:  "github.com/function61/pi-security-module/pkg/",
		FrontendModulePrefix: "generated/",
	}

	if err := codegen.ProcessModules(modules, opts); err != nil {
		return err
	}

	// PreCompilationVersion = code generation doesn't have access to version via regular method
	if err := codegen.ProcessFile(
		codegen.Inline("frontend/generated/version.ts", codegentemplates.FrontendVersion),
		codegen.NewVersionData(precompilationversion.PreCompilationVersion()),
	); err != nil {
		return err
	}

	return nil
}
