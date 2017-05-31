package main

type PasswordChanged struct {
	Id       string
	Password string
}

func (e *PasswordChanged) Apply() {
	for idx, s := range state.Secrets {
		if s.Id == e.Id {
			s.Password = e.Password
			state.Secrets[idx] = s
			return
		}
	}
}
