package vfs

import "os"

type File struct {
	BasePath string
}

func (v *File) Ls(path string) ([]string, []string) {
	return []string{}, []string{} // Todo
}

func (v *File) Stat(path string) (os.FileInfo, error) {
	return nil, nil // Todo
}
