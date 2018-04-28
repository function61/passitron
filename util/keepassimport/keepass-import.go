package keepassimport

import (
	"encoding/csv"
	"github.com/function61/pi-security-module/domain"
	"github.com/function61/pi-security-module/state"
	"log"
	"os"
	"time"
)

/*	Steps to make this work

	In Keepass 1.31 file > Export > CSV

	DO NOT Encode/replace newline characters by \n

	Fields to export:

		Group Tree
		Password Groups
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

const defaultUser = "2"

func Run(args []string) {
	if len(args) != 1 {
		log.Fatalf("Usage: <csv path>")
		return
	}

	csvPath := args[0]

	state.Initialize()
	defer state.Inst.Close()

	result := parseGenericCsv(csvPath)

	foldersJustCreated := map[string]string{}

	importStartedTime := time.Now()

	events := []domain.Event{}

	pushEvent := func(e domain.Event) {
		events = append(events, e)
	}

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

		groupPath := res["Password Groups"]
		if res["Group Tree"] != "" {
			groupPath = res["Group Tree"] + "\\" + res["Password Groups"]
		}

		if groupPath == "" {
			log.Fatal("need group path")
		}

		folder := state.FolderByName(groupPath)

		folderId := ""
		if folder != nil {
			folderId = folder.Id
		} else if _, has := foldersJustCreated[groupPath]; has {
			folderId = foldersJustCreated[groupPath]
		} else {
			folderId = domain.RandomId()

			pushEvent(domain.NewAccountFolderCreated(
				folderId,
				"root",
				groupPath,
				domain.Meta(importStartedTime, defaultUser)))

			foldersJustCreated[groupPath] = folderId
		}

		accountId := domain.RandomId()

		creationTime, err := time.Parse("2006-01-02T15:04:05", res["Creation Time"])
		if err != nil {
			panic(err)
		}

		modificationTime, err := time.Parse("2006-01-02T15:04:05", res["Last Modification"])
		if err != nil {
			panic(err)
		}

		pushEvent(domain.NewAccountCreated(
			accountId,
			folderId,
			res["Account"],
			domain.Meta(creationTime, defaultUser)))

		if res["Login Name"] != "" {
			pushEvent(domain.NewAccountUsernameChanged(
				accountId,
				res["Login Name"],
				domain.Meta(modificationTime, defaultUser)))
		}

		if res["Password"] != "" {
			pushEvent(domain.NewAccountPasswordAdded(
				accountId,
				domain.RandomId(),
				res["Password"],
				domain.Meta(modificationTime, defaultUser)))
		}

		if res["Comments"] != "" {
			pushEvent(domain.NewAccountDescriptionChanged(
				accountId,
				res["Comments"],
				domain.Meta(modificationTime, defaultUser)))
		}
	}

	state.Inst.EventLog.AppendBatch(events)

	log.Printf("%d event(s) applied", len(events))
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
