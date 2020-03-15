package state

import (
	"context"
	"encoding/json"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/eventhorizon/pkg/ehreader/ehreadertest"
	"github.com/function61/gokit/assert"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/pi-security-module/pkg/domain"
	"testing"
	"time"
)

const (
	testStreamName = "/t-42/pism"
	joonasUid      = "1"
	testAccId      = "accId1"
)

var (
	t0 = time.Date(2020, 2, 20, 14, 2, 0, 0, time.UTC)
)

type testContext struct {
	user     *UserStorage
	eventLog *ehreadertest.EventLog
	reader   *ehreader.Reader
	ctx      context.Context
}

func (t *testContext) appendAndLoad(e ehevent.Event) {
	t.eventLog.AppendE(testStreamName, e)

	if err := t.reader.LoadUntilRealtime(t.ctx); err != nil {
		panic(err)
	}
}

func (t *testContext) encrypt(data string) []byte {
	env, err := t.user.Crypto().Encrypt([]byte(data))
	if err != nil {
		panic(err)
	}
	return env
}

func TestMain(t *testing.T) {
	tc := &testContext{
		user:     newUserStorage(ehreader.TenantId("42")),
		eventLog: ehreadertest.NewEventLog(),
		ctx:      context.Background(),
	}

	tc.reader = ehreader.New(tc.user, tc.eventLog, nil)

	setupUser(t, tc)

	unlockDecryptionKey(t, tc)

	signIn(t, tc)

	setupFolders(t, tc)

	renameFolder(t, tc)

	moveAndDeleteFolder(t, tc)

	addAccount(t, tc)

	addPassword(t, tc)

	addSecretNote(t, tc)

	addOtpToken(t, tc)

	addKeylist(t, tc)

	addExternalToken(t, tc)

	addSshKey(t, tc)

	secretUsed(t, tc)

	deleteSecret(t, tc)

	renameAccount(t, tc)

	moveAccount(t, tc)

	deleteAccount(t, tc)
}

