package main

import (
	"fmt"
	nfeCrypt "nfe_3.0_go/nfe/crypt"
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
	fmt.Println("Return:", encode("/root_pathéèéè.png"))
}
