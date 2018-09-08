package vfs

import (
	"os"
	"time"
)

type FakeFile struct {
	name   string
	length int64
	mode   os.FileMode
}

func (f *FakeFile) Name() string {
	// A bit of a cheat: we only have a basename, so that's also ok for FileInfo.
	return f.name
}

func (f *FakeFile) Size() int64 {
	return f.length
}

func (f *FakeFile) Mode() os.FileMode {
	return f.mode
}

func (f *FakeFile) ModTime() time.Time {
	return time.Time{}
}

func (f *FakeFile) IsDir() bool {
	return false
}

func (f *FakeFile) Sys() interface{} {
	return nil
}
