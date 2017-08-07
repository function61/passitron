package util

import (
	"github.com/function61/pi-security-module/accountevent"
	folderevent "github.com/function61/pi-security-module/folder/event"
	sessionevent "github.com/function61/pi-security-module/session/event"
)

// WARNING: GENERATED FILE

func ApplyOneEvent(event interface{}) bool {
	// FIXME: use interface for this?

	switch e := event.(type) {
	default:
		return false
	case accountevent.DescriptionChanged:
		e.Apply()
	case folderevent.FolderCreated:
		e.Apply()
	case folderevent.FolderMoved:
		e.Apply()
	case folderevent.FolderRenamed:
		e.Apply()
	case accountevent.OtpTokenAdded:
		e.Apply()
	case accountevent.PasswordAdded:
		e.Apply()
	case accountevent.SshKeyAdded:
		e.Apply()
	case accountevent.AccountCreated:
		e.Apply()
	case accountevent.SecretDeleted:
		e.Apply()
	case accountevent.AccountDeleted:
		e.Apply()
	case accountevent.AccountRenamed:
		e.Apply()
	case accountevent.UsernameChanged:
		e.Apply()
	case accountevent.SecretUsed:
		e.Apply()
	case sessionevent.MasterPasswordChanged:
		e.Apply()
	}

	return true
}
