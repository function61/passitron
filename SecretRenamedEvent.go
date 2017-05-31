package main

type SecretRenamed struct {
	Id    string
	Title string
}

func (e *SecretRenamed) Apply() {
	for idx, s := range state.Secrets {
		if s.Id == e.Id {
			s.Title = e.Title
			state.Secrets[idx] = s
			return
		}
	}
}
