package main

type SecretDeleted struct {
	Id string
}

func (e *SecretDeleted) Apply() {
	for idx, s := range state.Secrets {
		if s.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			state.Secrets = append(state.Secrets[:idx], state.Secrets[idx+1:]...)
			return
		}
	}
}
