package commandhandlers

import (
	"github.com/function61/gokit/assert"
	"github.com/function61/pi-security-module/pkg/command"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/state"
	"testing"
	"time"
)

func TestScenario(t *testing.T) {
	st := state.NewTesting()

	accountId := createAccount(t, st)

	sshKey(t, st, accountId)
}

func createAccount(t *testing.T, st *state.State) string {
	ctx := defaultCtx(st)

	invoke(t, ctx, &AccountCreate{
		FolderId:       domain.RootFolderId,
		Title:          "my first account",
		Username:       "AzureDiamond",
		Password:       "hunter2",
		PasswordRepeat: "hunter2",
	})

	events := ctx.GetRaisedEvents()

	assert.True(t, len(events) == 3)

	e := events[0].(*domain.AccountCreated)

	return e.Id
}

func sshKey(t *testing.T, st *state.State, accountId string) {
	addFails := &AccountAddSshKey{
		Id:            accountId,
		SshPrivateKey: "invalid",
	}

	assert.True(t, addFails.Validate() == nil)
	assert.EqualString(t, addFails.Invoke(defaultCtx(st)).Error(), "Failed to parse PEM block")

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

	invoke(t, defaultCtx(st), &AccountAddSshKey{
		Id:            accountId,
		SshPrivateKey: dummyButWorkingSshKey,
	})

	acc := st.WrappedAccountById(accountId)

	secret := acc.Secrets[1]

	assert.EqualString(t,
		secret.Secret.SshPublicKeyAuthorized,
		"ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAIEA2By4FxK0z0yhhsGP7EOm+3Ze81ZuQf53xpan3o1W9dIAJHDtTUQgrCjwmIq23P/6kKPrJKRoev839TjyU6vXOGWgHYId97R5RuNv5VdZdHocTA662nkqy9UYLrULDTAfmXpuc7kS2d796RMlWNuJ8nBfK/Aaa4H9caZXkdSibQc=\n")
}

func defaultCtx(st *state.State) *command.Ctx {
	return &command.Ctx{
		State: st,
		Meta:  domain.Meta(time.Now(), domain.DefaultUserIdTODO),
	}
}

func invoke(t *testing.T, ctx *command.Ctx, cmd command.Command) {
	t.Helper()

	assert.True(t, cmd.Validate() == nil)
	assert.True(t, cmd.Invoke(ctx) == nil)

	ctx.State.EventLog.AppendBatch(ctx.GetRaisedEvents())
}
