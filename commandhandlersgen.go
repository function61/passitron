package main

// WARNING: GENERATED FILE

import (
	foldercommand "github.com/function61/pi-security-module/folder/command"
	secretcommand "github.com/function61/pi-security-module/secret/command"
	sessioncommand "github.com/function61/pi-security-module/session/command"
	"net/http"
)

var commandHandlers = map[string]func(w http.ResponseWriter, r *http.Request){
	"ChangeDescriptionRequest":    secretcommand.HandleChangeDescriptionRequest,
	"ChangePasswordRequest":       secretcommand.HandleChangePasswordRequest,
	"ChangeUsernameRequest":       secretcommand.HandleChangeUsernameRequest,
	"DeleteSecretRequest":         secretcommand.HandleDeleteSecretRequest,
	"FolderCreateRequest":         foldercommand.HandleFolderCreateRequest,
	"MoveFolderRequest":           foldercommand.HandleMoveFolderRequest,
	"RenameFolderRequest":         foldercommand.HandleRenameFolderRequest,
	"RenameSecretRequest":         secretcommand.HandleRenameSecretRequest,
	"SecretCreateRequest":         secretcommand.HandleSecretCreateRequest,
	"SetSshKeyRequest":            secretcommand.HandleSetSshKeyRequest,
	"SetOtpTokenRequest":          secretcommand.HandleSetOtpTokenRequest,
	"WriteKeepassRequest":         HandleWriteKeepassRequest,
	"UnsealRequest":               sessioncommand.HandleUnsealRequest,
	"ChangeMasterPasswordRequest": sessioncommand.HandleChangeMasterPassword,
}
