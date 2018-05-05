package codegen

import (
	"encoding/json"
	"os"
	"text/template"
)

func WriteTemplateFile(filename string, data interface{}, templateString string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	tpl, _ := template.New("").Parse(templateString)

	return tpl.Execute(file, data)
}

func DeserializeDomainFile(path string) (*DomainFile, error) {
	domainFile, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}

	file := &DomainFile{}
	if jsonErr := json.NewDecoder(domainFile).Decode(file); jsonErr != nil {
		return nil, jsonErr
	}

	return file, nil
}
