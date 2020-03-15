package state

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"github.com/function61/eventhorizon/pkg/ehevent"
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

type JwtConfig struct {
	SigningKey       string `json:"jwt_private_key"`
	AuthenticatorKey string `json:"jwt_public_key"`
}

func readAndValidateJwtConfig() (*JwtConfig, error) {
	cfg := &JwtConfig{}

	if err := jsonfile.Read(configFilename, cfg, true); err != nil {
		return nil, err
	}

	if cfg.AuthenticatorKey == "" {
		return nil, errors.New("AuthenticatorKey not set")
	}

	if cfg.SigningKey == "" {
		return nil, errors.New("SigningKey not set")
	}

	return cfg, nil
}

func writeJwtConfig(cfg *JwtConfig) error {
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

	cfg := &JwtConfig{
		SigningKey:       string(privKeyPem),
		AuthenticatorKey: string(pubKeyPem),
	}

	if err := writeJwtConfig(cfg); err != nil {
		return err
	}

	state, err := New(logex.Discard)
	if err != nil {
		return err
	}

	return createAdminUser(adminUsername, adminPassword, state)
}

func createAdminUser(adminUsername string, adminPassword string, state *AppState) error {
	storedPassword, err := storedpassword.Store(
		adminPassword,
		storedpassword.CurrentBestDerivationStrategy)
	if err != nil {
		return err
	}

	meta := ehevent.Meta(time.Now(), "2")

	userCreated := domain.NewUserCreated(
		meta.UserId,
		adminUsername,
		meta)

	password := domain.NewUserPasswordUpdated(
		meta.UserId,
		string(storedPassword),
		false,
		meta)

	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	decryptionKeyChanged, err := ExportPrivateKeyWithPassword(privKey, adminPassword, meta)
	if err != nil {
		return err
	}

	return state.EventLog.Append([]ehevent.Event{userCreated, password, decryptionKeyChanged})
}
