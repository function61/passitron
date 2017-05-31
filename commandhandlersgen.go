package main

// WARNING: GENERATED FILE

import "net/http"

var commandHandlers = map[string]func(w http.ResponseWriter, r *http.Request){
	"ChangeDescriptionRequest": HandleChangeDescriptionRequest,
	"ChangePasswordRequest":    HandleChangePasswordRequest,
	"ChangeUsernameRequest":    HandleChangeUsernameRequest,
	"DeleteSecretRequest":      HandleDeleteSecretRequest,
	"FolderCreateRequest":      HandleFolderCreateRequest,
	"MoveFolderRequest":        HandleMoveFolderRequest,
	"RenameFolderRequest":      HandleRenameFolderRequest,
	"RenameSecretRequest":      HandleRenameSecretRequest,
	"SecretCreateRequest":      HandleSecretCreateRequest,
	"SetOtpTokenRequest":       HandleSetOtpTokenRequest,
	"WriteKeepassRequest":      HandleWriteKeepassRequest,
}
