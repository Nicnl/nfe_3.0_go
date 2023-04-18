package presized_zip

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const concurrentChans = 64
const smallFileThreshold = 128 * 1024

func StreamZip(ctx context.Context, basePath string, w io.Writer, files []*zip.FileHeader) (outputErr error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "Recovered in StreamZip:", r)
			outputErr = fmt.Errorf("recovered in StreamZip: %v", r)
		}
	}()

	// Big buffer for fat file copy
	copyBuf := make([]byte, 8*1024*1024) // Buffer de 8Mo pour la copie

	// Async chan, reading small files in advance
	var (
		fullChans = make(chan chan []byte, concurrentChans)
		freeChans = make(chan chan []byte, concurrentChans)
	)

	for i := 0; i < concurrentChans; i++ {
		//fmt.Println("Creating chan", i)
		freeChans <- make(chan []byte)
	}
	//fmt.Println("Chans created")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "Recovered in StreamZip:", r)
			}
		}()

		for _, fh := range files {
			select {
			case <-ctx.Done():
				// Context was cancelled, close channels and return
				close(fullChans)
				close(freeChans)
				return
			default:
				if fh.UncompressedSize64 < smallFileThreshold {
					selectedChan := <-freeChans
					fullChans <- selectedChan

					go func(c chan []byte, curFile string) {
						//fmt.Println("Started goroutine, reading", curFile)
						rawData, err := os.ReadFile(curFile)
						if err != nil {
							panic(err)
						}

						c <- rawData
					}(selectedChan, filepath.Join(basePath, fh.Name))
				}
			}
		}
		close(fullChans)
	}()

	// Just in case
	defer func() {
		defer func() { recover() }()
		close(freeChans)
	}()

	defer func() {
		defer func() { recover() }()
		close(fullChans)
	}()

	// Actual zip writing
	zw := zip.NewWriter(w)

	err := zw.SetComment("") // No comment as it harder for deterministic size
	if err != nil {
		return err
	}

	for _, fh := range files {
		fw, err := zw.CreateHeader(fh)
		if err != nil {
			return err
		}

		// Write the file content
		if fh.UncompressedSize64 < smallFileThreshold {
			// Read small files from chan
			readChan := <-fullChans
			_, err := fw.Write(<-readChan)
			if err != nil {
				return fmt.Errorf("%s: %s / exp size: %d", fh.Name, err.Error(), fh.UncompressedSize64)
			}
			freeChans <- readChan
		} else {
			// Read big file from disk
			f, err := os.Open(filepath.Join(basePath, fh.Name))
			if err != nil {
				return fmt.Errorf("%s: %s", fh.Name, err.Error())
			}

			_, err = io.CopyBuffer(fw, f, copyBuf)
			if err != nil {
				return fmt.Errorf("%s: %s / exp size: %d", fh.Name, err.Error(), fh.UncompressedSize64)
			}
		}
	}

	err = zw.Close()
	if err != nil {
		return err
	}

	return nil
}
