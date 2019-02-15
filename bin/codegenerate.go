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

	modules := []*codegen.Module{
		codegen.NewModule("domain", "pkg/domain/types.json", "pkg/domain/events.json", ""),
		codegen.NewModule("commands", "", "", "pkg/commands/commands.json"),
		codegen.NewModule("apitypes", "pkg/apitypes/types.json", "", ""),
		codegen.NewModule("signingapi", "pkg/signingapi/types.json", "", ""),
	}

	opts := codegen.Opts{
		BackendModulePrefix:    "github.com/function61/pi-security-module/pkg/",
		FrontendModulePrefix:   "generated/",
		AutogenerateModuleDocs: true,
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
