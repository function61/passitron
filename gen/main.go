package main

import (
	"io/ioutil"
	"regexp"
	"strings"
)

func main() {
	genCommandHandlerMap()
	genEventHandlerMap()
}

/*
func ApplyOneEvent(event []interface{}) bool {
	switch e := event.(type) {
	default:
		return false
	case SecretCreated:
		e.Apply()
	case FolderCreated:
		e.Apply()
	case SecretRenamed:
		e.Apply()
	}

	return true
}
*/
func genEventHandlerMap() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		panic(err)
	}

	fileLines := []string{
		"package main",
		"",
		"// WARNING: GENERATED FILE",
		"",
		"func ApplyOneEvent(event interface{}) bool {",
		"	switch e := event.(type) {",
		"		default:",
		"			return false",
	}

	eventRe := regexp.MustCompile("^(.+)Event\\.go$")

	for _, file := range files {
		match := eventRe.FindStringSubmatch(file.Name())
		if len(match) == 0 {
			continue
		}

		// "SecretCreated"
		eventName := match[1]

		fileLines = append(fileLines, "		case "+eventName+":")
		fileLines = append(fileLines, "			e.Apply()")
	}

	fileLines = append(fileLines,
		"	}",
		"",
		"	return true",
		"}")

	fileLinesSerialized := strings.Join(fileLines, "\n")

	if err := ioutil.WriteFile("eventhandlersmap.go", []byte(fileLinesSerialized), 0644); err != nil {
		panic(err)
	}
}

func genCommandHandlerMap() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		panic(err)
	}

	fileLines := []string{
		"package main",
		"",
		"// WARNING: GENERATED FILE",
		"",
		"import \"net/http\"",
		"",
		"var commandHandlers = map[string]func(w http.ResponseWriter, r *http.Request){",
	}

	requestRe := regexp.MustCompile("^(.+Request)\\.go$")

	for _, file := range files {
		match := requestRe.FindStringSubmatch(file.Name())
		if len(match) == 0 {
			continue
		}

		// "FolderCreateRequest"
		reqName := match[1]
		reqHandler := "Handle" + reqName

		line := "	\"" + reqName + "\": " + reqHandler + ","

		fileLines = append(fileLines, line)
	}

	fileLines = append(fileLines, "}", "")

	fileLinesSerialized := strings.Join(fileLines, "\n")

	if err := ioutil.WriteFile("commandhandlersgen.go", []byte(fileLinesSerialized), 0644); err != nil {
		panic(err)
	}
}
