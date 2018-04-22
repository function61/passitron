package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/folder/event"
	sessionevent "github.com/function61/pi-security-module/session/event"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
	"github.com/function61/pi-security-module/util/keepassexport"
	"github.com/function61/pi-security-module/util/randompassword"
	"github.com/pquerna/otp"
	"golang.org/x/crypto/ssh"
)

var (
	errAccountNotFound         = errors.New("Account not found")
	errFolderNotFound          = errors.New("Folder not found")
	errDeleteNeedsConfirmation = errors.New("Delete needs confirmation")
)

func (a *AccountRename) Invoke(ctx *Ctx) error {
	if state.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.State.EventLog.Append(accountevent.AccountRenamed{
		Event: eventbase.NewEvent(),
		Id:    a.Account,
		Title: a.Title,
	})

	return nil
}

func (a *AccountChangeUsername) Invoke(ctx *Ctx) error {
	if state.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.State.EventLog.Append(accountevent.UsernameChanged{
		Event:    eventbase.NewEvent(),
		Id:       a.Account,
		Username: a.Username,
	})

	return nil
}

func (a *AccountChangeDescription) Invoke(ctx *Ctx) error {
	if state.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.State.EventLog.Append(accountevent.DescriptionChanged{
		Event:       eventbase.NewEvent(),
		Id:          a.Account,
		Description: a.Description,
	})

	return nil
}

func (a *AccountDeleteSecret) Invoke(ctx *Ctx) error {
	if state.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	// TODO: validate secret

	ctx.State.EventLog.Append(accountevent.SecretDeleted{
		Event:   eventbase.NewEvent(),
		Account: a.Account,
		Secret:  a.Secret,
	})

	return nil
}

func (a *AccountCreateFolder) Invoke(ctx *Ctx) error {
	if state.FolderById(a.Parent) == nil {
		return errFolderNotFound
	}

	ctx.State.EventLog.Append(event.FolderCreated{
		Event:    eventbase.NewEvent(),
		Id:       eventbase.RandomId(),
		ParentId: a.Parent,
		Name:     a.Name,
	})

	return nil
}

func (a *AccountRenameFolder) Invoke(ctx *Ctx) error {
	if state.FolderById(a.Id) == nil {
		return errFolderNotFound
	}

	ctx.State.EventLog.Append(event.FolderRenamed{
		Event: eventbase.NewEvent(),
		Id:    a.Id,
		Name:  a.Name,
	})

	return nil
}

func (a *AccountMoveFolder) Invoke(ctx *Ctx) error {
	if state.FolderById(a.Id) == nil {
		return errFolderNotFound
	}
	if state.FolderById(a.NewParent) == nil {
		return errFolderNotFound
	}

	ctx.State.EventLog.Append(event.FolderMoved{
		Event:    eventbase.NewEvent(),
		Id:       a.Id,
		ParentId: a.NewParent,
	})

	return nil
}

func (a *AccountCreate) Invoke(ctx *Ctx) error {
	accountId := eventbase.RandomId()

	events := []eventbase.EventInterface{
		accountevent.AccountCreated{
			Event:    eventbase.NewEvent(),
			Id:       accountId,
			FolderId: a.FolderId,
			Title:    a.Title,
		},
	}

	if a.Username != "" {
		events = append(events, accountevent.UsernameChanged{
			Event:    eventbase.NewEvent(),
			Id:       accountId,
			Username: a.Username,
		})
	}

	if a.Password != "" {
		// TODO: repeat password, but optional

		events = append(events, accountevent.PasswordAdded{
			Event:    eventbase.NewEvent(),
			Account:  accountId,
			Id:       eventbase.RandomId(),
			Password: a.Password,
		})
	}

	ctx.State.EventLog.AppendBatch(events)

	return nil
}

