package main

type SecretCreated struct {
	Id       string
	FolderId string
	Title    string
}

func (e *SecretCreated) Apply() {
	secret := InsecureSecret{
		Id:       e.Id,
		FolderId: e.FolderId,
		Title:    e.Title,
	}

	state.Secrets = append(state.Secrets, secret)
}
