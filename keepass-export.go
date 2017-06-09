package main

import (
	"github.com/function61/pi-security-module/state"
	"github.com/tobischo/gokeepasslib"
	"log"
	"os"
)

func exportRecursive(id string) gokeepasslib.Group {
	folder := state.FolderById(id)

	group := gokeepasslib.NewGroup()
	group.Name = folder.Name

	secrets := state.SecretsByFolder(folder.Id)

	for _, secret := range secrets {
		entry := gokeepasslib.NewEntry()
		entry.Values = append(entry.Values, gokeepasslib.ValueData{Key: "Title", Value: gokeepasslib.V{Content: secret.Title}})
		entry.Values = append(entry.Values, gokeepasslib.ValueData{Key: "UserName", Value: gokeepasslib.V{Content: secret.Username}})
		entry.Values = append(entry.Values, gokeepasslib.ValueData{Key: "Password", Value: gokeepasslib.V{Content: secret.ToInsecureSecret().Password}})

		log.Printf("Appending secret %s", secret.Title)

		group.Entries = append(group.Entries, entry)
	}

	subFolders := state.SubfoldersById(folder.Id)

	for _, subFolder := range subFolders {
		group.Groups = append(group.Groups, exportRecursive(subFolder.Id))
	}

	return group
}

func keepassExport(password string) {
	keepassOutFile, err := os.Create("keepass-export.kdbx")
	if err != nil {
		panic(err)
	}

	defer keepassOutFile.Close()

	rootGroup := exportRecursive("root")

	root := &gokeepasslib.RootData{
		Groups: []gokeepasslib.Group{rootGroup},
	}

	content := &gokeepasslib.DBContent{
		Meta: gokeepasslib.NewMetaData(),
		Root: root,
	}

	db := &gokeepasslib.Database{
		Signature:   &gokeepasslib.DefaultSig,
		Headers:     gokeepasslib.NewFileHeaders(),
		Credentials: gokeepasslib.NewPasswordCredentials(password),
		Content:     content,
	}

	keepassEncoder := gokeepasslib.NewEncoder(keepassOutFile)
	if err := keepassEncoder.Encode(db); err != nil {
		panic(err)
	}
}
