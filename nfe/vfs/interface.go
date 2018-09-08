package vfs

import "os"

// Virtual File System
// Interface permettant d'accéder facilement aux fichiers système
type Vfs interface {
	Ls(path string) ([]string, error)
	Stat(path string) (os.FileInfo, error)
}
