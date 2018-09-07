package crypt

import (
	"fmt"
	"nfe_3.0_go/nfe/vfs"
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
	//fmt.Println("since+duration =", since+duration)
	sinceStr := strconv.FormatInt(since+duration, 16)

	//decodedPath := pathEncodeRaw(path)
	//fmt.Println("decodedPath =", decodedPath)

	var b strings.Builder
	encodedPath := HexEncode(pathEncodeRaw(path), GlobUnique([]byte(fmt.Sprintf("%d", duration+since))))
	//fmt.Println("encodedPath key =", GlobUnique([]byte(fmt.Sprintf("%d", duration+since))))

	b.WriteString(encodedPath)
	b.WriteByte('-')
	b.WriteString(HexEncode(sinceStr, GlobUnique([]byte(encodedPath))))
	//fmt.Println("sinceStr =", sinceStr)
	//fmt.Println("sinceStr enc =", HexEncode(sinceStr, GlobUnique([]byte(encodedPath))))
	//fmt.Println("sinceStr key =", GlobUnique([]byte(encodedPath)))

	hashLength := 25

	// Une durée de 0 signifie qu'on génère un chemin expirable identique au PHP
	if duration == 0 {
		hashLength = 2
	}

	b.WriteString(GlobUnique([]byte(b.String()))[:hashLength])

	//fmt.Println("------------------------")
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

func subFind(currentPath string, searched string, v vfs.Vfs) (string, error) {
	//fmt.Println("v.Ls =>", currentPath)
	content, err := v.Ls(currentPath)
	if err != nil {
		return "", err
	}

	currentSearched := searched[:32]
	//fmt.Println("currentSearched =", currentSearched)

	for _, entry := range content {
		//fmt.Println("  -", entry, "  =>  ", GlobUnique([]byte(entry)))
		if GlobUnique([]byte(entry)) == currentSearched {
			if !strings.HasSuffix(currentPath, "/") {
				currentPath += "/"
			}
			currentPath += entry

			searched = searched[32:]
			if searched == "" {
				// Plus rien à chercher, fin de la récursion
				return currentPath, nil
			} else {
				return subFind(currentPath, searched, v)
			}
		}
	}

	return "", fmt.Errorf("no entry matching for hash '%s' in path '%s", searched, currentPath)
}

func Find(path string, v vfs.Vfs) (string, error) {
	if !CheckHash(path) {
		return "", fmt.Errorf("the checksum is invalid for the following path '%s'", path)
	}

	decodedPath := ""
	filenameHash := ""
	if pos := strings.IndexByte(path, '-'); pos != -1 {
		// Si c'est un chemin expirable
		// encodedPath := HexEncode(pathEncodeRaw(path), GlobUnique([]byte(fmt.Sprintf("%d", duration+since))))

		expiration := path[pos+1:]
		//expiration := path[pos+1:]

		encodedPath := path[:pos]

		//fmt.Println("path =", path)
		//fmt.Println("path =", encodedPath, expiration)

		sinceStrKey := GlobUnique([]byte(encodedPath))
		//fmt.Println("sinceStrKey =", sinceStrKey)

		sinceStrEncoded := expiration[:len(expiration)-25]
		//fmt.Println("sinceStrEncoded =", sinceStrEncoded)

		sinceStr := HexDecode(sinceStrEncoded, sinceStrKey)
		//fmt.Println("sinceStr =", sinceStr)

		since, err := strconv.ParseInt(sinceStr, 16, 64)
		if err != nil {
			return "", err
		}
		//fmt.Println("since =", since)

		encodedPathKey := GlobUnique([]byte(fmt.Sprintf("%d", since)))
		//fmt.Println("encodedPathKey =", encodedPathKey)

		decodedPath = HexDecode(encodedPath, encodedPathKey)
		//fmt.Println("decodedPath", decodedPath)
	} else {
		decodedPath = path[:len(path)-2]
	}

	if len(decodedPath) < 32 {
		return "", fmt.Errorf("the provided path is too short")
	}

	filenameHash = decodedPath[len(decodedPath)-32:]
	//fmt.Println("filenameHash", filenameHash)

	decodedPath = HexDecode(decodedPath[:len(decodedPath)-32], filenameHash)
	//fmt.Println("decodedPath", decodedPath)

	return subFind("/", decodedPath+filenameHash, v)
}
