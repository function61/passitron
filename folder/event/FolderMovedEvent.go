package event

import (
	"github.com/function61/pi-security-module/state"
)

type FolderMoved struct {
	Id       string
	ParentId string
}

func (e *FolderMoved) Apply() {
	for idx, s := range state.Data.Folders {
		if s.Id == e.Id {
			s.ParentId = e.ParentId
			state.Data.Folders[idx] = s
			return
		}
	}
}
