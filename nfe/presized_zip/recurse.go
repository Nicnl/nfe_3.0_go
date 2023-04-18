package presized_zip

import (
	"archive/zip"
	"os"
	"path/filepath"
)

var writeBuf = make([]byte, 10*1024*1024)

func _recurseDir(path string, subPath string, allFiles *[]*zip.FileHeader) error {
	entries, err := os.ReadDir(filepath.Join(path, subPath))
	if err != nil {
		panic(err)
	}

	if len(entries) == 0 {
		// Skip empty directories
		// Todo: handle empty directories?
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
				fileMode         = fileInfo.Mode()
			)

			// Check if filemode is link
			if fileMode&os.ModeSymlink != 0 || fileMode&os.ModeNamedPipe != 0 || fileMode&os.ModeSocket != 0 || fileMode&os.ModeDevice != 0 {
				continue
			}

			// Append the zip header to the list
			fh, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				return err
			}

			fh.Name = fileRelativePath
			fh.Method = zip.Store // Size must be deterministic for presized zip
			fh.Comment = ""       // Comment is not used, determinism is more important

			_allFiles = append(_allFiles, fh)
		}
	}
	*allFiles = _allFiles

	// Recurse into directories
	for _, dir := range dirs {
		err = _recurseDir(path, filepath.Join(subPath, dir.Name()), allFiles)
		if err != nil {
			return err
		}
	}

	return nil
}
