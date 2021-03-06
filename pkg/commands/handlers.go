package commands

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventkit/command"
	"github.com/function61/eventkit/httpcommand"
	"github.com/function61/gokit/cryptorandombytes"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/randompassword"
	"github.com/function61/gokit/storedpassword"
	"github.com/function61/passitron/pkg/apitypes"
	"github.com/function61/passitron/pkg/domain"
	"github.com/function61/passitron/pkg/keepassexport"
	"github.com/function61/passitron/pkg/state"
	"github.com/function61/passitron/pkg/u2futil"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/tstranex/u2f"
	"golang.org/x/crypto/ssh"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var (
	errAccountNotFound = errors.New("Account not found")
	errFolderNotFound  = errors.New("Folder not found")
)

type Handlers struct {
	state *state.AppState
	logl  *logex.Leveled
}

func (c *Handlers) userData(ctx *command.Ctx) *state.UserStorage {
	return c.state.User(ctx.Meta.UserId)
}

func New(state *state.AppState, logger *log.Logger) *Handlers {
	return &Handlers{state, logex.Levels(logger)}
}

func (h *Handlers) AccountRename(a *apitypes.AccountRename, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountRenamed(
		a.Account,
		a.Title,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountMove(a *apitypes.AccountMove, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountMoved(
		a.Account,
		a.NewParentFolder,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountChangeUsername(a *apitypes.AccountChangeUsername, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountUsernameChanged(
		a.Account,
		a.Username,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountChangeEmail(a *apitypes.AccountChangeEmail, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountEmailChanged(
		a.Account,
		a.Email,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountChangeUrl(a *apitypes.AccountChangeUrl, ctx *command.Ctx) error {
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

func (h *Handlers) AccountChangeDescription(a *apitypes.AccountChangeDescription, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountDescriptionChanged(
		a.Account,
		a.Description,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountDeleteSecret(a *apitypes.AccountDeleteSecret, ctx *command.Ctx) error {
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

func (h *Handlers) AccountCreateFolder(a *apitypes.AccountCreateFolder, ctx *command.Ctx) error {
	if h.userData(ctx).FolderById(a.Parent) == nil {
		return errFolderNotFound
	}

	ctx.RaisesEvent(domain.NewAccountFolderCreated(
		state.RandomId(),
		a.Parent,
		a.Name,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountDeleteFolder(a *apitypes.AccountDeleteFolder, ctx *command.Ctx) error {
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

func (h *Handlers) AccountRenameFolder(a *apitypes.AccountRenameFolder, ctx *command.Ctx) error {
	if h.userData(ctx).FolderById(a.Id) == nil {
		return errFolderNotFound
	}

	ctx.RaisesEvent(domain.NewAccountFolderRenamed(
		a.Id,
		a.Name,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountMoveFolder(a *apitypes.AccountMoveFolder, ctx *command.Ctx) error {
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

func (h *Handlers) AccountCreate(a *apitypes.AccountCreate, ctx *command.Ctx) error {
	accountId := state.RandomId()

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

	if a.Email != "" {
		ctx.RaisesEvent(domain.NewAccountEmailChanged(
			accountId,
			a.Email,
			ctx.Meta))
	}

	if a.Password != "" {
		// if PasswordRepeat given, verify it
		if err := verifyRepeatPassword(a.Password, a.PasswordRepeat); a.PasswordRepeat != "" && err != nil {
			return err
		}

		envelope, err := h.userData(ctx).Crypto().Encrypt([]byte(a.Password))
		if err != nil {
			return err
		}

		ctx.RaisesEvent(domain.NewAccountPasswordAdded(
			accountId,
			state.RandomId(),
			"", // not supported in this "quickly add password" use case
			envelope,
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

	ctx.CreatedRecordId(accountId)

	return nil
}

func (h *Handlers) AccountDelete(a *apitypes.AccountDelete, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Id) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountDeleted(
		a.Id,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountAddPassword(a *apitypes.AccountAddPassword, ctx *command.Ctx) error {
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

	envelope, err := h.userData(ctx).Crypto().Encrypt([]byte(password))
	if err != nil {
		return err
	}

	ctx.RaisesEvent(domain.NewAccountPasswordAdded(
		a.Account,
		state.RandomId(),
		a.Title,
		envelope,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountAddSecretNote(a *apitypes.AccountAddSecretNote, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	envelope, err := h.userData(ctx).Crypto().Encrypt([]byte(a.Note))
	if err != nil {
		return err
	}

	ctx.RaisesEvent(domain.NewAccountSecretNoteAdded(
		a.Account,
		state.RandomId(),
		a.Title,
		envelope,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountAddExternalU2FToken(a *apitypes.AccountAddExternalU2FToken, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountExternalTokenAdded(
		a.Account,
		state.RandomId(),
		domain.ExternalTokenKindU2f,
		a.Title,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountAddExternalYubicoOtpToken(a *apitypes.AccountAddExternalYubicoOtpToken, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	ctx.RaisesEvent(domain.NewAccountExternalTokenAdded(
		a.Account,
		state.RandomId(),
		domain.ExternalTokenKindYubicoOtp,
		a.Title,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountAddKeylist(a *apitypes.AccountAddKeylist, ctx *command.Ctx) error {
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

	keyExample := ""

	for i := 0; i < len(matches); i += 2 {
		item := domain.AccountKeylistAddedKeysItem{
			Key:   matches[i],
			Value: matches[i+1],
		}

		if keyExample == "" {
			keyExample = item.Key
		}

		if a.LengthOfKeys != 0 && len(item.Key) != a.LengthOfKeys {
			return errors.New("invalid length for key: " + item.Key)
		}

		if a.LengthOfValues != 0 && len(item.Value) != a.LengthOfValues {
			return errors.New("invalid length for value: " + item.Value)
		}

		keys = append(keys, item)
	}

	keysJson, err := json.Marshal(keys)
	if err != nil {
		return err
	}

	envelope, err := h.userData(ctx).Crypto().Encrypt(keysJson)
	if err != nil {
		return err
	}

	ctx.RaisesEvent(domain.NewAccountKeylistAdded(
		a.Account,
		state.RandomId(),
		a.Title,
		keyExample,
		envelope,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountAddSshKey(a *apitypes.AccountAddSshKey, ctx *command.Ctx) error {
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

	privateKeyReformatted := pem.EncodeToMemory(block)

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

	envelope, err := h.userData(ctx).Crypto().Encrypt([]byte(privateKeyReformatted))
	if err != nil {
		return err
	}

	ctx.RaisesEvent(domain.NewAccountSshKeyAdded(
		a.Id,
		state.RandomId(),
		envelope,
		publicKeyAuthorizedFormat,
		ctx.Meta))

	return nil
}

func (h *Handlers) AccountAddOtpToken(a *apitypes.AccountAddOtpToken, ctx *command.Ctx) error {
	if h.userData(ctx).WrappedAccountById(a.Account) == nil {
		return errAccountNotFound
	}

	if err := validateTotpProvisioningUrl(a.OtpProvisioningUrl); err != nil {
		return fmt.Errorf("invalid OtpProvisioningUrl: %s", err)
	}

	envelope, err := h.userData(ctx).Crypto().Encrypt([]byte(a.OtpProvisioningUrl))
	if err != nil {
		return err
	}

	ctx.RaisesEvent(domain.NewAccountOtpTokenAdded(
		a.Account,
		state.RandomId(),
		envelope,
		ctx.Meta))

	return nil
}

func (h *Handlers) UserChangeDecryptionKeyPassword(a *apitypes.UserChangeDecryptionKeyPassword, ctx *command.Ctx) error {
	if err := verifyRepeatPassword(a.NewMasterPassword, a.NewMasterPasswordRepeat); err != nil {
		return err
	}

	userDecryptionKeyChanged, err := h.userData(ctx).Crypto().ChangeDecryptionKeyPassword(
		a.NewMasterPassword,
		ctx.Meta)
	if err != nil {
		return err
	}

	ctx.RaisesEvent(userDecryptionKeyChanged)

	return nil
}

func (h *Handlers) SessionSignIn(a *apitypes.SessionSignIn, ctx *command.Ctx) error {
	userData := h.state.FindUserByUsername(a.Username)
	if userData == nil {
		return failAndSleepWithBadUsernameOrPassword()
	}

	sensitiveUser := userData.SensitiveUser()
	user := sensitiveUser.User

	upgradedPassword, err := storedpassword.Verify(
		storedpassword.StoredPassword(sensitiveUser.PasswordHash),
		a.Password,
		storedpassword.BuiltinStrategies)
	if err != nil {
		h.logl.Error.Printf("User %s failure signing in: %s", user.Username, err.Error())

		if err != storedpassword.ErrIncorrectPassword { // technical error
			return err
		}

		return failAndSleepWithBadUsernameOrPassword()
	}

	if upgradedPassword != "" {
		h.logl.Info.Printf(
			"Upgrading password of %s to %s",
			user.Username,
			storedpassword.CurrentBestDerivationStrategy.Id())

		ctx.RaisesEvent(domain.NewUserPasswordUpdated(
			user.Id,
			string(upgradedPassword),
			true, // => automatic upgrade of password
			ehevent.Meta(time.Now(), user.Id)))
	}

	if len(u2futil.GrabUsersU2FTokens(userData)) > 0 {
		if a.U2fChallengeResponse == nil {
			// triggers U2F flow in UI. this is a dirty hack to pass user ID + MAC to fetch
			// the U2F challenge..
			return httpcommand.NewHttpError(
				http.StatusUnauthorized,
				apitypes.ErrNeedU2fVerification,
				user.Id+":"+userData.SignInGetU2fChallengeMac().Sign())
		}

		u2fTokenUsedEvent, err := u2futil.SignatureOk(
			*a.U2fChallengeResponse,
			u2futil.ChallengeHashForSignIn(user.Id),
			userData)
		if err != nil {
			return err
		}

		// to audit & increase U2F counter (required mechanism in U2F)
		ctx.RaisesEvent(u2fTokenUsedEvent)
	}

	jwtSigner, err := httpauth.NewEcJwtSigner([]byte(h.state.ValidatedJwtConf().SigningKey))
	if err != nil {
		return err
	}

	token := jwtSigner.Sign(httpauth.UserDetails{
		Id: user.Id,
	}, time.Now())

	for _, cookie := range httpauth.ToCookiesWithCsrfProtection(token) {
		ctx.AddCookie(cookie)
	}

	ctx.RaisesEvent(domain.NewSessionSignedIn(
		ctx.RemoteAddr,
		ctx.UserAgent,
		ehevent.Meta(time.Now(), user.Id)))

	h.logl.Info.Printf("User %s signed in", user.Username)

	return nil
}

func (h *Handlers) SessionSignOut(a *apitypes.SessionSignOut, ctx *command.Ctx) error {
	h.logl.Info.Printf("User %s signed out", ctx.Meta.UserId)

	ctx.AddCookie(httpauth.DeleteLoginCookie())

	// TODO: raise an event

	return nil
}

func (h *Handlers) DatabaseExportToKeepass(a *apitypes.DatabaseExportToKeepass, ctx *command.Ctx) error {
	return keepassexport.Export(h.state, ctx.Meta.UserId, a.MasterPassword)
}

func (h *Handlers) UserUnlockDecryptionKey(a *apitypes.UserUnlockDecryptionKey, ctx *command.Ctx) error {
	if err := h.userData(ctx).Crypto().UnlockDecryptionKey(a.Password); err != nil {
		return err
	}

	ctx.RaisesEvent(domain.NewUserDecryptionKeyUnlocked(ctx.Meta))

	return nil
}

func (h *Handlers) UserAddAccessToken(a *apitypes.UserAddAccessToken, ctx *command.Ctx) error {
	if h.userData(ctx).SensitiveUser().AccessToken != "" {
		return errors.New("multiple access tokens not currently supported")
	}

	ctx.RaisesEvent(domain.NewUserAccessTokenAdded(
		a.User,
		state.RandomId(),
		cryptorandombytes.Base64Url(16),
		a.Description,
		ctx.Meta))

	return nil
}

func (h *Handlers) UserCreate(a *apitypes.UserCreate, ctx *command.Ctx) error {
	if err := verifyRepeatPassword(a.Password, a.PasswordRepeat); err != nil {
		return err
	}

	storedPassword, err := storedpassword.Store(
		a.Password,
		storedpassword.CurrentBestDerivationStrategy)
	if err != nil {
		return err
	}

	uid := state.RandomId()

	meta := ehevent.Meta(time.Now(), uid)

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

func (h *Handlers) UserChangePassword(a *apitypes.UserChangePassword, ctx *command.Ctx) error {
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

func (h *Handlers) UserRegisterU2FToken(a *apitypes.UserRegisterU2FToken, ctx *command.Ctx) error {
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
		// Also, probably should not be used anyway: https://www.imperialviolet.org/2018/03/27/webauthn.html#attestation
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

func failAndSleepWithBadUsernameOrPassword() error {
	// to lessen efficacy of brute forcing. yes, `storedpassword.Verify()` by design is
	// already slow, but this is an addititional layer of protection.
	time.Sleep(2 * time.Second)

	return errors.New("bad username or password")
}
