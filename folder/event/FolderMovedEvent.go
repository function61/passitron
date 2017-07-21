package event

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type FolderMoved struct {
	eventbase.Event
	Id       string
	ParentId string
}

func (e *FolderMoved) Apply() {
	for idx, s := range state.Inst.State.Folders {
		if s.Id == e.Id {
			s.ParentId = e.ParentId
			state.Inst.State.Folders[idx] = s
			return
		}
	}
}
