package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func main() {
	if err := genVersionFile(); err != nil {
		panic(err)
	}
	// genCommandHandlerMap()
	// genEventHandlerMap()
}

func genVersionFile() error {
	friendlyRevId := os.Getenv("FRIENDLY_REV_ID")
	if friendlyRevId == "" {
		friendlyRevId = "dev"
	}

	versionTemplate := `package main

// WARNING: generated file

const version = "%s"

func isDevVersion() bool {
	return version == "dev"
}
`

	fileSerialized := []byte(fmt.Sprintf(versionTemplate, friendlyRevId))

	if err := ioutil.WriteFile("version.go", fileSerialized, 0644); err != nil {
		return err
	}

	return nil
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
