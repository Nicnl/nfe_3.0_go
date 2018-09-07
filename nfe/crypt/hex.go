package crypt

import (
	"fmt"
	"strconv"
	"strings"
)

func HexEncode(hex string, key string) string {
	var b strings.Builder

	if key == "" {
		key = "?"
	} // So that ParseInt fails and we use 0 instead

	keyIndex := 0
	for _, chr := range hex {
		keyByte, err := strconv.ParseInt(string(key[keyIndex]), 16, 8)
		if err != nil {
			keyByte = 0
		}

		chrByte, err := strconv.ParseInt(string(chr), 16, 8)
		if err != nil {
			chrByte = 0
		}

		enc := (chrByte + keyByte) % 16
		if enc < 0 {
			enc += 16
		}

		b.WriteString(fmt.Sprintf("%x", enc))

		if keyIndex++; keyIndex >= len(key) {
			keyIndex = 0
		}
	}

	return b.String()
}
