package commandhandlers

import (
	"github.com/function61/gokit/assert"
	"github.com/function61/gokit/logex"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventkit/command"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"github.com/function61/pi-security-module/pkg/state"
	"testing"
	"time"
)

// chronologically successive small tests, but part of the bigger picture and the entire
// state of the database
func TestScenario(t *testing.T) {
	tstate := NewTestScenarioState(state.NewTesting())

	createAdminUser(t, tstate)

	changeAdminPassword(t, tstate)

	addAccessToken(t, tstate)

	tstate.firstAccountId = createAccount(t, tstate)

	signInAndSignOut(t, tstate)

	renameAccount(t, tstate)

	changeUsername(t, tstate)

	changeDescriptionAndUrl(t, tstate)

	sshKey(t, tstate)

	addPasswordAndRemoveIt(t, tstate)

	addSecretNoteAndRemoveIt(t, tstate)

	addOtpTokenAndRemoveIt(t, tstate)

	addKeylistAndRemoveIt(t, tstate)

	// this leaves "1st sub folder"
	createRenameMoveAndDeleteFolder(t, tstate)

	moveAccount(t, tstate)

	deleteAccount(t, tstate)

	// TODO: remove these bit by bit
	tstate.MarkCommandTested(Allocators["user.RegisterU2FToken"]())
	tstate.MarkCommandTested(Allocators["database.ExportToKeepass"]())
	tstate.MarkCommandTested(Allocators["database.Unseal"]())
	tstate.MarkCommandTested(Allocators["database.ChangeMasterPassword"]())
	tstate.MarkCommandTested(Allocators["user.Create"]())

	// make sure the scenario covered all commands that this application supports

	if len(tstate.GetUntestedCommands()) > 0 {
		t.Errorf("Untested commands: %v", tstate.GetUntestedCommands())
	}
}

func createAdminUser(t *testing.T, tstate *testScenarioState) {
	meta := tstate.DefaultCmdCtx().Meta

	userCreated := domain.NewUserCreated(
		meta.UserId,
		"admin",
		meta)

	// "nimda"
	storedPassword := "$pbkdf2-sha256-100k$PS7BCoKVAnYVrmZ-T_6m5_BaBNlnDEpK25rn4GnrdoQ$r1oMO-FMdkn1QEJlutQuEXOaacYhG9nNWvJ7GkVY4sM"

	passwordUpdated := domain.NewUserPasswordUpdated(
		meta.UserId,
		storedPassword,
		false,
		meta)

	if err := tstate.st.EventLog.Append([]event.Event{userCreated, passwordUpdated}); err != nil {
		panic(err)
	}
}

func changeAdminPassword(t *testing.T, tstate *testScenarioState) {
	cmdCtx := tstate.DefaultCmdCtx()

	tstate.InvokeSucceeds(t, cmdCtx, &UserChangePassword{
		User:           cmdCtx.Meta.UserId,
		Password:       "nimda2", // previous password was "nimda"
		PasswordRepeat: "nimda2",
	})
}

func addAccessToken(t *testing.T, tstate *testScenarioState) {
	cmdCtx := tstate.DefaultCmdCtx()

	assert.EqualString(t, tstate.st.DB.Users["2"].AccessToken, "")

	tstate.InvokeSucceeds(t, cmdCtx, &UserAddAccessToken{
		User: cmdCtx.Meta.UserId,
	})

	assert.Assert(t, len(tstate.st.DB.Users["2"].AccessToken) == 22)

	assert.EqualString(t, tstate.InvokeFails(t, cmdCtx, &UserAddAccessToken{
		User: cmdCtx.Meta.UserId,
	}), "multiple access tokens not currently supported")
}

func createAccount(t *testing.T, tstate *testScenarioState) string {
	cmdCtx := tstate.DefaultCmdCtx()

	tstate.InvokeSucceeds(t, cmdCtx, &AccountCreate{
		FolderId:       domain.RootFolderId,
		Url:            "https://www.example.com/login",
		Username:       "AzureDiamond",
		Password:       "hunter2",
		PasswordRepeat: "hunter2",
	})

	wacc := tstate.st.DB.WrappedAccounts[0]

	assert.EqualString(t, wacc.Account.Title, "www.example.com")
	assert.EqualString(t, wacc.Account.Url, "https://www.example.com/login")
	assert.EqualString(t, wacc.Account.Username, "AzureDiamond")

	return wacc.Account.Id
}

