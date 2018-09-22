package crypt

import (
	"fmt"
	"math/rand"
	"nfe_3.0_go/nfe/vfs"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

func PathEncodeExpirable(path string, duration time.Duration, since time.Time) string {
	limitTimestamp := since.Add(duration).Unix()

	//fmt.Println("since+duration =", since+duration)
	sinceStr := strconv.FormatInt(limitTimestamp, 16)

	//decodedPath := pathEncodeRaw(path)
	//fmt.Println("decodedPath =", decodedPath)

	var b strings.Builder
	encodedPath := HexEncode(pathEncodeRaw(path), GlobUnique([]byte(fmt.Sprintf("%d", limitTimestamp))))
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

	return "", fmt.Errorf("no entry matching for hash '%s' in path '%s'", searched, currentPath)
}

func FindTimeLimitIgnorable(path string, timeLimit time.Time, v vfs.Vfs, ignoreTimeLimit bool) (string, int64, error) {
	pathWithoutBandwidth, bandwidthLimit, err := GetBandwidthLimit(path)
	if err != nil {
		return "", -1, err
	}
	path = pathWithoutBandwidth

	if !CheckHash(path) {
		return "", -1, fmt.Errorf("the checksum is invalid for the following path '%s'", path)
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
			return "", -1, err
		}
		//fmt.Println("since =", since)

		if !ignoreTimeLimit && timeLimit.Unix() > since {
			return "", -1, fmt.Errorf("time limit is reached, path valid until '%d', given time limit is '%d', diff is '%d'", since, timeLimit.Unix(), timeLimit.Unix()-since)
			return "", -1, fmt.Errorf("link expired")
		}

		encodedPathKey := GlobUnique([]byte(fmt.Sprintf("%d", since)))
		//fmt.Println("encodedPathKey =", encodedPathKey)

		decodedPath = HexDecode(encodedPath, encodedPathKey)
		//fmt.Println("decodedPath", decodedPath)
	} else {
		decodedPath = path[:len(path)-2]
	}

	if len(decodedPath) < 32 {
		return "", -1, fmt.Errorf("the provided path is too short")
	}

	filenameHash = decodedPath[len(decodedPath)-32:]
	//fmt.Println("filenameHash", filenameHash)

	decodedPath = HexDecode(decodedPath[:len(decodedPath)-32], filenameHash)
	//fmt.Println("decodedPath", decodedPath)

	subFindStr, err := subFind("/", decodedPath+filenameHash, v)
	if err != nil {
		return "", -1, err
	}

	return subFindStr, bandwidthLimit, nil
}

func Find(path string, timeLimit time.Time, v vfs.Vfs) (string, int64, error) {
	return FindTimeLimitIgnorable(path, timeLimit, v, false)
}

var bandwidthLimitSeparator = "ghijklmnopqrstuvwxyz" // Tout sauf abcdef

func AddBandwidthLimit(link string, limit int64) string {
	timelimitSeparatorIndex := strings.Index(link, "-")

	linkGlob := GlobUnique([]byte(link))
	//fmt.Println("link =", link)
	//fmt.Println("linkGlob =", linkGlob)
	//fmt.Println("linkGlob =", linkGlob)

	limitStr := HexEncode(strconv.FormatInt(limit, 16), linkGlob)
	limitGlob := GlobUnique([]byte(limitStr))

	link = HexEncode(link, limitGlob)

	encodedLink := link + string(bandwidthLimitSeparator[rand.Intn(len(bandwidthLimitSeparator))]) + limitGlob + limitStr

	if timelimitSeparatorIndex >= 0 {
		encodedLink = encodedLink[:timelimitSeparatorIndex] + "-" + encodedLink[timelimitSeparatorIndex+1:]
	}

	//fmt.Println("___________________________")
	return encodedLink
}

func GetBandwidthLimit(link string) (string, int64, error) {
	separatorIndex := strings.IndexAny(link, bandwidthLimitSeparator)
	if separatorIndex == -1 {
		return link, 0, nil // No bandwidth limit whatsoever, input link is unchanged
	}

	timelimitSeparatorIndex := strings.Index(link, "-")
	if timelimitSeparatorIndex >= 0 {
		link = link[:timelimitSeparatorIndex] + "-" + link[timelimitSeparatorIndex+1:]
	}

	if separatorIndex+1+32 > len(link) {
		return "", -1, fmt.Errorf("the provided link does contains a bandwidth limit separator character, but it's length is too short")
	}

	linkPart := link[:separatorIndex]
	limitGlobExpected := link[separatorIndex+1 : separatorIndex+1+32]
	limitStr := link[separatorIndex+1+32:]

	//fmt.Println("linkPart =", linkPart)
	//fmt.Println("limitGlobExpected =", limitGlobExpected)
	//fmt.Println("limitStr =", limitStr)
	//fmt.Println()

	limitGlob := GlobUnique([]byte(limitStr))
	if limitGlob != limitGlobExpected {
		return "", -1, fmt.Errorf("the expected limitGlob is not equal to the given limitGlob")
	}
	//fmt.Println("limitGlob =", limitGlob)

	link = HexDecode(linkPart, limitGlob)
	if timelimitSeparatorIndex >= 0 {
		link = link[:timelimitSeparatorIndex] + "-" + link[timelimitSeparatorIndex+1:]
	}

	//fmt.Println("link =", link)
	linkGlob := GlobUnique([]byte(link))
	//fmt.Println("link =", link)
	//fmt.Println("linkGlob =", linkGlob)
	//fmt.Println()

	limitStr = HexDecode(limitStr, linkGlob)
	limit, err := strconv.ParseInt(limitStr, 16, 64)
	if err != nil {
		return "", -1, err
	}

	//fmt.Println("limit =", limit)

	return link, limit, nil
}
