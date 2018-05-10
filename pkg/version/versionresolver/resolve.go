package versionresolver

import (
	"os"
)

func ResolveVersion() string {
	friendlyRevId := os.Getenv("FRIENDLY_REV_ID")
	if friendlyRevId == "" {
		friendlyRevId = "dev"
	}

	return friendlyRevId
}
