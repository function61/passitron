package state

import (
	"errors"
	"github.com/function61/gokit/fileexists"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/storedpassword"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"time"
)

const (
	configFilename = "config.json"
)

type Config struct {
	JwtPrivateKey string `json:"jwt_private_key"`
	JwtPublicKey  string `json:"jwt_public_key"`
}

func readConfig() (*Config, error) {
	cfg := &Config{}
	return cfg, jsonfile.Read(configFilename, cfg, true)
}

func saveConfig(cfg *Config) error {
	return jsonfile.Write(configFilename, cfg)
}

func InitConfig(adminUsername string, adminPassword string) error {
	exists, err := fileexists.Exists(configFilename)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("config file already exists")
	}

	privKeyPem, pubKeyPem, err := httpauth.GenerateKey()
	if err != nil {
		return err
	}

	cfg := &Config{
		JwtPrivateKey: string(privKeyPem),
		JwtPublicKey:  string(pubKeyPem),
	}

	if err := saveConfig(cfg); err != nil {
		return err
	}

	state := New(logex.Discard)
	defer state.Close()

	storedPassword, err := storedpassword.Store(
		adminPassword,
		storedpassword.CurrentBestDerivationStrategy)
	if err != nil {
		return err
	}

	uid := "2"
	now := time.Now()

	userCreated := domain.NewUserCreated(
		uid,
		adminUsername,
		event.Meta(now, uid))

	password := domain.NewUserPasswordUpdated(
		uid,
		string(storedPassword),
		false,
		event.Meta(now, uid))

	return state.EventLog.Append([]event.Event{userCreated, password})
}
