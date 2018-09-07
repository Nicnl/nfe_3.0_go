package vfs

type File struct {
	BasePath string
}

func (v File) Ls(path string) ([]string, []string) {
	return []string{}, []string{}
}

func (v File) FileSize(path string) float64 {
	return 0
}
