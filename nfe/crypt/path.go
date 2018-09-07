package crypt

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func pathEncodeRaw(path string) string {
	path = strings.TrimLeft(path, "/")
	if !strings.Contains(path, "/") {
		return GlobUnique([]byte(path))
	}

	dir, filename := filepath.Split(path)
	dir = strings.TrimRight(dir, "/")

	filenameHash := GlobUnique([]byte(filename))

	var b strings.Builder
	for _, subdir := range strings.Split(dir, "/") {
		b.WriteString(GlobUnique([]byte(subdir)))
	}

	return HexEncode(b.String(), filenameHash) + filenameHash
}

func PathEncode(path string) string {
	var b strings.Builder

	b.WriteString(pathEncodeRaw(path))
	b.WriteString(GlobUnique([]byte(b.String()))[:2])

	return b.String()
}

func PathEncodeExpirable(path string, duration int64, since int64) string {
	sinceStr := strconv.FormatInt(since+duration, 16)

	var b strings.Builder
	encodedPath := HexEncode(pathEncodeRaw(path), GlobUnique([]byte(fmt.Sprintf("%d", since))))

	b.WriteString(encodedPath)
	b.WriteByte('-')
	b.WriteString(HexEncode(sinceStr, GlobUnique([]byte(encodedPath))))

	if duration == 0 {
		// Une durée de 0 signifie qu'on génère un chemin expirable identique au PHP
		b.WriteString(GlobUnique([]byte(b.String()))[:2])
		return b.String()
	} else {
		// Ou sinon on adopte le nouveau comportement plus safe
		pathPlusDuration := b.String()
		checksum := GlobUnique([]byte(b.String()))[:25]

		b.Reset()
		reencodedPath := HexEncode(pathPlusDuration, checksum)
		b.WriteString(reencodedPath[:len(encodedPath)])
		b.WriteByte('-')
		b.WriteString(reencodedPath[len(encodedPath)+1:])
		b.WriteString(checksum)
		return b.String()
	}

}
