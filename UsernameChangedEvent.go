package main

type UsernameChanged struct {
	Id       string
	Username string
}

func (e *UsernameChanged) Apply() {
	for idx, s := range state.Secrets {
		if s.Id == e.Id {
			s.Username = e.Username
			state.Secrets[idx] = s
			return
		}
	}
}
