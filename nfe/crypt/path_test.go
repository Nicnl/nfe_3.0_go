package crypt

import (
	"testing"
)

var pathEncodeTests = []struct {
	Expected string
	Input    string
}{
	{"d4056d1f64fd72d1b05df35fa9d782f9c5", "/example_file.png"},
	{"60b2bd67fba9af238324f988e8cdb031d4056d1f64fd72d1b05df35fa9d782f927", "/dir/example_file.png"},
	{"287699272da32076c790a2122f23e90657e0d2cb7a719f0bfd9d61b95abbabde65dd23519467eb3deca87bf5e31a671ed4056d1f64fd72d1b05df35fa9d782f941", "/double//slash/example_file.png"},
	{"c811df45af685d42eb0928ac85f32e316e2d6b941848398bebeb4e4af08de0081e", " no / trim.png "},
	{"3a6593f4ac43a5fbc3fec1d2f018a6444570f60301eeec0f7372a844d4642ac912afbc6b82ad8877bee1fc1a3e29ec2b5ef932d4080a20d819d3d261e1e9fbb7718612f4165b2bb5b10fcdc3f4b47ba4a027a83162185865575617991fcd4e9bd4056d1f64fd72d1b05df35fa9d782f9b3", "/very/long/chain/of/subdirs/everywhere/example_file.png"},
	{"ec6ebf24fbda2095965efaa0c2f3821d1e4590c8230b2c69f1fbc0af647733f1f3835ca8f47a44209816b9addbefcfc5b85a2addf42dcbedb0912e096a3b119311cd1d11c3cab2a46944374d64dfd24497a8a3491303a8bd128f8ac7a48f64eb5d6a8ce1e12d37939f346b356eba5a3d6194f3776f3e30680be0dd2ebcd7f34f576d9d917a67fe9485504d4b9a290efd648d5980ba078e97841b579ed7485652787d1b28a613061ee9525886b852f5d466a5244aa96c723526c2b3cdea589f6b52", "/ Voix/ ambiguë/ d'un/ cœur/ qui,/ au/ zéphyr,/ préfère/ les/ jattes/ de/ kiwis.png"},
	{"83eb75bc16842d3a4d407e6ab1e429e525", ""},
}

func TestPathEncodeExpected(t *testing.T) {
	GlobSalt = []byte("example_hash_salt")
	GlobUrlList = []byte("list.url.example.com")
	GlobUrlDown = []byte("down.url.example.com")

	for _, test := range pathEncodeTests {
		if res := PathEncode(test.Input); res != test.Expected {
			t.Errorf("expected output '%s' for input '%s', got '%s'", test.Expected, test.Input, res)
		}
	}
}
