package crypt

import (
	"bytes"
	"crypto/md5"
	"fmt"
)

var GlobSalt []byte
var GlobUrlList []byte
var GlobUrlDown []byte

func GlobUnique(input []byte) string {
	return Unique(input, GlobSalt, GlobUrlList, GlobUrlDown)
}

func Unique(input []byte, salt []byte, urlList []byte, urlDown []byte) string {
	var b bytes.Buffer

	b.Write(input)
	b.Write([]byte(fmt.Sprintf("%d", len(input))))

	b.Write(salt)
	b.Write([]byte(fmt.Sprintf("%d", len(salt))))

	b.Write(urlList)
	b.Write([]byte(fmt.Sprintf("%d", len(urlList))))

	b.Write(urlDown)
	b.Write([]byte(fmt.Sprintf("%d", len(urlDown))))

	return fmt.Sprintf("%x", md5.Sum(b.Bytes()))
}
