package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"strings"
)

var confSalt []byte = []byte("example_hash_salt")
var confUrlList []byte = []byte("files.nicnl.com")
var confUrlDown []byte = []byte("download.nicnl.com")

func unique(input []byte) string {
	var b bytes.Buffer

	b.Write(input)
	b.Write([]byte(fmt.Sprintf("%d", len(input))))

	b.Write(confSalt)
	b.Write([]byte(fmt.Sprintf("%d", len(confSalt))))

	b.Write(confUrlList)
	b.Write([]byte(fmt.Sprintf("%d", len(confUrlList))))

	b.Write(confUrlDown)
	b.Write([]byte(fmt.Sprintf("%d", len(confUrlDown))))

	fmt.Println(md5.Sum(input))

	return fmt.Sprintf("%x", md5.Sum(b.Bytes()))
}

func encode(path string) string {
	path = strings.TrimLeft(path, "/")
	if !strings.Contains(path, "/") {
		return unique([]byte(path))
	}

	fmt.Println("path =", path)

	return ""
}

func main() {
	fmt.Println("Return:", encode("/root_pathéèéè.png"))
}
