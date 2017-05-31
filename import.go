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

	Replace group tree-less entries with General
*/

func importStuff() {
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

		if res["Group Tree"] == "" {
			// log.Printf("res = %v", res)
			log.Fatal("need group tree")
		}

		folder := folderByName(res["Group Tree"])

		folderId := ""
		if folder != nil {
			folderId = folder.Id
		} else if _, has := foldersJustCreated[res["Group Tree"]]; has {
			folderId = foldersJustCreated[res["Group Tree"]]
		} else {
			folderId = cryptorandombytes.Hex(4)

			events = append(events, FolderCreated{
				Id:       folderId,
				ParentId: "root",
				Name:     res["Group Tree"],
			})

			foldersJustCreated[res["Group Tree"]] = folderId
		}

		secretId := cryptorandombytes.Hex(4)

		events = append(events, SecretCreated{
			Id:       secretId,
			FolderId: folderId,
			Title:    res["Account"],
			Username: res["Login Name"],
			Password: res["Password"],
		})

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
