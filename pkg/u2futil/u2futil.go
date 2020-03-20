package u2futil

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/passitron/pkg/apitypes"
	"github.com/function61/passitron/pkg/domain"
	"github.com/function61/passitron/pkg/state"
	"github.com/tstranex/u2f"
	"strings"
	"time"
)

var cnFromSslCertificate = ""

func GetAppIdHostname() string {
	if cnFromSslCertificate == "" {
		panic("cnFromSslCertificate not set")
	}

	// - MUST be the origin (as in browser context origin)
	// - MUST be HTTPS
	// - MUST include the port (if non-default)
	return "https://" + cnFromSslCertificate
}

func InjectCommonNameFromSslCertificate(cert *x509.Certificate) {
	cnFromSslCertificate = cert.Subject.CommonName
}

func MakeTrustedFacets() []string {
	// TODO: find out what this is, and why we have to include app ID in it again..
	return []string{GetAppIdHostname()}
}

func U2ftokenToRegistration(u2ftoken *state.U2FToken) *u2f.Registration {
	reg := &u2f.Registration{}

	dataDecoded, err := decodeBase64(u2ftoken.RegistrationData)
	if err != nil {
		panic(err)
	}

	if err := reg.UnmarshalBinary(dataDecoded); err != nil {
		panic(err)
	}

	return reg
}

func GrabUsersU2FTokens(userStorage *state.UserStorage) []u2f.Registration {
	regs := []u2f.Registration{}

	for _, token := range userStorage.U2FTokens() {
		regs = append(regs, *U2ftokenToRegistration(token))
	}

	return regs
}

func GrabUsersU2FTokenByKeyHandle(userStorage *state.UserStorage, keyHandle string) *state.U2FToken {
	for _, token := range userStorage.U2FTokens() {
		if token.KeyHandle == keyHandle {
			return token
		}
	}

	return nil
}

// this ugly hack because the API is so lacking
func RegisteredKeyFromRegistration(registration u2f.Registration) u2f.RegisteredKey {
	dummyChallenge, err := u2f.NewChallenge(GetAppIdHostname(), MakeTrustedFacets())
	if err != nil {
		panic(err)
	}

	dummyRegisterRequest := u2f.NewWebRegisterRequest(dummyChallenge, []u2f.Registration{registration})

	return dummyRegisterRequest.RegisteredKeys[0]
}

// copy-pasted from u2f library because the relevant API was not exported
func decodeBase64(s string) ([]byte, error) {
	return base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(s)
}

// conversions to/from generated apitypes

func RegisteredKeysToApiType(input []u2f.RegisteredKey) []apitypes.U2FRegisteredKey {
	keys := []apitypes.U2FRegisteredKey{}
	for _, key := range input {
		keys = append(keys, apitypes.U2FRegisteredKey{
			Version:   key.Version,
			KeyHandle: key.KeyHandle,
			AppID:     key.AppID,
		})
	}

	return keys
}

func SignRequestToApiType(input u2f.WebSignRequest) apitypes.U2FSignRequest {
	return apitypes.U2FSignRequest{
		AppID:          input.AppID,
		Challenge:      input.Challenge,
		RegisteredKeys: RegisteredKeysToApiType(input.RegisteredKeys),
	}
}

// FIXME: remove need for these To/From conversions by supporting lowercase keys in apitypes
// (problem: lowercase in Golang is unexported field)

func ChallengeToApiType(input u2f.Challenge) apitypes.U2FChallenge {
	return apitypes.U2FChallenge{
		Challenge:     base64.StdEncoding.EncodeToString(input.Challenge),
		Timestamp:     input.Timestamp,
		AppID:         input.AppID,
		TrustedFacets: input.TrustedFacets,
	}
}

func ChallengeFromApiType(input apitypes.U2FChallenge) u2f.Challenge {
	challenge, err := base64.StdEncoding.DecodeString(input.Challenge)
	if err != nil {
		panic(err)
	}

	return u2f.Challenge{
		Challenge:     challenge,
		Timestamp:     input.Timestamp,
		AppID:         input.AppID,
		TrustedFacets: input.TrustedFacets,
	}
}

func SignResponseFromApiType(input apitypes.U2FSignResult) u2f.SignResponse {
	return u2f.SignResponse{
		KeyHandle:     input.KeyHandle,
		SignatureData: input.SignatureData,
		ClientData:    input.ClientData,
	}
}

// this API should be offered by tstranex/u2f
func NewU2FCustomChallenge(challenge [32]byte) (*u2f.Challenge, error) {
	return &u2f.Challenge{
		Challenge:     challenge[:],
		Timestamp:     time.Now(),
		AppID:         GetAppIdHostname(),
		TrustedFacets: MakeTrustedFacets(),
	}, nil
}

func ChallengeHashForAccountSecrets(account apitypes.Account) [32]byte {
	return stringToU2FChallengeHash("accountsecrets", account.Id)
}

func ChallengeHashForSignIn(userId string) [32]byte {
	return stringToU2FChallengeHash("signin", userId)
}

func ChallengeHashForKeylistKey(accountId, secretId, keylistKey string) [32]byte {
	return stringToU2FChallengeHash("keylistkey", accountId, secretId, keylistKey)
}

func stringToU2FChallengeHash(components ...string) [32]byte {
	// validate because if one of these is empty, it'd be catastropic for security
	for _, component := range components {
		if component == "" {
			panic("stringToU2FChallengeHash: component empty")
		}
	}

	return sha256.Sum256([]byte(strings.Join(components, ":")))
}

func SignatureOk(
	response apitypes.U2FResponseBundle,
	expectedHash [32]byte,
	userStorage *state.UserStorage,
) (*domain.UserU2FTokenUsed, error) {
	nativeChallenge := ChallengeFromApiType(response.Challenge)

	if !bytes.Equal(nativeChallenge.Challenge, expectedHash[:]) {
		return nil, errors.New("invalid challenge hash")
	}

	u2ftoken := GrabUsersU2FTokenByKeyHandle(userStorage, response.SignResult.KeyHandle)
	if u2ftoken == nil {
		return nil, fmt.Errorf("U2F token not found by KeyHandle: %s", response.SignResult.KeyHandle)
	}

	newCounter, authErr := U2ftokenToRegistration(u2ftoken).Authenticate(
		SignResponseFromApiType(response.SignResult),
		nativeChallenge,
		u2ftoken.Counter)
	if authErr != nil {
		return nil, authErr
	}

	u2fTokenUsedEvent := domain.NewUserU2FTokenUsed(
		response.SignResult.KeyHandle,
		int(newCounter),
		ehevent.Meta(time.Now(), userStorage.UserId()))

	return u2fTokenUsedEvent, nil
}

func MakeChallengeBundle(
	challengeHash [32]byte,
	userData *state.UserStorage,
) (*apitypes.U2FChallengeBundle, error) {
	tokens := GrabUsersU2FTokens(userData)
	if len(tokens) == 0 {
		return nil, errors.New("makeChallengeBundle: no U2F tokens found")
	}

	challenge, err := NewU2FCustomChallenge(challengeHash)
	if err != nil {
		return nil, err
	}

	return &apitypes.U2FChallengeBundle{
		Challenge:   ChallengeToApiType(*challenge),
		SignRequest: SignRequestToApiType(*challenge.SignRequest(tokens)),
	}, nil
}
