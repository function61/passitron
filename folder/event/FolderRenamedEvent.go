package event

import (
	"github.com/function61/pi-security-module/state"
)

type FolderRenamed struct {
	Id   string
	Name string
}

func (e *FolderRenamed) Apply() {
	for idx, s := range state.Data.Folders {
		if s.Id == e.Id {
			s.Name = e.Name
			state.Data.Folders[idx] = s
			return
		}
	}
}