func signInAndSignOut(t *testing.T, tstate *testScenarioState) {
	ctx := tstate.DefaultCmdCtx()

	tstate.InvokeSucceeds(t, ctx, &SessionSignIn{
		Username: "admin",
		Password: "nimda2",
	})

	assert.Assert(t, ctx.SetCookie != nil)

	ctx = tstate.DefaultCmdCtx()

	tstate.InvokeSucceeds(t, ctx, &SessionSignOut{})

	assert.EqualString(t, ctx.SetCookie.Value, "del")
}

func renameAccount(t *testing.T, tstate *testScenarioState) {
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountRename{
		Account: tstate.firstAccountId,
		Title:   "www.example.com, renamed",
	})

	wacc := tstate.st.DB.WrappedAccounts[0]

	assert.EqualString(t, wacc.Account.Title, "www.example.com, renamed")
}

func changeUsername(t *testing.T, tstate *testScenarioState) {
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountChangeUsername{
		Account:  tstate.firstAccountId,
		Username: "joonas",
	})

	wacc := tstate.st.DB.WrappedAccounts[0]

	assert.EqualString(t, wacc.Account.Username, "joonas")
}

func changeDescriptionAndUrl(t *testing.T, tstate *testScenarioState) {
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountChangeDescription{
		Account:     tstate.firstAccountId,
		Description: "why hello there my friend",
	})

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountChangeUrl{
		Account: tstate.firstAccountId,
		Url:     "https://www.reddit.com/",
	})

	wacc := tstate.st.DB.WrappedAccounts[0]

	assert.EqualString(t, wacc.Account.Description, "why hello there my friend")
	assert.EqualString(t, wacc.Account.Url, "https://www.reddit.com/")
}

func sshKey(t *testing.T, tstate *testScenarioState) {
	assert.EqualString(t, tstate.InvokeFails(t, tstate.DefaultCmdCtx(), &AccountAddSshKey{
		Id:            tstate.firstAccountId,
		SshPrivateKey: "invalid",
	}), "Failed to parse PEM block")

	dummyButWorkingSshKey := `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQDYHLgXErTPTKGGwY/sQ6b7dl7zVm5B/nfGlqfejVb10gAkcO1N
RCCsKPCYirbc//qQo+skpGh6/zf1OPJTq9c4ZaAdgh33tHlG42/lV1l0ehxMDrra
eSrL1RgutQsNMB+Zem5zuRLZ3v3pEyVY24nycF8r8Bprgf1xpleR1KJtBwIBJQKB
gCMLk3lc+rnVE0ZI5ugK+HvOAY8+cr6X979Ws3AymHrjyKv2owWcWFNE6L7KYte6
zq+r4PEvauODVS6vSeQOBzlHhop49DiYzjIx/Slzmm4FoE4WehBY/5l2xw901HoW
nv1FJkXsEdWvtu8bw5GFJOlzBYCkpNRoEBk5myOl50DNAkEA+UdJR/tXyGN87c+e
9hZcvlHu0VFkE/756z2N/ysmCZMBZmdQp7YuCopavqSftnBVqlJoRr7piXi/Q8Rb
QJy/TwJBAN3wflBAmD3I6cFccW21cUPDJl14vEBY9OK5wWVsS5sNfj7wc+GZW242
ISlKt8Vgp9YVf7INuXbMFtSr2r+eSMkCQQChsbL+QisaMrHmXSjW+b+eC6HT4cRf
/1YAX0dZaBisQ63hjx+PYWn4/8wooiJogDeREtva3LMovQZxJWufiES9AkA1/Dpm
jEC1FTHww3WJY3xqbb05VLgrU+iKLS8K1ScltyyL2Z+llAF7rE1BZTOetqVdlobY
SIcPDwx464g8cpwVAkEAmXAzw61rhkMQDNaAMpNwINyhd1LM0nakzPJN4NeB5qJh
vD2QakbdLBUy2JF2E2GHmGyTXQ6yp4rWgcCVPeeFRw==
-----END RSA PRIVATE KEY-----`

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountAddSshKey{
		Id:            tstate.firstAccountId,
		SshPrivateKey: dummyButWorkingSshKey,
	})

	wacc := tstate.st.WrappedAccountById(tstate.firstAccountId)

	secret := wacc.Secrets[1]

	assert.Assert(t, secret.Secret.Kind == domain.SecretKindSshKey)
	assert.EqualString(t,
		secret.Secret.SshPublicKeyAuthorized,
		"ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAIEA2By4FxK0z0yhhsGP7EOm+3Ze81ZuQf53xpan3o1W9dIAJHDtTUQgrCjwmIq23P/6kKPrJKRoev839TjyU6vXOGWgHYId97R5RuNv5VdZdHocTA662nkqy9UYLrULDTAfmXpuc7kS2d796RMlWNuJ8nBfK/Aaa4H9caZXkdSibQc=\n")
}

