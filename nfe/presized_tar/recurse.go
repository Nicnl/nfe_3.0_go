package presized_tar

import (
	"archive/tar"
	"nfe_3.0_go/helpers/buffer_counter"
	"os"
	"path/filepath"
)

func _recurseDir(bc *buffer_counter.BufferCounter, tw *tar.Writer, path string, subPath string, allFiles *[]File) error {
	entries, err := os.ReadDir(filepath.Join(path, subPath))
	if err != nil {
		panic(err)
	}

	if len(entries) == 0 {
		return nil
	}

	// Extract files, and directories
	dirs := make([]os.DirEntry, 0, len(entries))

	// Todo: sort before appending, so that it is deterministic
	_allFiles := *allFiles
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			fileInfo, err := entry.Info()
			if err != nil {
				return err
			}

			var (
				fileRelativePath = filepath.Join(subPath, entry.Name())
				fileSize         = fileInfo.Size()
				fileMode         = fileInfo.Mode()
				fileTime         = fileInfo.ModTime()
			)

			// Check if filemode is link
			if fileMode&os.ModeSymlink != 0 || fileMode&os.ModeNamedPipe != 0 || fileMode&os.ModeSocket != 0 || fileMode&os.ModeDevice != 0 {
				continue
			}

			// Append file to the list
			_allFiles = append(_allFiles, File{
				RelativePath: fileRelativePath,
				AbsolutePath: filepath.Join(path, fileRelativePath),
				Size:         fileSize,
				ModTime:      fileTime,
				Mode:         int64(fileMode),
			})

			// Write tar header
			{
				err = tw.WriteHeader(&tar.Header{
					Name:    fileRelativePath,
					Size:    0,
					Mode:    int64(fileMode),
					ModTime: fileTime,
				})
				if err != nil {
					return err
				}
			}

			// Write fake file
			{
				bc.Size += int64(fileSize)

				padding := 512 - (fileSize % 512)
				if padding != 512 {
					bc.Size += padding
				}
			}
		}
	}
	*allFiles = _allFiles

	// Recurse into directories
	for _, dir := range dirs {
		err := _recurseDir(bc, tw, path, filepath.Join(subPath, dir.Name()), allFiles)
		if err != nil {
			return err
		}
	}

	return nil
}
