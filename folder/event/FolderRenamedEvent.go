package event

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type FolderRenamed struct {
	eventbase.Event
	Id   string
	Name string
}

func (e *FolderRenamed) Apply() {
	for idx, s := range state.Inst.State.Folders {
		if s.Id == e.Id {
			s.Name = e.Name
			state.Inst.State.Folders[idx] = s
			return
		}
	}
}
