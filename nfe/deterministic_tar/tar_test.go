package deterministic_tar

import (
	"fmt"
	"os"
	"testing"
)

func TestTar(t *testing.T) {
	path := "/home/nlim/Projects/aivianet_3.0/aivianet-vue"

	expectedSize, files, err := A_PrecalculateTarSize(path)
	if err != nil {
		panic(err)
	}
	fmt.Println("nbFiles =", len(files))
	fmt.Println("expectedSize", expectedSize)

	bc := BufferCounter{}
	err = B_tar_stream(&bc, files, expectedSize)
	if err != nil {
		panic(err)
	}

	fmt.Println("actualSize", bc.Size)
	fmt.Println("diff", bc.Size-expectedSize)

	if bc.Size-expectedSize == 0 {
		f, err := os.Create("test.tar")
		if err != nil {
			panic(err)
		}

		err = B_tar_stream(f, files, expectedSize)
		if err != nil {
			panic(err)
		}
	}
}
