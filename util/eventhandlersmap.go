package util

import (
	folderevent "github.com/function61/pi-security-module/folder/event"
	secretevent "github.com/function61/pi-security-module/secret/event"
)

// WARNING: GENERATED FILE

func ApplyOneEvent(event interface{}) bool {
	// FIXME: use interface for this?

	switch e := event.(type) {
	default:
		return false
	case secretevent.DescriptionChanged:
		e.Apply()
	case folderevent.FolderCreated:
		e.Apply()
	case folderevent.FolderMoved:
		e.Apply()
	case folderevent.FolderRenamed:
		e.Apply()
	case secretevent.OtpTokenSet:
		e.Apply()
	case secretevent.PasswordChanged:
		e.Apply()
	case secretevent.SecretCreated:
		e.Apply()
	case secretevent.SecretDeleted:
		e.Apply()
	case secretevent.SecretRenamed:
		e.Apply()
	case secretevent.UsernameChanged:
		e.Apply()
	}

	return true
}