func setupUser(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewUserCreated(joonasUid, "joonas", ehevent.MetaSystemUser(t0)))

	privKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPrivateKey([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0FPqri0cb2JZfXJ/DgYSF6vUp
wmJG8wVQZKjeGcjDOL5UlsuusFncCzWBQ7RKNUSesmQRMSGkVb1/3j+skZ6UtW+5u09lHNsj6tQ5
1s1SPrCBkedbNf0Tp0GbMJDyR4e9T04ZZwIDAQABAoGAFijko56+qGyN8M0RVyaRAXz++xTqHBLh
3tx4VgMtrQ+WEgCjhoTwo23KMBAuJGSYnRmoBZM3lMfTKevIkAidPExvYCdm5dYq3XToLkkLv5L2
pIIVOFMDG+KESnAFV7l2c+cnzRMW0+b6f8mR1CJzZuxVLL6Q02fvLi55/mbSYxECQQDeAw6fiIQX
GukBI4eMZZt4nscy2o12KyYner3VpoeE+Np2q+Z3pvAMd/aNzQ/W9WaI+NRfcxUJrmfPwIGm63il
AkEAxCL5HQb2bQr4ByorcMWm/hEP2MZzROV73yF41hPsRC9m66KrheO9HPTJuo3/9s5p+sqGxOlF
L0NDt4SkosjgGwJAFklyR1uZ/wPJjj611cdBcztlPdqoxssQGnh85BzCj/u3WqBpE2vjvyyvyI5k
X6zk7S0ljKtt2jny2+00VsBerQJBAJGC1Mg5Oydo5NwD6BiROrPxGo2bpTbu/fhrT8ebHkTz2epl
U9VQQSQzY1oZMVX8i1m5WUTLPz2yLJIBQVdXqhMCQBGoiuSoSjafUhV7i1cEGpb88h5NBYZzWXGZ
37sJ5QsW+sJyoNde3xH8vdXhzU7eT82D6X/scw9RZz+/6rCJ4p0=
-----END RSA PRIVATE KEY-----`))
	assert.Ok(t, err)

	userDecryptionKeyChanged, err := ExportPrivateKeyWithPassword(
		privKey,
		"myMasterPassword",
		ehevent.Meta(t0, joonasUid))
	assert.Ok(t, err)

	tc.appendAndLoad(userDecryptionKeyChanged)

	tc.appendAndLoad(
		domain.NewUserPasswordUpdated(
			joonasUid,
			"$pbkdf2-sha256-100k$_Ui6aWQtIAzyqL0nhzxZktjIpKh4KzuM4EzDRV8Ew-s$u1Yv0UYexUqpn6MtiZ_Obv7foqayElMc4_lWXX2DhV8", // nimda
			false,
			ehevent.Meta(t0, joonasUid)))

	tc.appendAndLoad(
		domain.NewUserAccessTokenAdded(
			joonasUid,
			"tid",
			"afsdjogfiast89asdkf",
			"Joonas's work computer",
			ehevent.Meta(t0, joonasUid)))

	assert.EqualJson(t, tc.user.sUser, `{
  "User": {
    "Created": "2020-02-20T14:02:00Z",
    "Id": "1",
    "PasswordLastChanged": "2020-02-20T14:02:00Z",
    "Username": "joonas"
  },
  "AccessToken": "afsdjogfiast89asdkf",
  "PasswordHash": "$pbkdf2-sha256-100k$_Ui6aWQtIAzyqL0nhzxZktjIpKh4KzuM4EzDRV8Ew-s$u1Yv0UYexUqpn6MtiZ_Obv7foqayElMc4_lWXX2DhV8"
}`)

	tc.appendAndLoad(
		domain.NewUserU2FTokenRegistered(
			"Joonas's primary U2F token",
			"keyHandle",
			"regData",
			"clientData",
			"version",
			ehevent.Meta(t0, joonasUid)))

	assert.EqualString(t, tc.user.u2FTokens[0].ClientData, "clientData")
	assert.Assert(t, tc.user.u2FTokens[0].Counter == 0)

	tc.appendAndLoad(
		domain.NewUserU2FTokenUsed(
			"keyHandle",
			314,
			ehevent.Meta(t0, joonasUid)))

	assert.Assert(t, tc.user.u2FTokens[0].Counter == 314)

	tc.appendAndLoad(
		domain.NewUserS3IntegrationConfigured(
			"myCoolBucket",
			"keyId",
			"secretLol",
			ehevent.Meta(t0, joonasUid)))

	assert.EqualJson(t, tc.user.s3ExportDetails, `{
  "Bucket": "myCoolBucket",
  "ApiKeyId": "keyId",
  "ApiKeySecret": "secretLol"
}`)
}

func unlockDecryptionKey(t *testing.T, tc *testContext) {
	assert.EqualString(
		t,
		tc.user.crypto.UnlockDecryptionKey("wrong password").Error(),
		"UnlockDecryptionKey: decryption error. wrong password?")

	assert.Ok(t, tc.user.crypto.UnlockDecryptionKey("myMasterPassword"))
}

func signIn(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewSessionSignedIn("127.0.0.1", "Mozilla Firefox v1.0", ehevent.Meta(t0, joonasUid)))
}

func setupFolders(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountFolderCreated("fld1", domain.RootFolderId, "General", ehevent.Meta(t0, joonasUid)))

	assert.EqualString(t, tc.user.folders[1].Name, "General")
}

func renameFolder(t *testing.T, tc *testContext) {
	assert.EqualString(t, tc.user.folders[1].Name, "General")

	tc.appendAndLoad(
		domain.NewAccountFolderRenamed("fld1", "General websites", ehevent.Meta(t0, joonasUid)))

	assert.EqualString(t, tc.user.folders[1].Name, "General websites")
}

func moveAndDeleteFolder(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountFolderCreated("fld2", domain.RootFolderId, "Folder to remove soon", ehevent.Meta(t0, joonasUid)))

	assert.Assert(t, len(tc.user.folders) == 3)

	assert.EqualString(t, tc.user.folders[2].ParentId, domain.RootFolderId)

	tc.appendAndLoad(
		domain.NewAccountFolderMoved("fld2", "fld1", ehevent.Meta(t0, joonasUid)))

	assert.EqualString(t, tc.user.folders[2].ParentId, "fld1")

	tc.appendAndLoad(
		domain.NewAccountFolderDeleted("fld2", ehevent.Meta(t0, joonasUid)))

	assert.Assert(t, len(tc.user.folders) == 2)
}

func addAccount(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountCreated(testAccId, domain.RootFolderId, "google.com", ehevent.Meta(t0, joonasUid)))

	tc.appendAndLoad(
		domain.NewAccountUrlChanged(testAccId, "https://google.com/", ehevent.Meta(t0, joonasUid)))

	tc.appendAndLoad(
		domain.NewAccountUsernameChanged(testAccId, "joonas@example.com", ehevent.Meta(t0, joonasUid)))

	tc.appendAndLoad(
		domain.NewAccountDescriptionChanged(testAccId, "Notes for account\nLine 2", ehevent.Meta(t0, joonasUid)))

	assert.EqualJson(t, tc.user.accounts[testAccId].Account, `{
  "Created": "2020-02-20T14:02:00Z",
  "Description": "Notes for account\nLine 2",
  "FolderId": "root",
  "Id": "accId1",
  "Title": "google.com",
  "Url": "https://google.com/",
  "Username": "joonas@example.com"
}`)
}

func addPassword(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountPasswordAdded(
			testAccId,
			"pwdId1",
			tc.encrypt("hunter2"),
			ehevent.Meta(t0, joonasUid)))

	secret := tc.user.accounts[testAccId].Secrets[0]

	assert.Assert(t, secret.Kind == domain.SecretKindPassword)

	pwd, err := tc.user.crypto.Decrypt(secret.Envelope)
	assert.Ok(t, err)

	assert.EqualString(t, string(pwd), "hunter2")
}

func addSecretNote(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountSecretNoteAdded(
			testAccId,
			"snId3",
			"Account recovery codes",
			tc.encrypt("01: abcd\n02: efgh\n03: ijkl\n04: mnop"),
			ehevent.Meta(t0, joonasUid)))

	secret := tc.user.accounts[testAccId].Secrets[1]

	assert.Assert(t, secret.Kind == domain.SecretKindNote)

	pwd, err := tc.user.crypto.Decrypt(secret.Envelope)
	assert.Ok(t, err)

	assert.EqualString(t, string(pwd), `01: abcd
02: efgh
03: ijkl
04: mnop`)
}

func addOtpToken(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountOtpTokenAdded(
			testAccId,
			"otpId4",
			tc.encrypt("otpauth://totp/Google%3Afoo%40example.com?secret=qlt6vmy6svfx4bt4rpmisaiyol6hihca&issuer=Google"),
			ehevent.Meta(t0, joonasUid)))

	secret := tc.user.accounts[testAccId].Secrets[2]

	assert.Assert(t, secret.Kind == domain.SecretKindOtpToken)

	otpProvisioningUrl, err := tc.user.DecryptOtpProvisioningUrl(secret)
	assert.Ok(t, err)

	assert.EqualString(
		t,
		otpProvisioningUrl,
		"otpauth://totp/Google%3Afoo%40example.com?secret=qlt6vmy6svfx4bt4rpmisaiyol6hihca&issuer=Google")
}

func addKeylist(t *testing.T, tc *testContext) {
	itemsJson, err := json.Marshal([]domain.AccountKeylistAddedKeysItem{
		{
			"01",
			"9876",
		},
		{
			"02",
			"5432",
		},
	})
	assert.Ok(t, err)

	tc.appendAndLoad(
		domain.NewAccountKeylistAdded(
			testAccId,
			"klId5",
			"Keylist 567",
			"01",
			tc.encrypt(string(itemsJson)),
			ehevent.Meta(t0, joonasUid)))

	secret := tc.user.accounts[testAccId].Secrets[3]
	assert.Assert(t, secret.Kind == domain.SecretKindKeylist)
	assert.EqualString(t, secret.Title, "Keylist 567")
	assert.EqualString(t, secret.keylistKeyExample, "01")

	klJson, err := tc.user.crypto.Decrypt(secret.Envelope)
	assert.Ok(t, err)

	kl := []domain.AccountKeylistAddedKeysItem{}
	assert.Ok(t, json.Unmarshal(klJson, &kl))

	assert.EqualJson(t, kl, `[
  {
    "Key": "01",
    "Value": "9876"
  },
  {
    "Key": "02",
    "Value": "5432"
  }
]`)
}

func addExternalToken(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountExternalTokenAdded(
			testAccId,
			"etId4",
			domain.ExternalTokenKindU2f,
			"Joonas's primary U2F token",
			ehevent.Meta(t0, joonasUid)))

	secret := tc.user.accounts[testAccId].Secrets[4]

	assert.Assert(t, secret.Kind == domain.SecretKindExternalToken)
	assert.Assert(t, *secret.externalTokenKind == domain.ExternalTokenKindU2f)

	assert.EqualString(t, secret.Title, "Joonas's primary U2F token")
}

func addSshKey(t *testing.T, tc *testContext) {
	dummyButWorkingKey := `-----BEGIN RSA PRIVATE KEY-----
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

	tc.appendAndLoad(
		domain.NewAccountSshKeyAdded(
			testAccId,
			"sshId5",
			tc.encrypt(dummyButWorkingKey),
			"fixme SshPublicKeyAuthorized",
			ehevent.Meta(t0, joonasUid)))

	secret := tc.user.accounts[testAccId].Secrets[5]

	assert.Assert(t, secret.Kind == domain.SecretKindSshKey)

	assert.EqualString(t, secret.SshPublicKeyAuthorized, "fixme SshPublicKeyAuthorized")

	sshKey, err := tc.user.crypto.Decrypt(secret.Envelope)
	assert.Ok(t, err)

	assert.EqualString(t, string(sshKey), dummyButWorkingKey)
}

func secretUsed(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountSecretUsed(
			testAccId,
			[]string{"klId5"},
			domain.SecretUsedTypeKeylistKeyExposed,
			"02",
			ehevent.Meta(t0, joonasUid)))
}

