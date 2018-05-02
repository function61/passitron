package commandhandlers

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/function61/pi-security-module/pkg/command"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/keepassexport"
	"github.com/function61/pi-security-module/pkg/randompassword"
	"github.com/pquerna/otp"
	"golang.org/x/crypto/ssh"
	"regexp"
)

var (
	errAccountNotFound         = errors.New("Account not found")
	errFolderNotFound          = errors.New("Folder not found")
	errDeleteNeedsConfirmation = errors.New("Delete needs confirmation")
)

func (a *AccountRename) Invoke(ctx *command.Ctx) error {
	if ctx.State.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountRenamed(
		a.Account,
		a.Title,
		ctx.Meta))

	return nil
}

func (a *AccountChangeUsername) Invoke(ctx *command.Ctx) error {
	if ctx.State.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountUsernameChanged(
		a.Account,
		a.Username,
		ctx.Meta))

	return nil
}

func (a *AccountChangeDescription) Invoke(ctx *command.Ctx) error {
	if ctx.State.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountDescriptionChanged(
		a.Account,
		a.Description,
		ctx.Meta))

	return nil
}

func (a *AccountDeleteSecret) Invoke(ctx *command.Ctx) error {
	if ctx.State.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	// TODO: validate secret

	ctx.RaisesEvent(domain.NewAccountSecretDeleted(
		a.Account,
		a.Secret,
		ctx.Meta))

	return nil
}

func (a *AccountCreateFolder) Invoke(ctx *command.Ctx) error {
	if ctx.State.FolderById(a.Parent) == nil {
		return errFolderNotFound
	}

	ctx.RaisesEvent(domain.NewAccountFolderCreated(
		domain.RandomId(),
		a.Parent,
		a.Name,
		ctx.Meta))

	return nil
}

func (a *AccountRenameFolder) Invoke(ctx *command.Ctx) error {
	if ctx.State.FolderById(a.Id) == nil {
		return errFolderNotFound
	}

	ctx.RaisesEvent(domain.NewAccountFolderRenamed(
		a.Id,
		a.Name,
		ctx.Meta))

	return nil
}

func (a *AccountMoveFolder) Invoke(ctx *command.Ctx) error {
	if ctx.State.FolderById(a.Id) == nil {
		return errFolderNotFound
	}
	if ctx.State.FolderById(a.NewParent) == nil {
		return errFolderNotFound
	}

	ctx.RaisesEvent(domain.NewAccountFolderMoved(
		a.Id,
		a.NewParent,
		ctx.Meta))

	return nil
}

func (a *AccountCreate) Invoke(ctx *command.Ctx) error {
	accountId := domain.RandomId()

	ctx.RaisesEvent(domain.NewAccountCreated(
		accountId,
		a.FolderId,
		a.Title,
		ctx.Meta))

	if a.Username != "" {
		ctx.RaisesEvent(domain.NewAccountUsernameChanged(
			accountId,
			a.Username,
			ctx.Meta))
	}

	if a.Password != "" {
		// TODO: repeat password, but optional

		ctx.RaisesEvent(domain.NewAccountPasswordAdded(
			accountId,
			domain.RandomId(),
			a.Password,
			ctx.Meta))
	}

	return nil
}

func (a *AccountDelete) Invoke(ctx *command.Ctx) error {
	if ctx.State.AccountById(a.Id) == nil {
		return errAccountNotFound
	}

	if !a.Confirm {
		return errDeleteNeedsConfirmation
	}

	ctx.RaisesEvent(domain.NewAccountDeleted(
		a.Id,
		ctx.Meta))

	return nil
}

func (a *AccountAddPassword) Invoke(ctx *command.Ctx) error {
	if ctx.State.AccountById(a.Id) == nil {
		return errAccountNotFound
	}

	if a.Password != a.PasswordRepeat {
		return errors.New("PasswordRepeat different than Password")
	}

	password := a.Password

	if password == "_auto" {
		password = randompassword.Build(randompassword.DefaultAlphabet, 16)
	}

	ctx.RaisesEvent(domain.NewAccountPasswordAdded(
		a.Id,
		domain.RandomId(),
		password,
		ctx.Meta))

	return nil
}

func (a *AccountAddKeylist) Invoke(ctx *command.Ctx) error {
	if ctx.State.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	keys := []domain.AccountKeylistAddedKeysItem{}

	keylistParseRe := regexp.MustCompile("([a-zA-Z0-9]+)")

	matches := keylistParseRe.FindAllString(a.Keylist, -1)
	if matches == nil {
		return errors.New("unable to parse keylist")
	}

	if a.ExpectedKeyCount == 0 || a.ExpectedKeyCount*2 != len(matches) {
		return errors.New("ExpectedKeyCount does not match with parsed keylist")
	}

	for i := 0; i < len(matches); i += 2 {
		item := domain.AccountKeylistAddedKeysItem{
			Key:   matches[i],
			Value: matches[i+1],
		}

		if a.LengthOfKeys != 0 && len(item.Key) != a.LengthOfKeys {
			return errors.New("invalid length for key: " + item.Key)
		}

		if a.LengthOfValues != 0 && len(item.Value) != a.LengthOfValues {
			return errors.New("invalid length for value: " + item.Value)
		}

		keys = append(keys, item)
	}

	ctx.RaisesEvent(domain.NewAccountKeylistAdded(
		a.Account,
		domain.RandomId(),
		a.Title,
		keys,
		ctx.Meta))

	return nil
}

func (a *AccountAddSshKey) Invoke(ctx *command.Ctx) error {
	if ctx.State.AccountById(a.Id) == nil {
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

	ctx.RaisesEvent(domain.NewAccountSshKeyAdded(
		a.Id,
		domain.RandomId(),
		privateKeyReformatted,
		publicKeyAuthorizedFormat,
		ctx.Meta))

	return nil
}

func (a *AccountAddOtpToken) Invoke(ctx *command.Ctx) error {
	if ctx.State.AccountById(a.Account) == nil {
		return errAccountNotFound
	}

	// just validate key
	_, errOtpParse := otp.NewKeyFromURL(a.OtpProvisioningUrl)
	if errOtpParse != nil {
		return fmt.Errorf("Failed to parse OtpProvisioningUrl: %s", errOtpParse.Error())
	}

	ctx.RaisesEvent(domain.NewAccountOtpTokenAdded(
		a.Account,
		domain.RandomId(),
		a.OtpProvisioningUrl,
		ctx.Meta))

	return nil
}

func (a *DatabaseChangeMasterPassword) Invoke(ctx *command.Ctx) error {
	if a.NewMasterPassword != a.NewMasterPasswordRepeat {
		return errors.New("NewMasterPassword not same as NewMasterPasswordRepeat")
	}

	ctx.RaisesEvent(domain.NewDatabaseMasterPasswordChanged(
		a.NewMasterPassword,
		ctx.Meta))

	return nil
}

func (a *DatabaseExportToKeepass) Invoke(ctx *command.Ctx) error {
	return keepassexport.Export(ctx.State)
}

func (a *DatabaseUnseal) Invoke(ctx *command.Ctx) error {
	// TODO: predictable comparison time
	if ctx.State.GetMasterPassword() != a.MasterPassword {
		return errors.New("invalid password")
	}

	if ctx.State.IsUnsealed() {
		return errors.New("state already unsealed")
	}
	ctx.State.SetSealed(false)

	ctx.RaisesEvent(domain.NewDatabaseUnsealed(ctx.Meta))

	return nil
}
