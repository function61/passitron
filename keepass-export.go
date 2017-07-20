package main

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"encoding/xml"
	"github.com/function61/pi-security-module/state"
	"github.com/tobischo/gokeepasslib"
	"log"
	"os"
)

func xmlEscape(value string) string {
	escaped := &bytes.Buffer{}

	if err := xml.EscapeText(escaped, []byte(value)); err != nil {
		panic(err)
	}

	return escaped.String()
}

func mkValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: xmlEscape(value)}}
}

func mkProtectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: xmlEscape(value), Protected: true}}
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

func exportRecursive(id string, meta *gokeepasslib.MetaData) gokeepasslib.Group {
	folder := state.FolderById(id)

	group := gokeepasslib.NewGroup()
	group.Name = folder.Name

	secrets := state.SecretsByFolder(folder.Id)

	for _, secret := range secrets {
		insecureSecret := secret.ToInsecureSecret()

		entry := gokeepasslib.NewEntry()
		entry.Values = append(entry.Values, mkValue("Title", secret.Title))
		entry.Values = append(entry.Values, mkValue("UserName", secret.Username))
		entry.Values = append(entry.Values, mkProtectedValue("Password", insecureSecret.Password))

		if insecureSecret.SshPrivateKey != "" {
			filename := secret.Id + ".id_rsa"

			plaintextSshBlock, rest := pem.Decode([]byte(insecureSecret.SshPrivateKey))
			if len(rest) > 0 {
				panic("Extra data included in PEM content")
			}

			encryptedSshKey := encryptPemBlock(
				plaintextSshBlock,
				[]byte(state.Inst.GetMasterPassword()))

			binary := meta.Binaries.Add(pem.EncodeToMemory(encryptedSshKey))
			binaryReference := binary.CreateReference(filename)

			entry.Binaries = append(entry.Binaries, binaryReference)
		}

		log.Printf("Appending secret %s", secret.Title)

		group.Entries = append(group.Entries, entry)
	}

	subFolders := state.SubfoldersById(folder.Id)

	for _, subFolder := range subFolders {
		group.Groups = append(group.Groups, exportRecursive(subFolder.Id, meta))
	}

	return group
}

func keepassExport(password string) {
	keepassOutFile, err := os.Create("keepass-export.kdbx")
	if err != nil {
		panic(err)
	}

	defer keepassOutFile.Close()

	meta := gokeepasslib.NewMetaData()

	content := &gokeepasslib.DBContent{
		Meta: meta,
	}

	rootGroup := exportRecursive("root", meta)

	content.Root = &gokeepasslib.RootData{
		Groups: []gokeepasslib.Group{rootGroup},
	}

	db := &gokeepasslib.Database{
		Signature:   &gokeepasslib.DefaultSig,
		Headers:     gokeepasslib.NewFileHeaders(),
		Credentials: gokeepasslib.NewPasswordCredentials(password),
		Content:     content,
	}

	db.LockProtectedEntries()

	keepassEncoder := gokeepasslib.NewEncoder(keepassOutFile)
	if err := keepassEncoder.Encode(db); err != nil {
		panic(err)
	}
}
