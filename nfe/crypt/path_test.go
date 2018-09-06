package crypt

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	GlobSalt = []byte("example_hash_salt")
	GlobUrlList = []byte("list.url.example.com")
	GlobUrlDown = []byte("down.url.example.com")
	os.Exit(m.Run())
}

var pathEncodeTests = []struct {
	Expected string
	Input    string
}{
	{"d4056d1f64fd72d1b05df35fa9d782f9c5", "/example_file.png"},
}

func TestPathEncodeExpected(t *testing.T) {
	for _, test := range pathEncodeTests {
		if res := PathEncode(test.Input); res != test.Expected {
			t.Errorf("expected output '%s' for input '%s', got '%s'", test.Expected, test.Input, res)
		}
	}
}
