package event

// TODO: this is for once we get timestamps
type MasterPasswordChanged struct {
}

func (e *MasterPasswordChanged) Apply() {
	// noop
}
