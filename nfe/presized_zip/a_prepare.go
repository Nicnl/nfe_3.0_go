package presized_zip

import "archive/zip"

const (
	uint32max = (1 << 32) - 1
	uint16max = (1 << 16) - 1
)

func PrepareZip(basePath string) (uint64, []*zip.FileHeader, error) {
	files := make([]*zip.FileHeader, 0)
	err := _recurseDir(basePath, "", &files)
	if err != nil {
		return 0, nil, err
	}

	// Calculate the size of the file headers and their offsets in the zip file
	var (
		fileHeaderSize = uint64(0)
		offsets        = make([]uint64, len(files))
	)
	for i, fh := range files {
		offsets[i] = fileHeaderSize

		// Size: local file header
		{
			fileHeaderSize += 30                    // base header size
			fileHeaderSize += uint64(len(fh.Name))  // filename length
			fileHeaderSize += uint64(len(fh.Extra)) // extra length

			// The go zip library will append an "extended timestamp" extra field record
			if !fh.Modified.IsZero() {
				fileHeaderSize += 9
			}
		}

		// Size: data
		{
			fileHeaderSize += uint64(fh.UncompressedSize64)
		}

		// Size: data descriptor
		isZip64 := fh.UncompressedSize64 >= uint32max
		{
			if isZip64 {
				fileHeaderSize += 4 + 4 + 8 + 8 // zip64 data descriptor size
			} else {
				fileHeaderSize += 4 + 4 + 4 + 4 // data descriptor size
			}
		}
	}

	//fmt.Println("fileHeaderSize =", fileHeaderSize)

	// Predicate the central record size
	centralRecordSize := uint64(0)
	{
		for i, fh := range files {
			centralRecordSize += 46                      // central directory entry size
			centralRecordSize += uint64(len(fh.Name))    // filename length
			centralRecordSize += uint64(len(fh.Extra))   // extra length
			centralRecordSize += uint64(len(fh.Comment)) // comment length

			// The go zip library will append an "extended timestamp" extra field record
			if !fh.Modified.IsZero() {
				centralRecordSize += 9
			}

			// Extra field record for zip64 files
			if fh.UncompressedSize64 >= uint32max || offsets[i] >= uint32max {
				centralRecordSize += 28
			}
		}

		centralRecordSize += 22 // end of central directory record size
		centralRecordSize += 0  // No comment

	}

	// Predicate if a zip64 end of central directory record is needed
	{
		if len(files) >= uint16max || fileHeaderSize >= uint32max || centralRecordSize >= uint32max {
			centralRecordSize += 56 + 20 // go std: directory64EndLen + directory64LocLen
		}
	}

	// Calculate the final total size of the zip file
	totalSize := uint64(0)
	{
		totalSize += fileHeaderSize
		totalSize += centralRecordSize
	}

	return totalSize, files, nil
}
