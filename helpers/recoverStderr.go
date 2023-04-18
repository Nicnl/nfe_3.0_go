package helpers

import (
	"fmt"
	"os"
)

func RecoverStderr() {
	if err := recover(); err != nil {
		fmt.Fprintf(os.Stderr, "[dfb36fb0] recover d'un panic\n%s\n", err)
	}
}
