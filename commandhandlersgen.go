package main

// WARNING: GENERATED FILE

import (
	"github.com/function61/pi-security-module/accountcommand"
	foldercommand "github.com/function61/pi-security-module/folder/command"
	sessioncommand "github.com/function61/pi-security-module/session/command"
	"net/http"
)

var commandHandlers = map[string]func(w http.ResponseWriter, r *http.Request){
	"ChangeDescriptionRequest":    accountcommand.HandleChangeDescriptionRequest,
	"ChangePasswordRequest":       accountcommand.HandleChangePasswordRequest,
	"ChangeUsernameRequest":       accountcommand.HandleChangeUsernameRequest,
	"DeleteAccountRequest":        accountcommand.HandleDeleteAccountRequest,
	"FolderCreateRequest":         foldercommand.HandleFolderCreateRequest,
	"MoveFolderRequest":           foldercommand.HandleMoveFolderRequest,
	"RenameFolderRequest":         foldercommand.HandleRenameFolderRequest,
	"RenameSecretRequest":         accountcommand.HandleRenameSecretRequest,
	"SecretCreateRequest":         accountcommand.HandleSecretCreateRequest,
	"SetSshKeyRequest":            accountcommand.HandleSetSshKeyRequest,
	"SetOtpTokenRequest":          accountcommand.HandleSetOtpTokenRequest,
	"DeleteSecretRequest":         accountcommand.HandleDeleteSecretRequest,
	"WriteKeepassRequest":         HandleWriteKeepassRequest,
	"UnsealRequest":               sessioncommand.HandleUnsealRequest,
	"ChangeMasterPasswordRequest": sessioncommand.HandleChangeMasterPassword,
}
