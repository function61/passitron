package keepassimport

import (
	"encoding/csv"
	"github.com/function61/pi-security-module/accountevent"
	folderevent "github.com/function61/pi-security-module/folder/event"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
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

func Run(args []string) {
	if len(args) != 1 {
		log.Fatalf("Usage: <csv path>")
		return
	}

	csvPath := args[0]

	state.Initialize()
	defer state.Inst.Close()

	// TODO: expecting hardcoded password here in initialization phase.
	// this is not a catastropic security concern as after import we can
	// change master pwd.
	if err := state.Inst.Unseal("defaultpwd"); err != nil {
		panic(err)
	}

	result := parseGenericCsv(csvPath)

	foldersJustCreated := map[string]string{}

	events := []eventbase.EventInterface{}

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
			folderId = eventbase.RandomId()

			events = append(events, folderevent.FolderCreated{
				Event:    eventbase.NewEvent(),
				Id:       folderId,
				ParentId: "root",
				Name:     groupPath,
			})

			foldersJustCreated[groupPath] = folderId
		}

		accountId := eventbase.RandomId()

		creationTime, err := time.Parse("2006-01-02T15:04:05", res["Creation Time"])
		if err != nil {
			panic(err)
		}

		modificationTime, err := time.Parse("2006-01-02T15:04:05", res["Last Modification"])
		if err != nil {
			panic(err)
		}

		events = append(events, accountevent.AccountCreated{
			Event:    eventbase.NewEventWithTimestamp(creationTime),
			Id:       accountId,
			FolderId: folderId,
			Title:    res["Account"],
		})

		if res["Login Name"] != "" {
			events = append(events, accountevent.UsernameChanged{
				Event:    eventbase.NewEvent(),
				Id:       accountId,
				Username: res["Login Name"],
			})
		}

		if res["Password"] != "" {
			events = append(events, accountevent.PasswordAdded{
				Event:    eventbase.NewEventWithTimestamp(modificationTime),
				Account:  accountId,
				Id:       eventbase.RandomId(),
				Password: res["Password"],
			})
		}

		if res["Comments"] != "" {
			events = append(events, accountevent.DescriptionChanged{
				Event:       eventbase.NewEvent(),
				Id:          accountId,
				Description: res["Comments"],
			})
		}
	}

	state.Inst.EventLog.AppendBatch(events)

	log.Printf("%d event(s) applied", len(events))

	state.Inst.Save()

	log.Printf("State saved")
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
