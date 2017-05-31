package main

type SecretCreated struct {
	Id       string
	FolderId string
	Title    string
	Username string
	Password string
}

func (e *SecretCreated) Apply() {
	secret := InsecureSecret{
		Id:       e.Id,
		FolderId: e.FolderId,
		Title:    e.Title,
		Username: e.Username,
		Password: e.Password,
	}

	state.Secrets = append(state.Secrets, secret)
}
