package main

type FolderMoved struct {
	Id       string
	ParentId string
}

func (e *FolderMoved) Apply() {
	for idx, s := range state.Folders {
		if s.Id == e.Id {
			s.ParentId = e.ParentId
			state.Folders[idx] = s
			return
		}
	}
}
