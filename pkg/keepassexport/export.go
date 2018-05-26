package keepassexport

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/mattetti/filebuffer"
	"github.com/tobischo/gokeepasslib"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

func Export(st *state.State) error {
	if st.S3ExportBucket == "" || st.S3ExportApiKey == "" || st.S3ExportSecret == "" {
		return errors.New("S3ExportBucket, S3ExportApiKey or S3ExportSecret undefined")
	}

	var keepassOutFile bytes.Buffer

	if err := keepassExport(st.GetMasterPassword(), &keepassOutFile, st); err != nil {
		return err
	}

	manualCredential := credentials.NewStaticCredentials(
		st.S3ExportApiKey,
		st.S3ExportSecret,
		"")

	awsSession, errSession := session.NewSession()
	if errSession != nil {
		return errSession
	}

	s3Client := s3.New(awsSession, aws.NewConfig().WithCredentials(manualCredential).WithRegion(endpoints.UsEast1RegionID))

	// why filebuffer?
	// https://stackoverflow.com/questions/20602131/io-writeseeker-and-io-readseeker-from-byte-or-file#comment60685488_20602219

	remotePath := "/databases/" + time.Now().UTC().Format(time.RFC3339) + ".kdbx"

	_, errS3Put := s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(st.S3ExportBucket),
		Key:    aws.String(remotePath),
		Body:   filebuffer.New(keepassOutFile.Bytes()),
	})
	if errS3Put != nil {
		return errS3Put
	}

	log.Printf("Keepass database uploaded to %s:%s", st.S3ExportBucket, remotePath)

	return nil
}

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

func exportKeylistAsText(secret apitypes.Secret) string {
	lines := []string{
		"Keylist " + secret.Title,
		"--------------------------",
	}

	for _, entry := range secret.KeylistKeys {
		line := entry.Key + ": " + entry.Value

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func exportRecursive(id string, meta *gokeepasslib.MetaData, st *state.State) (gokeepasslib.Group, int) {
	entriesExported := 0

	folder := st.FolderById(id)

	group := gokeepasslib.NewGroup()
	group.Name = folder.Name

	waccs := st.WrappedAccountsByFolder(folder.Id)

	for _, wacc := range waccs {
		for idx, secret := range wacc.Secrets {
			title := wacc.Account.Title

			if idx > 0 { // append index, if many secrets in account
				title = title + " " + strconv.Itoa(idx)
			}

			entry := gokeepasslib.NewEntry()
			entry.Values = append(entry.Values, mkValue("Title", title))
			entry.Values = append(entry.Values, mkValue("UserName", wacc.Account.Username))

			notes := wacc.Account.Description

			switch domain.SecretKindExhaustive44d6e3(string(secret.Secret.Kind)) {
			case domain.SecretKindKeylist:
				keylistAsText := exportKeylistAsText(secret.Secret)

				if notes == "" {
					notes = keylistAsText
				} else {
					notes = notes + "\n" + keylistAsText
				}
			case domain.SecretKindPassword:
				entry.Values = append(entry.Values, mkProtectedValue("Password", secret.Secret.Password))

			case domain.SecretKindSshKey:
				filename := wacc.Account.Id + ".id_rsa"

				plaintextSshBlock, rest := pem.Decode([]byte(secret.SshPrivateKey))
				if len(rest) > 0 {
					panic("Extra data included in PEM content")
				}

				encryptedSshKey := encryptPemBlock(
					plaintextSshBlock,
					[]byte(st.GetMasterPassword()))

				binary := meta.Binaries.Add(pem.EncodeToMemory(encryptedSshKey))
				binaryReference := binary.CreateReference(filename)

				entry.Binaries = append(entry.Binaries, binaryReference)

			case domain.SecretKindOtpToken:
				entry.Values = append(entry.Values, mkProtectedValue("Password", secret.OtpProvisioningUrl))

			default:
				panic("invalid secret kind: " + secret.Secret.Kind)
			}

			entry.Values = append(entry.Values, mkValue("Notes", notes))

			group.Entries = append(group.Entries, entry)

			entriesExported++
		}
	}

	subFolders := st.SubfoldersById(folder.Id)

	for _, subFolder := range subFolders {
		subGroup, subentriesExported := exportRecursive(subFolder.Id, meta, st)

		group.Groups = append(group.Groups, subGroup)

		entriesExported += subentriesExported
	}

	return group, entriesExported
}

func keepassExport(masterPassword string, output io.Writer, st *state.State) error {
	meta := gokeepasslib.NewMetaData()

	content := &gokeepasslib.DBContent{
		Meta: meta,
	}

	rootGroup, entriesExported := exportRecursive(domain.RootFolderId, meta, st)

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
