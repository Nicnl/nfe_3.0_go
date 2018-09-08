package vfs

import (
	"fmt"
	"os"
	"path/filepath"
)

type Fake struct {
	Structure map[string][]string
}

func (v *Fake) Ls(path string) ([]string, error) {
	content, ok := v.Structure[path]
	if !ok {
		return nil, fmt.Errorf("path not registered in fake vfs")
	}

	return content, nil
}

func (v *Fake) Stat(path string) (os.FileInfo, error) {
	dir, filename := filepath.Split(path)

	content, err := v.Ls(path)
	if err != nil {
		return nil, err
	}

	found := false
	var foundEntry string
	for _, entry := range content {
		if entry == filename {
			found = true
			foundEntry = entry
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("file '%s' not present in the fake vfs path '%s'", filename, dir)
	}

	return &FakeFile{
		name:   foundEntry,
		length: 20 * 1024 * 1024,
		mode:   0777,
	}, nil
}
