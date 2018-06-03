package state

import (
	"github.com/function61/eventhorizon/util/ass"
	"github.com/function61/pi-security-module/pkg/domain"
	"testing"
	"time"
)

func TestScenario(t *testing.T) {
	st := NewTesting()

	firstAccountCreated(t, st)
	addPassword(t, st)
	addKeylist(t, st)
	renameAccount(t, st)
	deleteAccount(t, st)
	deleteKeylist(t, st)
	otpToken(t, st)

	waccs := st.WrappedAccountsByFolder(domain.RootFolderId)
	ass.EqualInt(t, len(waccs), 1)

	ass.EqualString(t, waccs[0].Account.Id, "acc1")
}

func meta() domain.EventMeta {
	return domain.Meta(time.Now(), domain.DefaultUserIdTODO)
}

func event(st *State, ev domain.Event) {
	st.EventLog.Append(ev)
}

func firstAccountCreated(t *testing.T, st *State) {
	event(st, domain.NewAccountCreated("acc1", domain.RootFolderId, "Example account", meta()))
	event(st, domain.NewAccountUsernameChanged("acc1", "AzureDiamond", meta()))
	event(st, domain.NewAccountDescriptionChanged("acc1", "my cool account", meta()))

	wacc := st.WrappedAccountById("acc1")

	ass.EqualString(t, wacc.Account.Id, "acc1")
	ass.EqualString(t, wacc.Account.Title, "Example account")
	ass.EqualString(t, wacc.Account.Username, "AzureDiamond")
	ass.EqualString(t, wacc.Account.Description, "my cool account")
}

func addPassword(t *testing.T, st *State) {
	event(st, domain.NewAccountPasswordAdded("acc1", "sec1", "hunter2", meta()))

	wacc := st.WrappedAccountById("acc1")

	ass.EqualInt(t, len(wacc.Secrets), 1)

	ass.EqualString(t, wacc.Secrets[0].Secret.Id, "sec1")
	ass.EqualString(t, string(wacc.Secrets[0].Secret.Kind), domain.SecretKindPassword)
	ass.EqualString(t, wacc.Secrets[0].Secret.Password, "hunter2")
}

func addKeylist(t *testing.T, st *State) {
	event(st, domain.NewAccountKeylistAdded("acc1", "sec2", "Keylist 1234", []domain.AccountKeylistAddedKeysItem{
		{Key: "01", Value: "9765"},
		{Key: "02", Value: "8421"},
		{Key: "03", Value: "1298"},
	}, meta()))

	wacc := st.WrappedAccountById("acc1")

	ass.EqualInt(t, len(wacc.Secrets), 2)

	ass.EqualString(t, wacc.Secrets[1].Secret.Id, "sec2")
	ass.EqualString(t, string(wacc.Secrets[1].Secret.Kind), domain.SecretKindKeylist)
	ass.EqualInt(t, len(wacc.Secrets[1].KeylistKeys), 3)
	ass.EqualString(t, wacc.Secrets[1].KeylistKeys[0].Key, "01")
	ass.EqualString(t, wacc.Secrets[1].KeylistKeys[0].Value, "9765")
	ass.EqualString(t, wacc.Secrets[1].KeylistKeys[1].Key, "02")
	ass.EqualString(t, wacc.Secrets[1].KeylistKeys[1].Value, "8421")
	ass.EqualString(t, wacc.Secrets[1].KeylistKeys[2].Key, "03")
	ass.EqualString(t, wacc.Secrets[1].KeylistKeys[2].Value, "1298")
}

func renameAccount(t *testing.T, st *State) {
	// before rename
	ass.EqualString(t, st.WrappedAccountById("acc1").Account.Title, "Example account")

	event(st, domain.NewAccountRenamed("acc1", "Renamed example account", meta()))

	wacc := st.WrappedAccountById("acc1")

	ass.EqualString(t, wacc.Account.Title, "Renamed example account")
}

func deleteAccount(t *testing.T, st *State) {
	ass.EqualInt(t, len(st.WrappedAccountsByFolder(domain.RootFolderId)), 1)

	event(st, domain.NewAccountCreated("acc2", domain.RootFolderId, "Example account", meta()))

	ass.EqualInt(t, len(st.WrappedAccountsByFolder(domain.RootFolderId)), 2)

	event(st, domain.NewAccountDeleted("acc2", meta()))

	ass.EqualInt(t, len(st.WrappedAccountsByFolder(domain.RootFolderId)), 1)
}

func deleteKeylist(t *testing.T, st *State) {
	ass.EqualInt(t, len(st.WrappedAccountById("acc1").Secrets), 2)

	event(st, domain.NewAccountSecretDeleted("acc1", "sec2", meta()))

	ass.EqualInt(t, len(st.WrappedAccountById("acc1").Secrets), 1)
}

func otpToken(t *testing.T, st *State) {
	ass.EqualInt(t, len(st.WrappedAccountById("acc1").Secrets), 1)

	event(st, domain.NewAccountOtpTokenAdded(
		"acc1",
		"sec3",
		"otpauth://totp/Google%3Afoo%40example.com?secret=qlt6vmy6svfx4bt4rpmisaiyol6hihca&issuer=Google",
		meta()))

	ass.EqualInt(t, len(st.WrappedAccountById("acc1").Secrets), 2)

	ass.EqualString(
		t,
		st.WrappedAccountById("acc1").Secrets[1].OtpProvisioningUrl,
		"otpauth://totp/Google%3Afoo%40example.com?secret=qlt6vmy6svfx4bt4rpmisaiyol6hihca&issuer=Google")
}
