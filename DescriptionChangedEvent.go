package main

type DescriptionChanged struct {
	Id          string
	Description string
}

func (e *DescriptionChanged) Apply() {
	for idx, s := range state.Secrets {
		if s.Id == e.Id {
			s.Description = e.Description
			state.Secrets[idx] = s
			return
		}
	}
}
