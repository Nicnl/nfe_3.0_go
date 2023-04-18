package presized_zip

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestZip(t *testing.T) {
	dirPath := strings.TrimSpace(os.Getenv("TEST_ZIP_COMPRESS_DIR_PATH"))
	if dirPath == "" {
		dirPath = `C:\Users\Nicnl\Desktop\stable-diffusion-webui`
	}

	zipPath := dirPath + ".zip"

	expectedFileSize, files, err := PrepareZip(dirPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("expectedFileSize =", expectedFileSize)

	// open file truncate/Create:
	f, err := os.OpenFile(zipPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	err = StreamZip(context.Background(), dirPath, f, files)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	fileInfo, err := os.Stat(zipPath)
	if err != nil {
		panic(err)
	}

	fmt.Println("fileInfo.Size() =", fileInfo.Size())
}
