package state

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const (
	statefilePath = "state.json"
)

func SubfoldersById(id string) []Folder {
	subFolders := []Folder{}

	for _, f := range Data.Folders {
		if f.ParentId != id {
			continue
		}

		subFolders = append(subFolders, f)
	}

	return subFolders
}

func SecretsByFolder(id string) []Secret {
	secrets := []Secret{}

	for _, s := range Data.Secrets {
		if s.FolderId != id {
			continue
		}

		secrets = append(secrets, s.ToSecureSecret())
	}

	return secrets
}

func SecretById(id string) *Secret {
	for _, s := range Data.Secrets {
		if s.Id == id {
			secret := s.ToSecureSecret()
			return &secret
		}
	}

	return nil
}

func FolderById(id string) *Folder {
	for _, f := range Data.Folders {
		if f.Id == id {
			return &f
		}
	}

	return nil
}

func FolderByName(name string) *Folder {
	for _, f := range Data.Folders {
		if f.Name == name {
			return &f
		}
	}

	return nil
}

func (s *Secret) GetPassword() *ExposedPassword {
	otpProof := ""

	if s.otpProvisioningUrl != "" {
		key, err := otp.NewKeyFromURL(s.otpProvisioningUrl)
		if err != nil {
			panic(err)
		}

		otpProof, err = totp.GenerateCode(key.Secret(), time.Now())
		if err != nil {
			panic(err)
		}
	}

	return &ExposedPassword{s.password, otpProof}
}

type ExposedPassword struct {
	Password string
	OtpProof string
}

// "secure secret": contains all fields from InsecureSecret *except* password
// is unexported
type Secret struct {
	Id                 string
	FolderId           string
	Title              string
	Username           string
	password           string
	otpProvisioningUrl string
	Description        string
	// created
	// password last changed
}

type InsecureSecret struct {
	Id                 string
	FolderId           string
	Title              string
	Username           string
	Password           string
	OtpProvisioningUrl string
	Description        string
	// created
	// password last changed
}

func (i *InsecureSecret) ToSecureSecret() Secret {
	return Secret{
		Id:                 i.Id,
		FolderId:           i.FolderId,
		Title:              i.Title,
		Username:           i.Username,
		password:           i.Password,
		otpProvisioningUrl: i.OtpProvisioningUrl,
		Description:        i.Description,
	}
}

func (s *Secret) ToInsecureSecret() InsecureSecret {
	return InsecureSecret{
		Id:                 s.Id,
		FolderId:           s.FolderId,
		Title:              s.Title,
		Username:           s.Username,
		Password:           s.password,
		OtpProvisioningUrl: s.otpProvisioningUrl,
		Description:        s.Description,
	}
}

type Folder struct {
	Id       string
	ParentId string
	Name     string
}

type Statefile struct {
	Secrets []InsecureSecret
	Folders []Folder
}

func passwordTo256BitEncryptionKey(pwd string, salt []byte) [32]byte {
	started := time.Now()

	// 1.4sec @ 100k on Raspberry Pi 2
	// https://github.com/borgbackup/borg/issues/77#issuecomment-130459726
	iterationCount := 100 * 1000
	encryptionKey := pbkdf2.Key(
		[]byte(pwd),
		salt,
		iterationCount,
		32,
		sha256.New)

	if len(encryptionKey) != 32 {
		panic("returned pbkdf2 key not 32 bytes")
	}

	var ret [32]byte
	copy(ret[:], encryptionKey)

	log.Printf("Encryption key derived from password in %s", time.Since(started))

	return ret
}

func (s *Statefile) Save() {
	jsonBytes, errJson := json.MarshalIndent(s, "", "    ")
	if errJson != nil {
		panic(errJson)
	}

	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}

	// using Seal() nonce as PBKDF2 salt
	encryptionKey := passwordTo256BitEncryptionKey("hunter2", nonce[:])

	encryptedBytes := []byte{}
	encryptedBytes = secretbox.Seal(nonce[:], jsonBytes, &nonce, &encryptionKey)

	if err := ioutil.WriteFile(statefilePath, encryptedBytes, 0644); err != nil {
		panic(err)
	}
}

func writeBlankStatefile() {
	rootFolder := Folder{
		Id:       "root",
		ParentId: "",
		Name:     "root",
	}

	Data = &Statefile{
		Secrets: []InsecureSecret{},
		Folders: []Folder{rootFolder},
	}

	Data.Save()
}

var Data *Statefile

func Initialize() {
	Data, _ = ReadStatefile()
}

func ReadStatefile() (*Statefile, error) {
	var s Statefile

	if _, err := os.Stat(statefilePath); os.IsNotExist(err) {
		log.Printf("Statefile does not exist. Initializing %s", statefilePath)

		writeBlankStatefile()
	}

	encryptedBytes, err := ioutil.ReadFile(statefilePath)
	if err != nil {
		panic(err)
	}

	// When you decrypt, you must use the same nonce and key you used to
	// encrypt the message. One way to achieve this is to store the nonce
	// alongside the encrypted message. Above, we stored the nonce in the first
	// 24 bytes of the encrypted text.
	// 24 bytes of nonce seems fine https://security.stackexchange.com/a/112592
	var decryptNonce [24]byte
	copy(decryptNonce[:], encryptedBytes[:24])

	// using Seal() nonce as PBKDF2 salt
	secretKeyBytes := passwordTo256BitEncryptionKey("hunter2", decryptNonce[:])

	jsonBytes, ok := secretbox.Open(nil, encryptedBytes[24:], &decryptNonce, &secretKeyBytes)
	if !ok {
		panic("decryption error. wrong password?")
	}

	if err := json.Unmarshal(jsonBytes, &s); err != nil {
		panic(err)
	}

	return &s, nil
}
