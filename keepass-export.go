package main

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"github.com/function61/pi-security-module/state"
	"github.com/tobischo/gokeepasslib"
	"io"
	"log"
	"strconv"
)

func mkValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func mkProtectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value, Protected: true}}
}

func encryptPemBlock(plaintextBlock *pem.Block, password []byte) *pem.Block {
	ciphertextBlock, err := x509.EncryptPEMBlock(
		rand.Reader,
		plaintextBlock.Type,
		plaintextBlock.Bytes,
		password,
		x509.PEMCipher3DES)

	if err != nil {
		panic(err)
	}

	return ciphertextBlock
}

func exportRecursive(id string, meta *gokeepasslib.MetaData) (gokeepasslib.Group, int) {
	entriesExported := 0

	folder := state.FolderById(id)

	group := gokeepasslib.NewGroup()
	group.Name = folder.Name

	accounts := state.AccountsByFolder(folder.Id)

	for _, secureAccount := range accounts {
		account := secureAccount.ToInsecureAccount()

		for idx, secret := range account.Secrets {
			title := account.Title

			if idx > 0 { // append index, if many secrets in account
				title = title + " " + strconv.Itoa(idx)
			}

			entry := gokeepasslib.NewEntry()
			entry.Values = append(entry.Values, mkValue("Title", title))
			entry.Values = append(entry.Values, mkValue("UserName", account.Username))
			entry.Values = append(entry.Values, mkValue("Notes", account.Description))

			switch secret.Kind {
			default:
				panic("invalid secret kind: " + secret.Kind)
			case state.SecretKindPassword:
				entry.Values = append(entry.Values, mkProtectedValue("Password", secret.Password))

			case state.SecretKindSshKey:
				filename := account.Id + ".id_rsa"

				plaintextSshBlock, rest := pem.Decode([]byte(secret.SshPrivateKey))
				if len(rest) > 0 {
					panic("Extra data included in PEM content")
				}

				encryptedSshKey := encryptPemBlock(
					plaintextSshBlock,
					[]byte(state.Inst.GetMasterPassword()))

				binary := meta.Binaries.Add(pem.EncodeToMemory(encryptedSshKey))
				binaryReference := binary.CreateReference(filename)

				entry.Binaries = append(entry.Binaries, binaryReference)

			case state.SecretKindOtpToken:
				entry.Values = append(entry.Values, mkProtectedValue("Password", secret.OtpProvisioningUrl))

			}

			log.Printf("Appending account %s", account.Title)

			group.Entries = append(group.Entries, entry)

			entriesExported++
		}
	}

	subFolders := state.SubfoldersById(folder.Id)

	for _, subFolder := range subFolders {
		subGroup, subentriesExported := exportRecursive(subFolder.Id, meta)

		group.Groups = append(group.Groups, subGroup)

		entriesExported += subentriesExported
	}

	return group, entriesExported
}

func keepassExport(masterPassword string, output io.Writer) error {
	meta := gokeepasslib.NewMetaData()

	content := &gokeepasslib.DBContent{
		Meta: meta,
	}

	rootGroup, entriesExported := exportRecursive("root", meta)

	content.Root = &gokeepasslib.RootData{
		Groups: []gokeepasslib.Group{rootGroup},
	}

	db := &gokeepasslib.Database{
		Signature:   &gokeepasslib.DefaultSig,
		Headers:     gokeepasslib.NewFileHeaders(),
		Credentials: gokeepasslib.NewPasswordCredentials(masterPassword),
		Content:     content,
	}

	db.LockProtectedEntries()

	keepassEncoder := gokeepasslib.NewEncoder(output)
	if err := keepassEncoder.Encode(db); err != nil {
		return err
	}

	log.Printf("keepassExport: %d entries(s) exported", entriesExported)

	return nil
}
