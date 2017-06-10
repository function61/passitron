package event

import (
	"github.com/function61/pi-security-module/state"
)

type FolderCreated struct {
	Id       string
	ParentId string
	Name     string
}

func (e *FolderCreated) Apply() {
	newFolder := state.Folder{
		Id:       e.Id,
		ParentId: e.ParentId,
		Name:     e.Name,
	}

	state.Inst.State.Folders = append(state.Inst.State.Folders, newFolder)
}
