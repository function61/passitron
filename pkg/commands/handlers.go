package commands

import (
	"crypto/subtle"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/function61/gokit/cryptorandombytes"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/randompassword"
	"github.com/function61/gokit/storedpassword"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventkit/command"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"github.com/function61/pi-security-module/pkg/keepassexport"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/function61/pi-security-module/pkg/u2futil"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/tstranex/u2f"
	"golang.org/x/crypto/ssh"
	"log"
	"net/url"
	"regexp"
	"time"
)

var (
	errAccountNotFound = errors.New("Account not found")
	errFolderNotFound  = errors.New("Folder not found")
)

type CommandHandlers struct {
	state *state.AppState
	logl  *logex.Leveled
}

func (c *CommandHandlers) userData(ctx *command.Ctx) *state.UserStorage {
	return c.state.DB.UserScope[ctx.Meta.UserId]
}

func New(state *state.AppState, logger *log.Logger) *CommandHandlers {
	return &CommandHandlers{state, logex.Levels(logger)}
}

func (h *CommandHandlers) AccountRename(a *AccountRename, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountRenamed(
		a.Account,
		a.Title,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountMove(a *AccountMove, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountMoved(
		a.Account,
		a.NewParentFolder,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountChangeUsername(a *AccountChangeUsername, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountUsernameChanged(
		a.Account,
		a.Username,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountChangeUrl(a *AccountChangeUrl, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	if a.Url != "" {
		if _, err := url.Parse(a.Url); err != nil {
			return err
		}
	}

	ctx.RaisesEvent(domain.NewAccountUrlChanged(
		a.Account,
		a.Url,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountChangeDescription(a *AccountChangeDescription, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountDescriptionChanged(
		a.Account,
		a.Description,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountDeleteSecret(a *AccountDeleteSecret, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	// TODO: validate secret

	ctx.RaisesEvent(domain.NewAccountSecretDeleted(
		a.Account,
		a.Secret,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountCreateFolder(a *AccountCreateFolder, ctx *command.Ctx) error {
	if h.userData(ctx).FolderById(a.Parent) == nil {
		return errFolderNotFound
	}

	ctx.RaisesEvent(domain.NewAccountFolderCreated(
		event.RandomId(),
		a.Parent,
		a.Name,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountDeleteFolder(a *AccountDeleteFolder, ctx *command.Ctx) error {
	if h.userData(ctx).FolderById(a.Id) == nil {
		return errFolderNotFound
	}

	subFolders := h.userData(ctx).SubfoldersByParentId(a.Id)
	accounts := h.userData(ctx).WrappedAccountsByFolder(a.Id)

	if len(subFolders) > 0 || len(accounts) > 0 {
		return errors.New("folder not empty")
	}

	ctx.RaisesEvent(domain.NewAccountFolderDeleted(
		a.Id,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountRenameFolder(a *AccountRenameFolder, ctx *command.Ctx) error {
	if h.userData(ctx).FolderById(a.Id) == nil {
		return errFolderNotFound
	}

	ctx.RaisesEvent(domain.NewAccountFolderRenamed(
		a.Id,
		a.Name,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountMoveFolder(a *AccountMoveFolder, ctx *command.Ctx) error {
	if h.userData(ctx).FolderById(a.Id) == nil {
		return errFolderNotFound
	}
	if h.userData(ctx).FolderById(a.NewParent) == nil {
		return errFolderNotFound
	}

	ctx.RaisesEvent(domain.NewAccountFolderMoved(
		a.Id,
		a.NewParent,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountCreate(a *AccountCreate, ctx *command.Ctx) error {
	accountId := event.RandomId()

	title := a.Title

	if title == "" && a.Url != "" {
		urlParsed, err := url.Parse(a.Url)
		if err != nil {
			return err
		}

		title = urlParsed.Hostname()
	}

	if title == "" {
		return errors.New("you must specify at least Title or the Url")
	}

	ctx.RaisesEvent(domain.NewAccountCreated(
		accountId,
		a.FolderId,
		title,
		ctx.Meta))

	if a.Username != "" {
		ctx.RaisesEvent(domain.NewAccountUsernameChanged(
			accountId,
			a.Username,
			ctx.Meta))
	}

	if a.Password != "" {
		// if PasswordRepeat given, verify it
		if err := verifyRepeatPassword(a.Password, a.PasswordRepeat); a.PasswordRepeat != "" && err != nil {
			return err
		}

		ctx.RaisesEvent(domain.NewAccountPasswordAdded(
			accountId,
			event.RandomId(),
			a.Password,
			ctx.Meta))
	}

	if a.Url != "" {
		if _, err := url.Parse(a.Url); err != nil {
			return err
		}

		ctx.RaisesEvent(domain.NewAccountUrlChanged(
			accountId,
			a.Url,
			ctx.Meta))
	}

	return nil
}

func (h *CommandHandlers) AccountDelete(a *AccountDelete, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Id) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountDeleted(
		a.Id,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountAddPassword(a *AccountAddPassword, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	if err := verifyRepeatPassword(a.Password, a.PasswordRepeat); err != nil {
		return err
	}

	password := a.Password

	if password == "_auto" {
		password = randompassword.Build(randompassword.DefaultAlphabet, 16)
	}

	ctx.RaisesEvent(domain.NewAccountPasswordAdded(
		a.Account,
		event.RandomId(),
		password,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountAddSecretNote(a *AccountAddSecretNote, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountSecretNoteAdded(
		a.Account,
		event.RandomId(),
		a.Title,
		a.Note,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountAddExternalU2FToken(a *AccountAddExternalU2FToken, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountExternalTokenAdded(
		a.Account,
		event.RandomId(),
		domain.ExternalTokenKindU2f,
		a.Title,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountAddExternalYubicoOtpToken(a *AccountAddExternalYubicoOtpToken, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountExternalTokenAdded(
		a.Account,
		event.RandomId(),
		domain.ExternalTokenKindYubicoOtp,
		a.Title,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountAddKeylist(a *AccountAddKeylist, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
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
		event.RandomId(),
		a.Title,
		keys,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountAddSshKey(a *AccountAddSshKey, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Id) == nil {
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
		event.RandomId(),
		privateKeyReformatted,
		publicKeyAuthorizedFormat,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) AccountAddOtpToken(a *AccountAddOtpToken, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	// just validate key
	if err := validateTotpProvisioningUrl(a.OtpProvisioningUrl); err != nil {
		return fmt.Errorf("invalid OtpProvisioningUrl: %s", err)
	}

	ctx.RaisesEvent(domain.NewAccountOtpTokenAdded(
		a.Account,
		event.RandomId(),
		a.OtpProvisioningUrl,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) DatabaseChangeMasterPassword(a *DatabaseChangeMasterPassword, ctx *command.Ctx) error {
	if err := verifyRepeatPassword(a.NewMasterPassword, a.NewMasterPasswordRepeat); err != nil {
		return err
	}

	ctx.RaisesEvent(domain.NewDatabaseMasterPasswordChanged(
		a.NewMasterPassword,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) SessionSignIn(a *SessionSignIn, ctx *command.Ctx) error {
	var user *state.SensitiveUser
	for _, userStorage := range h.state.DB.UserScope {
		if userStorage.SensitiveUser.User.Username == a.Username {
			user = &userStorage.SensitiveUser
			break
		}
	}

	if user == nil {
		return errors.New("bad username or password") // TODO: deduplicate sleep also here
	}

	upgradedPassword, err := storedpassword.Verify(
		storedpassword.StoredPassword(user.PasswordHash),
		a.Password,
		storedpassword.BuiltinStrategies)
	if err != nil {
		h.logl.Error.Printf("User %s failure signing in: %s", user.User.Username, err.Error())

		if err != storedpassword.ErrIncorrectPassword { // technical error
			return err
		}

		// to lessen efficacy of brute forcing. yes, `Verify()` by design is already slow,
		// but this is an addititional layer of protection.
		time.Sleep(2 * time.Second)
		return errors.New("bad username or password")
	}

	if upgradedPassword != "" {
		h.logl.Info.Printf(
			"Upgrading password of %s to %s",
			user.User.Username,
			storedpassword.CurrentBestDerivationStrategy.Id())

		ctx.RaisesEvent(domain.NewUserPasswordUpdated(
			user.User.Id,
			string(upgradedPassword),
			true, // => automatic upgrade of password
			event.Meta(time.Now(), user.User.Id)))
	}

	jwtSigner, err := httpauth.NewEcJwtSigner(h.state.GetJwtSigningKey())
	if err != nil {
		return err
	}

	token := jwtSigner.Sign(httpauth.UserDetails{
		Id: user.User.Id,
	})

	ctx.SetCookie = httpauth.ToCookie(token)

	ctx.RaisesEvent(domain.NewSessionSignedIn(
		ctx.RemoteAddr,
		ctx.UserAgent,
		event.Meta(time.Now(), user.User.Id)))

	h.logl.Info.Printf("User %s signed in", user.User.Username)

	return nil
}

func (h *CommandHandlers) SessionSignOut(a *SessionSignOut, ctx *command.Ctx) error {
	h.logl.Info.Printf("User %s signed out", ctx.Meta.UserId)

	ctx.SetCookie = httpauth.DeleteLoginCookie()

	// TODO: raise an event

	return nil
}

func (h *CommandHandlers) DatabaseExportToKeepass(a *DatabaseExportToKeepass, ctx *command.Ctx) error {
	return keepassexport.Export(h.state, ctx.Meta.UserId)
}

func (h *CommandHandlers) DatabaseUnseal(a *DatabaseUnseal, ctx *command.Ctx) error {
	if subtle.ConstantTimeCompare([]byte(h.state.GetMasterPassword()), []byte(a.MasterPassword)) != 1 {
		return errors.New("invalid password")
	}

	if h.state.IsUnsealed() {
		return errors.New("state already unsealed")
	}
	h.state.SetSealed(false)

	ctx.RaisesEvent(domain.NewDatabaseUnsealed(ctx.Meta))

	return nil
}

func (h *CommandHandlers) UserAddAccessToken(a *UserAddAccessToken, ctx *command.Ctx) error {
	if h.userData(ctx).SensitiveUser.AccessToken != "" {
		return errors.New("multiple access tokens not currently supported")
	}

	ctx.RaisesEvent(domain.NewUserAccessTokenAdded(
		a.User,
		event.RandomId(),
		cryptorandombytes.Base64Url(16),
		a.Description,
		ctx.Meta))

	return nil
}

func (h *CommandHandlers) UserCreate(a *UserCreate, ctx *command.Ctx) error {
	if err := verifyRepeatPassword(a.Password, a.PasswordRepeat); err != nil {
		return err
	}

	storedPassword, err := storedpassword.Store(
		a.Password,
		storedpassword.CurrentBestDerivationStrategy)
	if err != nil {
		return err
	}

	// FIXME: this has a race condition until the underlying event storage engine can
	//        support conditional append
	uid := h.state.NextFreeUserId()

	meta := event.Meta(time.Now(), uid)

	ctx.RaisesEvent(domain.NewUserCreated(
		uid,
		a.Username,
		meta))

	ctx.RaisesEvent(domain.NewUserPasswordUpdated(
		uid,
		string(storedPassword),
		false,
		meta))

	return nil
}

func (h *CommandHandlers) UserChangePassword(a *UserChangePassword, ctx *command.Ctx) error {
	// TODO: verify current password

	if err := verifyRepeatPassword(a.Password, a.PasswordRepeat); err != nil {
		return err
	}

	passwordHashed, err := storedpassword.Store(a.Password, storedpassword.CurrentBestDerivationStrategy)
	if err != nil {
		return err
	}

	ctx.RaisesEvent(domain.NewUserPasswordUpdated(a.User, string(passwordHashed), false, ctx.Meta))

	return nil
}

func (h *CommandHandlers) UserRegisterU2FToken(a *UserRegisterU2FToken, ctx *command.Ctx) error {
	var input apitypes.RegisterResponse
	if err := json.Unmarshal([]byte(a.Request), &input); err != nil {
		return err
	}

	regResp := u2f.RegisterResponse{
		Version:          input.RegisterResponse.Version,
		RegistrationData: input.RegisterResponse.RegistrationData,
		ClientData:       input.RegisterResponse.ClientData,
	}

	registration, err := u2f.Register(regResp, u2futil.ChallengeFromApiType(input.Challenge), &u2f.Config{
		// Chrome 66+ doesn't return the device's attestation
		// certificate by default.
		SkipAttestationVerify: true,
	})
	if err != nil {
		return err
	}

	registeredKey := u2futil.RegisteredKeyFromRegistration(*registration)

	ctx.RaisesEvent(domain.NewUserU2FTokenRegistered(
		a.Name,
		registeredKey.KeyHandle, // KeyHandle in Registration is binary for some reason..
		input.RegisterResponse.RegistrationData,
		input.RegisterResponse.ClientData,
		input.RegisterResponse.Version,
		ctx.Meta))

	return nil
}

func validateTotpProvisioningUrl(provisioningUrl string) error {
	key, err := otp.NewKeyFromURL(provisioningUrl)
	if err != nil {
		return err
	}

	// apparently NewKeyFromURL() can succeed even if GenerateCode() could fail,
	// so that's why we must actually go this far as to verify this
	if _, err := totp.GenerateCode(key.Secret(), time.Now()); err != nil {
		return err
	}

	return nil
}

func verifyRepeatPassword(pwd, pwdRepeat string) error {
	if pwd != pwdRepeat {
		return errors.New("password and repeated password different")
	}

	return nil
}