func addPasswordAndRemoveIt(t *testing.T, tstate *testScenarioState) {
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountAddPassword{
		Account:        tstate.firstAccountId,
		Password:       "foobar",
		PasswordRepeat: "foobar",
	})

	wacc := tstate.st.WrappedAccountById(tstate.firstAccountId)

	assert.Assert(t, len(wacc.Secrets) == 3)

	newPassword := wacc.Secrets[2]

	assert.Assert(t, newPassword.Secret.Kind == domain.SecretKindPassword)
	assert.EqualString(t, newPassword.Secret.Password, "foobar")

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountDeleteSecret{
		Account: tstate.firstAccountId,
		Secret:  newPassword.Secret.Id,
	})

	assert.Assert(t, len(tstate.st.WrappedAccountById(tstate.firstAccountId).Secrets) == 2)
}

func addSecretNoteAndRemoveIt(t *testing.T, tstate *testScenarioState) {
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountAddSecretNote{
		Account: tstate.firstAccountId,
		Title:   "Account recovery codes",
		Note:    "01: foo    02: bar    ...",
	})

	wacc := tstate.st.WrappedAccountById(tstate.firstAccountId)

	assert.Assert(t, len(wacc.Secrets) == 3)

	note := wacc.Secrets[2]
	assert.Assert(t, note.Secret.Kind == domain.SecretKindNote)
	assert.EqualString(t, note.Secret.Title, "Account recovery codes")
	assert.EqualString(t, note.Secret.Note, "01: foo    02: bar    ...")

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountDeleteSecret{
		Account: tstate.firstAccountId,
		Secret:  note.Secret.Id,
	})

	assert.Assert(t, len(tstate.st.WrappedAccountById(tstate.firstAccountId).Secrets) == 2)
}

func addOtpTokenAndRemoveIt(t *testing.T, tstate *testScenarioState) {
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountAddOtpToken{
		Account:            tstate.firstAccountId,
		OtpProvisioningUrl: "otpauth://totp/Google%3Afoo%40example.com?secret=qlt6vmy6svfx4bt4rpmisaiyol6hihca&issuer=Google",
	})

	wacc := tstate.st.WrappedAccountById(tstate.firstAccountId)

	assert.Assert(t, len(wacc.Secrets) == 3)

	totp := wacc.Secrets[2]

	assert.Assert(t, totp.Secret.Kind == domain.SecretKindOtpToken)
	assert.EqualString(t, totp.OtpProvisioningUrl, "otpauth://totp/Google%3Afoo%40example.com?secret=qlt6vmy6svfx4bt4rpmisaiyol6hihca&issuer=Google")

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountDeleteSecret{
		Account: tstate.firstAccountId,
		Secret:  totp.Secret.Id,
	})

	assert.Assert(t, len(tstate.st.WrappedAccountById(tstate.firstAccountId).Secrets) == 2)
}

func addKeylistAndRemoveIt(t *testing.T, tstate *testScenarioState) {
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountAddKeylist{
		Account:          tstate.firstAccountId,
		Title:            "My legacy bank",
		ExpectedKeyCount: 3,
		LengthOfKeys:     2,
		LengthOfValues:   4,
		Keylist:          "01  1234\n02  5678\n03  9012\n",
	})

	wacc := tstate.st.WrappedAccountById(tstate.firstAccountId)

	assert.Assert(t, len(wacc.Secrets) == 3)

	keylist := wacc.Secrets[2]

	assert.Assert(t, keylist.Secret.Kind == domain.SecretKindKeylist)
	assert.EqualString(t, keylist.Secret.Title, "My legacy bank")
	assert.Assert(t, len(keylist.KeylistKeys) == 3)
	assert.EqualString(t, keylist.KeylistKeys[0].Key, "01")
	assert.EqualString(t, keylist.KeylistKeys[0].Value, "1234")
	assert.EqualString(t, keylist.KeylistKeys[1].Key, "02")
	assert.EqualString(t, keylist.KeylistKeys[1].Value, "5678")
	assert.EqualString(t, keylist.KeylistKeys[2].Key, "03")
	assert.EqualString(t, keylist.KeylistKeys[2].Value, "9012")

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountDeleteSecret{
		Account: tstate.firstAccountId,
		Secret:  keylist.Secret.Id,
	})

	assert.Assert(t, len(tstate.st.WrappedAccountById(tstate.firstAccountId).Secrets) == 2)
}

