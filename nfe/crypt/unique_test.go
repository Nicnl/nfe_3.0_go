package crypt

import (
	"testing"
)

var confSalt []byte = []byte("example_hash_salt")
var confUrlList []byte = []byte("list.url.example.com")
var confUrlDown []byte = []byte("down.url.example.com")

var tests = []struct {
	Expected string
	Input    string
}{
	{"5a633766d3288c4afede93832d6ef688", "Hello World!"},
	{"e5022159024b540fa0837863959996bb", "THIS IS OVER > 9000"},
	{"0f0ef94789750a8238e14fe3a81f9323", "Voix ambiguë d'un cœur qui, au zéphyr, préfère les jattes de kiwis."},
	{"6f0959c138bbdf5eeea6522b1295f4f5", "The quick brown fox jumps over the lazy dog."},
	{"a5fc6512a0ee2a75699fc12c3d1e8f8b", " notrim "},
}

func TestUniqueExpected(t *testing.T) {
	for _, test := range tests {
		if res := Unique([]byte(test.Input), confSalt, confUrlList, confUrlDown); res != test.Expected {
			t.Errorf("expected output '%s' for input '%s', '%s', '%s', '%s', got '%s'", test.Expected, test.Input, string(confSalt), string(confUrlList), string(confUrlDown), res)
		}
	}
}

func TestUniqueExpectedAllNil(t *testing.T) {
	expected := "4a7d1ed414474e4033ac29ccb8653d9b"
	if res := Unique(nil, nil, nil, nil); res != expected {
		t.Errorf("expected '%s', got '%s'", expected, res)
	}
}

func TestUniqueExpectedAllEmpty(t *testing.T) {
	expected := "4a7d1ed414474e4033ac29ccb8653d9b"
	if res := Unique([]byte{}, []byte{}, []byte{}, []byte{}); res != expected {
		t.Errorf("expected '%s', got '%s'", expected, res)
	}
}
