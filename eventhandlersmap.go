package main

// WARNING: GENERATED FILE

func ApplyOneEvent(event interface{}) bool {
	switch e := event.(type) {
	default:
		return false
	case DescriptionChanged:
		e.Apply()
	case FolderCreated:
		e.Apply()
	case OtpTokenSet:
		e.Apply()
	case PasswordChanged:
		e.Apply()
	case SecretCreated:
		e.Apply()
	case SecretDeleted:
		e.Apply()
	case SecretRenamed:
		e.Apply()
	}

	return true
}