func createRenameMoveAndDeleteFolder(t *testing.T, tstate *testScenarioState) {
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountCreateFolder{
		Parent: domain.RootFolderId,
		Name:   "1st sub folder",
	})

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountCreateFolder{
		Parent: domain.RootFolderId,
		Name:   "2nd sub folder",
	})

	assert.Assert(t, len(tstate.st.DB.Folders) == 3)

	// both should be root's parents
	assert.EqualString(t, tstate.st.DB.Folders[1].ParentId, domain.RootFolderId)
	assert.EqualString(t, tstate.st.DB.Folders[2].ParentId, domain.RootFolderId)

	// now rename and move "2nd sub folder" under "1st sub folder"
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountRenameFolder{
		Id:   tstate.st.DB.Folders[2].Id,
		Name: "sub sub folder",
	})

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountMoveFolder{
		Id:        tstate.st.DB.Folders[2].Id,
		NewParent: tstate.st.DB.Folders[1].Id,
	})

	assert.EqualString(t, tstate.st.DB.Folders[2].Name, "sub sub folder")
	assert.EqualString(t, tstate.st.DB.Folders[2].ParentId, tstate.st.DB.Folders[1].Id)

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountDeleteFolder{
		Id: tstate.st.DB.Folders[2].Id,
	})

	assert.Assert(t, len(tstate.st.DB.Folders) == 2)
}

func moveAccount(t *testing.T, tstate *testScenarioState) {
	wacc := tstate.st.WrappedAccountById(tstate.firstAccountId)

	assert.EqualString(t, wacc.Account.FolderId, domain.RootFolderId)

	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountMove{
		Account:         tstate.firstAccountId,
		NewParentFolder: tstate.st.DB.Folders[1].Id,
	})

	wacc = tstate.st.WrappedAccountById(tstate.firstAccountId)

	assert.EqualString(t, wacc.Account.FolderId, tstate.st.DB.Folders[1].Id)
}

func deleteAccount(t *testing.T, tstate *testScenarioState) {
	tstate.InvokeSucceeds(t, tstate.DefaultCmdCtx(), &AccountDelete{
		Id: tstate.firstAccountId,
	})

	assert.Assert(t, len(tstate.st.DB.WrappedAccounts) == 0)
}

// the rest are utilities used for testing

// used to pass test context along
type testScenarioState struct {
	st               *state.AppState
	handlers         Handlers
	untestedCommands map[string]bool
	firstAccountId   string
}

func NewTestScenarioState(st *state.AppState) *testScenarioState {
	untestedCommands := map[string]bool{}

	for commandKey, _ := range Allocators {
		untestedCommands[commandKey] = true
	}

	return &testScenarioState{
		st:               st,
		handlers:         New(st, logex.Discard),
		untestedCommands: untestedCommands,
	}
}

func (tstate *testScenarioState) DefaultCmdCtx() *command.Ctx {
	return &command.Ctx{
		Meta: event.Meta(time.Now(), "2"),
	}
}

func (tstate *testScenarioState) GetUntestedCommands() []string {
	untested := []string{}

	for cmdKey, _ := range tstate.untestedCommands {
		untested = append(untested, cmdKey)
	}

	return untested
}

func (tstate *testScenarioState) MarkCommandTested(cmd command.Command) {
	delete(tstate.untestedCommands, cmd.Key())
}

func (tstate *testScenarioState) InvokeSucceeds(t *testing.T, ctx *command.Ctx, cmd command.Command) {
	t.Helper()

	if err := cmd.Validate(); err != nil {
		t.Error(err)
	}

	if err := cmd.Invoke(ctx, tstate.handlers); err != nil {
		t.Errorf("Command invoke failed: %s", err.Error())
	}

	if err := tstate.st.EventLog.Append(ctx.GetRaisedEvents()); err != nil {
		panic(err)
	}

	tstate.MarkCommandTested(cmd)
}

func (tstate *testScenarioState) InvokeFails(t *testing.T, ctx *command.Ctx, cmd command.Command) string {
	t.Helper()

	if err := cmd.Validate(); err != nil {
		t.Error(err)
	}

	err := cmd.Invoke(ctx, tstate.handlers)

	if err == nil {
		t.Error("expecting error")
	}

	return err.Error()
}