func deleteSecret(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountExternalTokenAdded(
			testAccId,
			"dummyId",
			domain.ExternalTokenKindU2f,
			"Dummy token",
			ehevent.Meta(t0, joonasUid)))

	assert.Assert(t, len(tc.user.accounts[testAccId].Secrets) == 7)

	tc.appendAndLoad(
		domain.NewAccountSecretDeleted(
			testAccId,
			"dummyId",
			ehevent.Meta(t0, joonasUid)))

	assert.Assert(t, len(tc.user.accounts[testAccId].Secrets) == 6)
}

func renameAccount(t *testing.T, tc *testContext) {
	tc.appendAndLoad(
		domain.NewAccountRenamed(testAccId, "google-is-evil.com", ehevent.Meta(t0, joonasUid)))

	assert.EqualString(
		t,
		tc.user.accounts[testAccId].Account.Title,
		"google-is-evil.com")
}

func moveAccount(t *testing.T, tc *testContext) {
	assert.EqualString(
		t,
		tc.user.accounts[testAccId].Account.FolderId,
		domain.RootFolderId)

	tc.appendAndLoad(
		domain.NewAccountMoved(testAccId, "fld1", ehevent.Meta(t0, joonasUid)))

	assert.EqualString(
		t,
		tc.user.accounts[testAccId].Account.FolderId,
		"fld1")
}

