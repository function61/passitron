package keepassimport

import (
	"encoding/csv"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/event"
	"github.com/function61/pi-security-module/pkg/state"
	"io"
	"log"
	"os"
	"time"
)

/*	Steps to make this work

	In Keepass 1.31 file > Export > CSV

	DO NOT "Encode/replace newline characters by \n"

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

func Run(csvPath string, userId string) {
	st := state.New()
	defer st.Close()

	csvFile, err := os.Open(csvPath)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	result, err := parseGenericCsv(csvFile)
	if err != nil {
		log.Fatalf("parseGenericCsv failed: %s", err.Error())
	}

	foldersJustCreated := map[string]string{}

	importStartedTime := time.Now()

	events := []event.Event{}

	pushEvent := func(e event.Event) {
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

		folderId := ""
		if _, has := foldersJustCreated[groupPath]; has {
			folderId = foldersJustCreated[groupPath]
		} else {
			folderId = event.RandomId()

			pushEvent(domain.NewAccountFolderCreated(
				folderId,
				domain.RootFolderId,
				groupPath,
				event.Meta(importStartedTime, userId)))

			foldersJustCreated[groupPath] = folderId
		}

		accountId := event.RandomId()

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
			event.Meta(creationTime, userId)))

		if res["Login Name"] != "" {
			pushEvent(domain.NewAccountUsernameChanged(
				accountId,
				res["Login Name"],
				event.Meta(modificationTime, userId)))
		}

		if res["Password"] != "" {
			pushEvent(domain.NewAccountPasswordAdded(
				accountId,
				event.RandomId(),
				res["Password"],
				event.Meta(modificationTime, userId)))
		}

		if res["Comments"] != "" {
			pushEvent(domain.NewAccountDescriptionChanged(
				accountId,
				res["Comments"],
				event.Meta(modificationTime, userId)))
		}
	}

	st.EventLog.AppendBatch(events)

	log.Printf("%d event(s) applied", len(events))
}

func parseGenericCsv(input io.Reader) ([]map[string]string, error) {
	csvReader := csv.NewReader(input)

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
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

	return result, nil
}
