package useraccounts

type Account struct {
	Id       string
	Password string // TODO: PBKDF2 love
}

type Repository interface {
	// returns nil, nil if username not found
	// returns nil, err on e.g. error accessing DB
	FindByUsername(username string) (*Account, error)
}

// just a dummy for now, until we implement the real one
var DummyRepository Repository = &dummyRepository{}

type dummyRepository struct{}

func (d *dummyRepository) FindByUsername(username string) (*Account, error) {
	if username != "joonas" {
		return nil, nil
	}

	return &Account{
		Id:       "2",
		Password: "poop",
	}, nil
}
