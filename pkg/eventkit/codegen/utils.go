package codegen

import (
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
		return fmt.Errorf("WriteTemplateFile Parse %s: %v", filename, err)
	}

	if err := tpl.Execute(file, data); err != nil {
		return fmt.Errorf("WriteTemplateFile %s: %v", filename, err)
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

func allOk(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}
