package event

import (
	"github.com/function61/pi-security-module/state"
)

type FolderRenamed struct {
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
