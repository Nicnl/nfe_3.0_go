package vfs

import "fmt"

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

func (v *Fake) FileSize(path string) (int64, error) {
	return 1024 * 1024, nil
}
