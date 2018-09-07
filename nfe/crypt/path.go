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
	encodedPath := HexEncode(pathEncodeRaw(path), GlobUnique([]byte(fmt.Sprintf("%d", duration+since))))

	b.WriteString(encodedPath)
	b.WriteByte('-')
	b.WriteString(HexEncode(sinceStr, GlobUnique([]byte(encodedPath))))

	hashLength := 25

	// Une durée de 0 signifie qu'on génère un chemin expirable identique au PHP
	if duration == 0 {
		hashLength = 2
	}

	b.WriteString(GlobUnique([]byte(b.String()))[:hashLength])
	return b.String()
}

func CheckHash(input string) bool {
	hashLength := 2

	// En mode "nouveau comportement" on s'en fous des anciens hash de taille 2
	if strings.Contains(input, "-") {
		hashLength = 25
	}

	if hashLength > len(input) {
		return false
	}

	return GlobUnique([]byte(input[:len(input)-hashLength]))[:hashLength] == input[len(input)-hashLength:]
}
