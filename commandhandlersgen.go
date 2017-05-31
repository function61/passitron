package main

// WARNING: GENERATED FILE

import "net/http"

var commandHandlers = map[string]func(w http.ResponseWriter, r *http.Request){
	"ChangeDescriptionRequest": HandleChangeDescriptionRequest,
	"ChangePasswordRequest":    HandleChangePasswordRequest,
	"DeleteSecretRequest":      HandleDeleteSecretRequest,
	"FolderCreateRequest":      HandleFolderCreateRequest,
	"RenameSecretRequest":      HandleRenameSecretRequest,
	"SecretCreateRequest":      HandleSecretCreateRequest,
	"WriteKeepassRequest":      HandleWriteKeepassRequest,
}
