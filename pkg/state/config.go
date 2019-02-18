package state

import (
	"errors"
	"github.com/function61/eventkit/event"
	"github.com/function61/gokit/fileexists"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/storedpassword"
	"github.com/function61/pi-security-module/pkg/domain"
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

	meta := event.Meta(time.Now(), state.NextFreeUserId())

	userCreated := domain.NewUserCreated(
		meta.UserId,
		adminUsername,
		meta)

	password := domain.NewUserPasswordUpdated(
		meta.UserId,
		string(storedPassword),
		false,
		meta)

	return state.EventLog.Append([]event.Event{userCreated, password})
}
