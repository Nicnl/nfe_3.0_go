package presized_zip

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"nfe_3.0_go/helpers"
	"os"
	"path/filepath"
	"runtime"
)

const concurrentChans = 64
const smallFileThreshold = 128 * 1024

func StreamZip(ctx context.Context, basePath string, w io.Writer, files []*zip.FileHeader) (outputErr error) {
	if os.Getenv("PRINT_MEM_AFTER_ARCHIVE") == "1" {
		defer func() {
			// debug print memory stack heap gc

			bToMb := func(b uint64) uint64 {
				return b / 1024 / 1024
			}

			fmt.Println("StreamZip: Memory stats:")
			m := new(runtime.MemStats)
			runtime.ReadMemStats(m)
			fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
			fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
			fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
			fmt.Printf("\tNumGC = %v", m.NumGC)
			fmt.Println()
		}()
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "Recovered in StreamZip:", r)
			outputErr = fmt.Errorf("recovered in StreamZip: %v", r)
		}
	}()

	// Big buffer for fat file copy
	copyBuf := make([]byte, 8*1024*1024) // Buffer de 8Mo pour la copie
	defer func() {
		defer helpers.RecoverStderr()
		copyBuf = nil
	}()

	// Async chan, reading small files in advance
	var (
		fullChans = make(chan chan []byte, concurrentChans)
		freeChans = make(chan chan []byte, concurrentChans)

		fullChansNeedsClosing = true
		freeChansNeedsClosing = true
	)

	closeFullChans := func() {
		if fullChansNeedsClosing {
			defer helpers.RecoverStderr()
			close(fullChans)
			fullChansNeedsClosing = false
		}
	}

	closeFreeChans := func() {
		if freeChansNeedsClosing {
			defer helpers.RecoverStderr()
			close(freeChans)
			freeChansNeedsClosing = false
		}
	}

	defer closeFullChans()
	defer closeFreeChans()

	for i := 0; i < concurrentChans; i++ {
		//fmt.Println("Creating chan", i)
		c := make(chan []byte, 1)
		defer func() { close(c) }()
		freeChans <- c
	}
	//fmt.Println("Chans created")

	go func() {
		defer helpers.RecoverStderr()

		for _, fh := range files {
			select {
			case <-ctx.Done():
				// Context was cancelled, close channels and return
				closeFullChans()
				closeFreeChans()
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

						select {
						case c <- rawData:
							// Write to channel successfully
						default:
							// Channel is closed, do nothing
						}
					}(selectedChan, filepath.Join(basePath, fh.Name))
				}
			}
		}
		closeFullChans()
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
			rawData := <-readChan
			if uint64(len(rawData)) != fh.UncompressedSize64 {
				return fmt.Errorf("%s: read %d bytes, expected %d", fh.Name, len(rawData), fh.UncompressedSize64)
			}

			_, err := fw.Write(rawData)
			if err != nil {
				return fmt.Errorf("%s: %s / exp size: %d", fh.Name, err.Error(), fh.UncompressedSize64)
			}
			freeChans <- readChan
		} else {
			err = func() error {
				// Read big file from disk
				f, err := os.Open(filepath.Join(basePath, fh.Name))
				if err != nil {
					return fmt.Errorf("%s: %s", fh.Name, err.Error())
				}
				defer f.Close()

				wrote, err := io.CopyBuffer(fw, f, copyBuf)
				if err != nil {
					return fmt.Errorf("%s: %s / exp size: %d", fh.Name, err.Error(), fh.UncompressedSize64)
				}

				if uint64(wrote) != fh.UncompressedSize64 {
					return fmt.Errorf("%s: wrote %d bytes, expected %d", fh.Name, wrote, fh.UncompressedSize64)
				}

				return nil
			}()
			if err != nil {
				return fmt.Errorf("%s: %s", fh.Name, err.Error())
			}
		}
	}

	err = zw.Close()
	if err != nil {
		return err
	}

	return nil
}
