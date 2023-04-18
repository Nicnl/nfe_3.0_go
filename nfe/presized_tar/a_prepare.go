package presized_tar

import (
	"archive/tar"
	"nfe_3.0_go/helpers/buffer_counter"
	"time"
)

type File struct {
	RelativePath string
	AbsolutePath string
	Size         int64
	ModTime      time.Time
	Mode         int64
}

func Prepare(path string) (expectedSize int64, files []File, err error) {
	bc := buffer_counter.BufferCounter{}
	tw := tar.NewWriter(&bc)

	files = make([]File, 0)
	err = _recurseDir(&bc, tw, path, "", &files)
	if err != nil {
		return -1, nil, err
	}

	err = tw.Close()
	if err != nil {
		return -1, nil, err
	}

	expectedSize = bc.Size
	return
}
