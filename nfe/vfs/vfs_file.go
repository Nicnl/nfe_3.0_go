package vfs

import (
	"io/ioutil"
	"os"
	"strings"
)

type File struct {
	basePath string
}

func New(basePath string) Vfs {
	return &File{
		basePath: strings.TrimRight(basePath, "/"),
	}
}

func (v *File) Ls(path string) ([]string, error) {
	path = "/" + strings.TrimLeft(path, "/")

	files, err := ioutil.ReadDir(v.basePath + path)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, f := range files {
		out = append(out, f.Name())
	}

	return out, nil
}

func (v *File) Stat(path string) (os.FileInfo, error) {
	path = "/" + strings.TrimLeft(path, "/")

	stat, err := os.Stat(v.basePath + path)
	if err != nil {
		return nil, err
	}

	return stat, nil
}
