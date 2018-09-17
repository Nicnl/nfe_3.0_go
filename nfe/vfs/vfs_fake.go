package vfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
		length: 12,
		mode:   0777,
	}, nil
}

func (v *Fake) Open(path string) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader([]byte("Hello World!"))), nil
}

func (v *Fake) OpenSeek(path string, atSeek int64) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader([]byte("Hello World!"))), nil
}

func (v *Fake) SubVfs(path string) Vfs {
	return v
}
