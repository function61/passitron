package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

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
