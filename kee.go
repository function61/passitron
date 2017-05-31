package main

import (
	"github.com/tobischo/gokeepasslib"
	"os"
)

func writeKee(password string) {
	keepassOutFile, err := os.Create("kee.kdbx")
	if err != nil {
		panic(err)
	}

	defer keepassOutFile.Close()

	// testGroup := gokeepasslib.NewGroup()
	// testGroup.Name = "hello"

	/*
		content := gokeepasslib.NewDBContent()
		content.Root.Groups = append(content.Root.Groups, testGroup)
	*/

	group := gokeepasslib.NewGroup()
	group.Name = "loq_export"

	entry := gokeepasslib.NewEntry()
	entry.Values = append(entry.Values, gokeepasslib.ValueData{Key: "Title", Value: gokeepasslib.V{Content: "op.fi"}})
	entry.Values = append(entry.Values, gokeepasslib.ValueData{Key: "UserName", Value: gokeepasslib.V{Content: "pompo500"}})
	entry.Values = append(entry.Values, gokeepasslib.ValueData{Key: "Password", Value: gokeepasslib.V{Content: "1234"}})

	group.Entries = append(group.Entries, entry)

	root := &gokeepasslib.RootData{
		Groups: []gokeepasslib.Group{group},
	}

	content := &gokeepasslib.DBContent{
		Meta: gokeepasslib.NewMetaData(),
		Root: root,
	}

	// db := gokeepasslib.NewDatabase()
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
