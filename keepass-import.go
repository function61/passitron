package main

import (
	"./util/cryptorandombytes" // FIXME
	"encoding/csv"
	"log"
	"os"
)

/*	Steps to make this work

	In Keepass 1.31 file > Export > CSV

	DO NOT Encode/replace newline characters by \n

	Fields to export:

		Group Tree
		Group
		Title
		User Name
		Password
		URL
		Notes
		Creation Time
		Last Modification
		Attachment

	Convert to utf-8
	Replace \" with ""
*/

func keepassImport() {
	result := parseGenericCsv("keepass2.csv")

	foldersJustCreated := map[string]string{}

	events := []interface{}{}

	for _, res := range result {
		// skip attachments because practically all of them are SSH keys which
		// we'll enter manually in more structured format
		if res["Attachment"] != "" {
			log.Printf(
				"Skipping entry: %s attachment = %s\n",
				res["Account"],
				res["Attachment Description"])
			continue
		}

		groupPath := res["Group"]
		if res["Group Tree"] != "" {
			groupPath = res["Group Tree"] + "\\" + res["Group"]
		}

		if groupPath == "" {
			log.Fatal("need group path")
		}

		folder := folderByName(groupPath)

		folderId := ""
		if folder != nil {
			folderId = folder.Id
		} else if _, has := foldersJustCreated[groupPath]; has {
			folderId = foldersJustCreated[groupPath]
		} else {
			folderId = cryptorandombytes.Hex(4)

			events = append(events, FolderCreated{
				Id:       folderId,
				ParentId: "root",
				Name:     groupPath,
			})

			foldersJustCreated[groupPath] = folderId
		}

		secretId := cryptorandombytes.Hex(4)

		events = append(events, SecretCreated{
			Id:       secretId,
			FolderId: folderId,
			Title:    res["Account"],
		})

		if res["Login Name"] != "" {
			events = append(events, UsernameChanged{
				Id:       secretId,
				Username: res["Login Name"],
			})
		}

		if res["Password"] != "" {
			events = append(events, PasswordChanged{
				Id:       secretId,
				Password: res["Password"],
			})
		}

		if res["Comments"] != "" {
			events = append(events, DescriptionChanged{
				Id:          secretId,
				Description: res["Comments"],
			})
		}
	}

	ApplyEvents(events)

	log.Printf("%d event(s) applied", len(events))

	state.Save()

	log.Printf("State saved")
}

func folderByName(name string) *Folder {
	for _, f := range state.Folders {
		if f.Name == name {
			return &f
		}
	}

	return nil
}

func parseGenericCsv(filename string) []map[string]string {
	in, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(in)

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	result := []map[string]string{}

	headings := records[0]

	body := records[1:]

	for _, record := range body {
		res := map[string]string{}

		for idx, key := range headings {
			res[key] = record[idx]
		}

		result = append(result, res)
	}

	return result
}
