package deterministic_tar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
)

func B_tar_stream(w io.Writer, files []File, expectedSize int64) error {
	copyBuf := make([]byte, 2*1024*1024)
	tw := tar.NewWriter(w)

	for _, file := range files {
		// Write the tar header
		err := tw.WriteHeader(&tar.Header{
			Name:    file.RelativePath,
			Size:    file.Size, // Set to 0 so that the buf only contains the headers
			Mode:    file.Mode,
			ModTime: file.ModTime,
		})
		if err != nil {
			return err
		}

		// Write the file content
		f, err := os.Open(file.AbsolutePath)
		if err != nil {
			return fmt.Errorf("%s: %s", file.AbsolutePath, err.Error())
		}

		written, err := io.CopyBuffer(tw, f, copyBuf)
		if err != nil {
			return fmt.Errorf("%s: %s / exp size: %d", file.AbsolutePath, err.Error(), file.Size)
		}
		if written != file.Size {
			return fmt.Errorf("written: %d, should have wrote: %d", written, file.Size)
		} else {
			//fmt.Printf("written: %d\n", written)
		}
	}

	err := tw.Close()
	if err != nil {
		return err
	}

	return nil
}
