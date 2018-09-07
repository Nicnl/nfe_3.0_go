package main

import (
	"fmt"
	nfeCrypt "nfe_3.0_go/nfe/crypt"
	"os"
	"path/filepath"
	"strings"
)

func encode(path string) string {
	path = strings.TrimLeft(path, "/")
	if !strings.Contains(path, "/") {
		return nfeCrypt.GlobUnique([]byte(path))
	}

	fmt.Println("path =", path)

	return ""
}

func main() {

	err := filepath.Walk("C:/", func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})
	if err != nil {
		panic(err)
	}
}
