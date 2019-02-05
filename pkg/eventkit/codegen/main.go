package codegen

import (
	"github.com/function61/gokit/jsonfile"
	"io/ioutil"
)

type FileToGenerate struct {
	targetPath     string
	obtainTemplate func() (string, error)
}

func Run(
	domainJsonPath string,
	apitypesJsonPath string,
	commandsJsonPath string,
	version string,
	files []FileToGenerate,
) error {
	domainSpecs := &DomainFile{}
	if err := jsonfile.Read(domainJsonPath, domainSpecs, true); err != nil {
		return err
	}

	applicationTypes := &ApplicationTypesDefinition{}
	if err := jsonfile.Read(apitypesJsonPath, applicationTypes, true); err != nil {
		return err
	}

	commandSpecs := &CommandSpecFile{}
	if err := jsonfile.Read(commandsJsonPath, commandSpecs, true); err != nil {
		return err
	}
	if err := commandSpecs.Validate(); err != nil {
		return err
	}

	eventDefs, eventStructsAsGoCode := ProcessEvents(domainSpecs)

	data := &TplData{
		Version:              version,
		DomainSpecs:          domainSpecs,
		CommandSpecs:         commandSpecs,
		ApplicationTypes:     applicationTypes,
		StringEnums:          ProcessStringEnums(domainSpecs),
		EventDefs:            eventDefs,
		EventStructsAsGoCode: eventStructsAsGoCode,
	}

	for _, file := range files {
		if err := renderOneTemplate(file, data); err != nil {
			return err
		}
	}

	return nil
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

func renderOneTemplate(target FileToGenerate, data *TplData) error {
	templateContent, err := target.obtainTemplate()
	if err != nil {
		return err
	}

	return WriteTemplateFile(target.targetPath, data, templateContent)
}
