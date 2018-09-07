package crypt

import (
	"testing"
)

var hexEncodeTests = []struct {
	Expected string
	Input    string
	Key      string
}{
	{"4c904b06b4650dbfa7ad6bea87ced26150", "d4056d1f64fd72d1b05df35fa9d782f9c5", "789beef750"},
	{"3f6558bfd563d0499e510be6b2f78c33", "e5022159024b540fa0837863959996bb", "5a633766d3288c4afede93832d6ef688a5fc6512a0ee2a75699fc12c3d1e8f8b"},
	{"e5e2215502fb540f90837863959976bbe302c159d24b", "The quick brown fox jumps over the lazy dog.", "e5022159024b540fa0837863959996bb"},

	// This one fails, but it shouldn't (php produces this output)
	//{"e5022b54024b54dfa08f78639599965be5022159024b54ffa06376639f9976b8c5022159", "Voix ambiguë d'un cœur qui, au zéphyr, préfère les jattes de kiwis.", "e5022159024b540fa0837863959996bb"},
}

func TestHexEncode(t *testing.T) {
	for _, test := range hexEncodeTests {
		if res := HexEncode(test.Input, test.Key); res != test.Expected {
			t.Errorf("expected output '%s' for input '%s', got '%s'", test.Expected, test.Input, res)
		}
	}
}

var hexDecodeTests = []struct {
	Input string
	Key   string
}{
	{"d4056d1f64fd72d1b05df35fa9d782f9c5", "789beef750"},
	{"e5022159024b540fa0837863959996bb", "5a633766d3288c4afede93832d6ef688a5fc6512a0ee2a75699fc12c3d1e8f8b"},
}

func TestHexDecode(t *testing.T) {
	for _, test := range hexDecodeTests {
		enc := HexEncode(test.Input, test.Key)
		dec := HexDecode(enc, test.Key)
		if dec != test.Input {
			t.Errorf("expected '%s' but got '%s'", test.Input, dec)
		}
	}
}
