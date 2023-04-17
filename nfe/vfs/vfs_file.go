package vfs

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

	patternsMatch := os.Getenv("FILTER_OUT")
	if patternsMatch == "" {
		for _, f := range files {
			out = append(out, f.Name())
		}
	} else {
		patternList := strings.Split(patternsMatch, ";")

		for _, f := range files {
			keep := true
			for _, pattern := range patternList {
				match, err := filepath.Match(pattern, f.Name())
				if err != nil || match {
					keep = false
					break
				}
			}

			if keep {
				out = append(out, f.Name())
			}
		}
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

func (v *File) Open(path string) (io.ReadCloser, error) {
	path = "/" + strings.TrimLeft(path, "/")

	f, err := os.Open(v.basePath + path)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (v *File) OpenSeek(path string, atSeek int64) (io.ReadCloser, error) {
	path = "/" + strings.TrimLeft(path, "/")

	f, err := os.Open(v.basePath + path)
	if err != nil {
		return nil, err
	}

	_, err = f.Seek(atSeek, 0)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (v *File) AbsolutePath(path string) string {
	return filepath.Join(v.basePath, path)
}

func (v *File) SubVfs(path string) Vfs {
	if path == "" {
		return v
	}

	path = "/" + strings.TrimLeft(path, "/")

	return New(v.basePath + path)
}
