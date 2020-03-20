package main

import (
	"fmt"
	"github.com/function61/eventkit/codegen"
	"github.com/function61/eventkit/codegen/codegentemplates"
	"github.com/function61/gokit/dynversion/precompilationversion"
	"os"
)

//go:generate go run codegenerate.go

func main() {
	if err := logic(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func logic() error {
	// normalize to root of the project
	if err := os.Chdir(".."); err != nil {
		return err
	}

	return mainInternal()
}

func mainInternal() error {
	modules := []*codegen.Module{
		codegen.NewModule("domain", "pkg/domain/types.json", "pkg/domain/events.json", ""),
		codegen.NewModule("apitypes", "pkg/apitypes/types.json", "", "pkg/apitypes/commands.json"),
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
