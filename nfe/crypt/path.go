package crypt

import (
	"fmt"
	"strings"
)

func pathEncodeRaw(path string) string {
	path = strings.TrimLeft(path, "/")
	if !strings.Contains(path, "/") {
		return GlobUnique([]byte(path))
	}

	return ""
}

func PathEncode(path string) string {
	var b strings.Builder

	encoded := pathEncodeRaw(path)
	fmt.Println(fmt.Sprintf("%x", GlobUnique([]byte(encoded))[:1]))

	b.WriteString(GlobUnique([]byte(encoded))[:2])

	return b.String()
}
