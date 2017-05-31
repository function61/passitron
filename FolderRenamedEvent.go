package main

type FolderRenamed struct {
	Id   string
	Name string
}

func (e *FolderRenamed) Apply() {
	for idx, s := range state.Folders {
		if s.Id == e.Id {
			s.Name = e.Name
			state.Folders[idx] = s
			return
		}
	}
}
