package eventlog

import (
	"bufio"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	folderevent "github.com/function61/pi-security-module/folder/event"
	sessionevent "github.com/function61/pi-security-module/session/event"
	"github.com/function61/pi-security-module/util/eventbase"
	"log"
	"os"
	"regexp"
)

const (
	logfilePath = "events.log"
)

var evParseRe = regexp.MustCompile("^([^ ]+) (.+)")

func ReadOldEvents() error {
	if _, err := os.Stat(logfilePath); os.IsNotExist(err) {
		log.Fatal("events.log not present. Either create it or resolve the problem.")
	}

	logFile, err := os.Open(logfilePath)
	if err != nil {
		panic(err)
	}

	logLineScanner := bufio.NewScanner(logFile)
	logLineScanner.Split(bufio.ScanLines)

	eventsRead := 0

	for logLineScanner.Scan() {
		match := evParseRe.FindSubmatch(logLineScanner.Bytes())
		if len(match) != 3 {
			panic(errors.New("failed to parse log event"))
		}

		payload := match[2]

		var e eventbase.EventInterface
		switch string(match[1]) {
		default:
			panic(errors.New("unknown event type: " + string(match[1])))
		case "FolderCreated":
			e = folderevent.FolderCreatedFromSerialized(payload)
		case "AccountCreated":
			e = accountevent.AccountCreatedFromSerialized(payload)
		case "UsernameChanged":
			e = accountevent.UsernameChangedFromSerialized(payload)
		case "PasswordAdded":
			e = accountevent.PasswordAddedFromSerialized(payload)
		case "DescriptionChanged":
			e = accountevent.DescriptionChangedFromSerialized(payload)
		case "DatabaseUnsealed":
			e = sessionevent.DatabaseUnsealedFromSerialized(payload)
		case "SecretUsed":
			e = accountevent.SecretUsedFromSerialized(payload)
		case "MasterPasswordChanged":
			e = sessionevent.MasterPasswordChangedFromSerialized(payload)
		case "S3IntegrationConfigured":
			e = sessionevent.S3IntegrationConfiguredFromSerialized(payload)
		case "AccountDeleted":
			e = accountevent.AccountDeletedFromSerialized(payload)
		case "AccountRenamed":
			e = accountevent.AccountRenamedFromSerialized(payload)
		case "SecretDeleted":
			e = accountevent.SecretDeletedFromSerialized(payload)
		case "SshKeyAdded":
			e = accountevent.SshKeyAddedFromSerialized(payload)
		case "OtpTokenAdded":
			e = accountevent.OtpTokenAddedFromSerialized(payload)
		case "FolderMoved":
			e = folderevent.FolderMovedFromSerialized(payload)
		case "FolderRenamed":
			e = folderevent.FolderRenamedFromSerialized(payload)
		}

		// bypasses writing to event log
		e.Apply()

		eventsRead++
	}

	if logLineScanner.Err() != nil {
		log.Fatalf("logLineScanner error: %s", logLineScanner.Err().Error())
	}

	log.Printf("ReadOldEvents(): read %d event(s)", eventsRead)

	return nil
}
