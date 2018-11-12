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

func exportKeylistAsText(wsecret state.WrappedSecret) string {
	lines := []string{
		"Keylist " + wsecret.Secret.Title,
		"--------------------------",
	}

	for _, entry := range wsecret.KeylistKeys {
		line := entry.Key + ": " + entry.Value

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func entryForAccount(account apitypes.Account, idx int, notesAppend string) *gokeepasslib.Entry {
	entry := gokeepasslib.NewEntry()
	title := account.Title

	if idx > 0 { // append index, if many secrets in account
		title = title + " " + strconv.Itoa(idx)
	}

	entry.Values = append(entry.Values, mkValue("Title", title))
	entry.Values = append(entry.Values, mkValue("UserName", account.Username))

	notes := account.Description

	if notesAppend != "" {
		if notes == "" {
			notes = notesAppend
		} else {
			notes = notes + "\n" + notesAppend
		}
	}

	entry.Values = append(entry.Values, mkValue("Notes", notes))

	return &entry
}

func exportRecursive(id string, meta *gokeepasslib.MetaData, st *state.State) (gokeepasslib.Group, int) {
	entriesExported := 0

	folder := st.FolderById(id)

	group := gokeepasslib.NewGroup()
	group.Name = folder.Name

	waccs := st.WrappedAccountsByFolder(folder.Id)

	for _, wacc := range waccs {
		for idx, secret := range wacc.Secrets {
			var entry *gokeepasslib.Entry = nil
			switch domain.SecretKindExhaustiveff1746(string(secret.Secret.Kind)) {
			case domain.SecretKindKeylist:
				entry = entryForAccount(wacc.Account, idx, exportKeylistAsText(secret))
			case domain.SecretKindPassword:
				entry = entryForAccount(wacc.Account, idx, "")
				entry.Values = append(entry.Values, mkProtectedValue("Password", secret.Secret.Password))
			case domain.SecretKindSshKey:
				entry = entryForAccount(wacc.Account, idx, "")
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
			case domain.SecretKindNote:
				entry = entryForAccount(wacc.Account, idx, secret.Secret.Note)
			case domain.SecretKindOtpToken:
				entry = entryForAccount(wacc.Account, idx, "")
				entry.Values = append(entry.Values, mkProtectedValue("Password", secret.OtpProvisioningUrl))
			default:
				panic("invalid secret kind: " + secret.Secret.Kind)
			}

			group.Entries = append(group.Entries, *entry)
			entriesExported++
		}

		// our datamodel differs somewhat from Keepass's (0-1 secrets per one account, keepass has exactly one),
		// so make sure entry gets created even if account doesn't have any secrets
		if len(wacc.Secrets) == 0 {
			group.Entries = append(group.Entries, *entryForAccount(wacc.Account, 0, ""))
			entriesExported++
		}
	}

	subFolders := st.SubfoldersByParentId(folder.Id)

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
