package presized_tar

import (
	"fmt"
	"nfe_3.0_go/helpers/buffer_counter"
	"testing"
)

func TestTar(t *testing.T) {
	path := "C:\\Users\\Nicnl\\Desktop\\stable-diffusion-webui\\outputs"

	expectedSize, files, err := Prepare(path)
	if err != nil {
		panic(err)
	}
	fmt.Println("nbFiles =", len(files))
	fmt.Println("expectedSize", expectedSize)

	bc := buffer_counter.BufferCounter{}
	err = Stream(&bc, files)
	if err != nil {
		panic(err)
	}

	fmt.Println("actualSize", bc.Size)
	fmt.Println("diff", bc.Size-expectedSize)
}
