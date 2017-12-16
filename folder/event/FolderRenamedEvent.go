package event

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type FolderRenamed struct {
	eventbase.Event
	Id   string
	Name string
}

func (e FolderRenamed) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "FolderRenamed " + string(asJson)
}

func FolderRenamedFromSerialized(payload []byte) *FolderRenamed {
	var e FolderRenamed
	if err := json.Unmarshal(payload, &e); err != nil {
		panic(err)
	}
	return &e
}

func (e FolderRenamed) Apply() {
	for idx, s := range state.Inst.State.Folders {
		if s.Id == e.Id {
			s.Name = e.Name
			state.Inst.State.Folders[idx] = s
			return
		}
	}
}
