package vfs

import "os"

type Vfs interface {
	Ls(path string) ([]string, error)
	Stat(path string) (os.FileInfo, error)
}