func (a *AccountDelete) Invoke(ctx *Ctx) error {
	if state.AccountById(a.Id) == nil {
		return errAccountNotFound
	}

	if !a.Confirm {
		return errDeleteNeedsConfirmation
	}

	ctx.State.EventLog.Append(accountevent.AccountDeleted{
		Event: eventbase.NewEvent(),
		Id:    a.Id,
	})

	return nil
}

func (a *AccountAddPassword) Invoke(ctx *Ctx) error {
	if state.AccountById(a.Id) == nil {
		return errAccountNotFound
	}

	if a.Password != a.PasswordRepeat {
		return errors.New("PasswordRepeat different than Password")
	}

	if a.Password == "_auto" {
		a.Password = randompassword.Build(randompassword.DefaultAlphabet, 16)
	}

	ctx.State.EventLog.Append(accountevent.PasswordAdded{
		Event:    eventbase.NewEvent(),
		Account:  a.Id,
		Id:       eventbase.RandomId(),
		Password: a.Password,
	})

	return nil
}

func (a *AccountAddSshKey) Invoke(ctx *Ctx) error {
	if state.AccountById(a.Id) == nil {
		return errAccountNotFound
	}

	// validate and re-format SSH key
	block, rest := pem.Decode([]byte(a.SshPrivateKey))
	if block == nil {
		return errors.New("Failed to parse PEM block")
	}

	if len(rest) > 0 {
		return errors.New("Extra data included in PEM content")
	}

	if block.Type != "RSA PRIVATE KEY" {
		return errors.New("Currently we only support RSA format keys")
	}

	if x509.IsEncryptedPEMBlock(block) {
		// TODO: maybe implement here in import phase
		return errors.New("We do not support encypted PEM blocks yet")
	}

	privateKeyReformatted := string(pem.EncodeToMemory(block))

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	// convert to SSH public key
	publicKeySsh, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	publicKeyAuthorizedFormat := string(ssh.MarshalAuthorizedKey(publicKeySsh))

	ctx.State.EventLog.Append(accountevent.SshKeyAdded{
		Event:                  eventbase.NewEvent(),
		Account:                a.Id,
		Id:                     eventbase.RandomId(),
		SshPrivateKey:          privateKeyReformatted,
		SshPublicKeyAuthorized: publicKeyAuthorizedFormat,
	})

	return nil
}

func (a *AccountAddOtpToken) Invoke(ctx *Ctx) error {
	if state.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	// just validate key
	_, errOtpParse := otp.NewKeyFromURL(a.OtpProvisioningUrl)
	if errOtpParse != nil {
		return fmt.Errorf("Failed to parse OtpProvisioningUrl: %s", errOtpParse.Error())
	}

	ctx.State.EventLog.Append(accountevent.OtpTokenAdded{
		Event:              eventbase.NewEvent(),
		Account:            a.Account,
		Id:                 eventbase.RandomId(),
		OtpProvisioningUrl: a.OtpProvisioningUrl,
	})

	return nil
}

func (a *DatabaseChangeMasterPassword) Invoke(ctx *Ctx) error {
	if a.NewMasterPassword != a.NewMasterPasswordRepeat {
		return errors.New("NewMasterPassword not same as NewMasterPasswordRepeat")
	}

	ctx.State.EventLog.Append(sessionevent.MasterPasswordChanged{
		Event:    eventbase.NewEvent(),
		Password: a.NewMasterPassword,
	})

	return nil
}

func (a *DatabaseExportToKeepass) Invoke(ctx *Ctx) error {
	return keepassexport.Export()
}

func (a *DatabaseUnseal) Invoke(ctx *Ctx) error {
	// TODO: predictable comparison time
	if ctx.State.GetMasterPassword() != a.MasterPassword {
		return errors.New("invalid password")
	}

	if ctx.State.IsUnsealed() {
		return errors.New("state already unsealed")
	}
	ctx.State.SetSealed(false)

	ctx.State.EventLog.Append(sessionevent.DatabaseUnsealed{
		Event: eventbase.NewEvent(),
	})

	return nil
}
