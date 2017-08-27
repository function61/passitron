package event

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type FolderCreated struct {
	eventbase.Event
	Id       string
	ParentId string
	Name     string
}

func (e FolderCreated) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "FolderCreated " + string(asJson)
}

func (e FolderCreated) Apply() {
	newFolder := state.Folder{
		Id:       e.Id,
		ParentId: e.ParentId,
		Name:     e.Name,
	}

	state.Inst.State.Folders = append(state.Inst.State.Folders, newFolder)
}
