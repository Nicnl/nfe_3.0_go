package presized_zip

type BufferCounter struct {
	Size uint64
}

func (bc *BufferCounter) Write(p []byte) (n int, err error) {
	bc.Size += uint64(len(p))
	return len(p), nil
}
