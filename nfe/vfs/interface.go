package vfs

import (
	"io"
	"os"
)

// Virtual File System
// Interface permettant d'accéder facilement aux fichiers système
type Vfs interface {
	Ls(path string) ([]string, error)
	Stat(path string) (os.FileInfo, error)
	Open(path string) (io.ReadCloser, error)
	OpenSeek(path string, atSeek int64) (io.ReadCloser, error)
	SubVfs(path string) Vfs
	AbsolutePath(path string) string
}
