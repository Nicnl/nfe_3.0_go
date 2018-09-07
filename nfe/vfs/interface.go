package vfs

type Vfs interface {
	Ls(path string) ([]string, error)
	FileSize(path string) (int64, error)
}
