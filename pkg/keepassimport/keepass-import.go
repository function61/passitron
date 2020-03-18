package keepassimport

import (
	"encoding/csv"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/pi-security-module/pkg/domain"
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

func Run(csvPath string, userId string) error {
	st, err := state.New(nil)
	if err != nil {
		return err
	}

	userCrypto := st.User(userId).Crypto()

	csvFile, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	result, err := parseGenericCsv(csvFile)
	if err != nil {
		return fmt.Errorf("parseGenericCsv failed: %s", err.Error())
	}

	foldersJustCreated := map[string]string{}

	importStartedTime := time.Now()

	events := []ehevent.Event{}

	pushEvent := func(e ehevent.Event) {
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
			folderId = state.RandomId()

			pushEvent(domain.NewAccountFolderCreated(
				folderId,
				domain.RootFolderId,
				groupPath,
				ehevent.Meta(importStartedTime, userId)))

			foldersJustCreated[groupPath] = folderId
		}

		accountId := state.RandomId()

		creationTime, err := time.Parse("2006-01-02T15:04:05", res["Creation Time"])
		if err != nil {
			return err
		}

		modificationTime, err := time.Parse("2006-01-02T15:04:05", res["Last Modification"])
		if err != nil {
			return err
		}

		pushEvent(domain.NewAccountCreated(
			accountId,
			folderId,
			res["Account"],
			ehevent.Meta(creationTime, userId)))

		if res["Login Name"] != "" {
			pushEvent(domain.NewAccountUsernameChanged(
				accountId,
				res["Login Name"],
				ehevent.Meta(modificationTime, userId)))
		}

		if res["Password"] != "" {
			envelope, err := userCrypto.Encrypt([]byte(res["Password"]))
			if err != nil {
				return err
			}

			pushEvent(domain.NewAccountPasswordAdded(
				accountId,
				state.RandomId(),
				envelope,
				ehevent.Meta(modificationTime, userId)))
		}

		if res["Comments"] != "" {
			pushEvent(domain.NewAccountDescriptionChanged(
				accountId,
				res["Comments"],
				ehevent.Meta(modificationTime, userId)))
		}
	}

	if err := st.EventLog.Append(events); err != nil {
		return err
	}

	log.Printf("%d event(s) applied", len(events))

	return nil
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
