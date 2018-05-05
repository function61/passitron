package versioncodegen

import (
	"fmt"
	"io/ioutil"
	"os"
)

func Generate() error {
	friendlyRevId := os.Getenv("FRIENDLY_REV_ID")
	if friendlyRevId == "" {
		friendlyRevId = "dev"
	}

	versionTemplate := `package version

// WARNING: generated file

const Version = "%s"

func IsDevVersion() bool {
	return Version == "dev"
}
`

	fileSerialized := []byte(fmt.Sprintf(versionTemplate, friendlyRevId))

	if err := ioutil.WriteFile("pkg/version/version.go", fileSerialized, 0644); err != nil {
		return err
	}

	return nil
}
