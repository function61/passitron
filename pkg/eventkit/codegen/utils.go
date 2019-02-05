package codegen

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/template"
)

func WriteTemplateFile(filename string, data interface{}, templateString string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	templateFuncs := template.FuncMap{
		"StripQueryFromUrl": stripQueryFromUrl,
		"UppercaseFirst":    func(input string) string { return strings.ToUpper(input[0:1]) + input[1:] },
	}

	tpl, err := template.New("").Funcs(templateFuncs).Parse(templateString)
	if err != nil {
		return err
	}

	return tpl.Execute(file, data)
}

func DeserializeJsonFile(path string, data interface{}) error {
	file, openErr := os.Open(path)
	if openErr != nil {
		return openErr
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("DeserializeJsonFile: %s: %s", path, err.Error())
	}

	return nil
}

// "/search?q={stuff}" => "/search"
func stripQueryFromUrl(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		panic(err)
	}

	return u.Path
}
