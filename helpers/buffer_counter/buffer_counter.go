package buffer_counter

type BufferCounter struct {
	Size int64
}

func (bc *BufferCounter) Write(p []byte) (n int, err error) {
	bc.Size += int64(len(p))
	return len(p), nil
}
