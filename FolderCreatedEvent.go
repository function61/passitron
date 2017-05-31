package main

type FolderCreated struct {
	Id       string
	ParentId string
	Name     string
}

func (e *FolderCreated) Apply() {
	newFolder := Folder{
		Id:       e.Id,
		ParentId: e.ParentId,
		Name:     e.Name,
	}

	state.Folders = append(state.Folders, newFolder)
}