func deleteAccount(t *testing.T, tc *testContext) {
	accId := "accId2"

	tc.appendAndLoad(
		domain.NewAccountCreated(accId, domain.RootFolderId, "facebook.com", ehevent.Meta(t0, joonasUid)))

	assert.Assert(t, len(tc.user.accounts) == 2)

	tc.appendAndLoad(
		domain.NewAccountDeleted(accId, ehevent.Meta(t0, joonasUid)))

	assert.Assert(t, len(tc.user.accounts) == 1)
}

/*
const (
	testValidRsaPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0FPqri0cb2JZfXJ/DgYSF6vUp
wmJG8wVQZKjeGcjDOL5UlsuusFncCzWBQ7RKNUSesmQRMSGkVb1/3j+skZ6UtW+5u09lHNsj6tQ5
1s1SPrCBkedbNf0Tp0GbMJDyR4e9T04ZZwIDAQABAoGAFijko56+qGyN8M0RVyaRAXz++xTqHBLh
3tx4VgMtrQ+WEgCjhoTwo23KMBAuJGSYnRmoBZM3lMfTKevIkAidPExvYCdm5dYq3XToLkkLv5L2
pIIVOFMDG+KESnAFV7l2c+cnzRMW0+b6f8mR1CJzZuxVLL6Q02fvLi55/mbSYxECQQDeAw6fiIQX
GukBI4eMZZt4nscy2o12KyYner3VpoeE+Np2q+Z3pvAMd/aNzQ/W9WaI+NRfcxUJrmfPwIGm63il
AkEAxCL5HQb2bQr4ByorcMWm/hEP2MZzROV73yF41hPsRC9m66KrheO9HPTJuo3/9s5p+sqGxOlF
L0NDt4SkosjgGwJAFklyR1uZ/wPJjj611cdBcztlPdqoxssQGnh85BzCj/u3WqBpE2vjvyyvyI5k
X6zk7S0ljKtt2jny2+00VsBerQJBAJGC1Mg5Oydo5NwD6BiROrPxGo2bpTbu/fhrT8ebHkTz2epl
U9VQQSQzY1oZMVX8i1m5WUTLPz2yLJIBQVdXqhMCQBGoiuSoSjafUhV7i1cEGpb88h5NBYZzWXGZ
37sJ5QsW+sJyoNde3xH8vdXhzU7eT82D6X/scw9RZz+/6rCJ4p0=
-----END RSA PRIVATE KEY-----`
)
*/
