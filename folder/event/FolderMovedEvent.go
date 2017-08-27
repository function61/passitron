package event

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type FolderMoved struct {
	eventbase.Event
	Id       string
	ParentId string
}

func (e FolderMoved) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "FolderMoved " + string(asJson)
}

func (e FolderMoved) Apply() {
	for idx, s := range state.Inst.State.Folders {
		if s.Id == e.Id {
			s.ParentId = e.ParentId
			state.Inst.State.Folders[idx] = s
			return
		}
	}
}
