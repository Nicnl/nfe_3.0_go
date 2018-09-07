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

var pathEncodeExpirableTests = []struct {
	Expected string
	Input    string
	Duration int64
	Since    int64
}{
	{"23d4bbb3a8851890209491c7fc3fcfe9-8a6c7ce73e", "/example_file.png", 0, 1536285675},
	{"bf810b0b3f3145e2f36b97f03b25fd2123d4bbb3a8851890209491c7fc3fcfe9-1815f99690", "/dir/example_file.png", 0, 1536285675},
	{"7745e7cb613bc63537d7408a728b26f6a6bf206fbe0935ca6dd40f21ad13e8ceb4ac71f5d8ff81fc5cef196d3672a40e23d4bbb3a8851890209491c7fc3fcfe9-c49a7cf7a6", "/double//slash/example_file.png", 0, 1536285675},
	{"17e02de9e3f0f3015b40c614d85b6b21bdfcb9385cd0df4a5b22ecb243e52df8-0a07d84fdf", " no / trim.png ", 0, 1536285675},
	{"8934e198e0db4bba33356f4a4370e334944f44a7457682cee3b946bc27cc67b9617e0a0fc6352e362e289a828181291badc880784c92c697891a70d9344138a7c05560985ae3c17421466b3b471cb894fff6f6d5a6a0fe24c79db50162258b8b23d4bbb3a8851890209491c7fc3fcfe9-20753697cc", "/very/long/chain/of/subdirs/everywhere/example_file.png", 0, 1536285675},
	{"3b3d0dc83f62c65406959818155bcf0d6d14ee6c6793c22861326e17b7df70e14252aa4c3802eaef085d57152e470cb50729787138b561ac20d8cc71bd935e83609c6bb507525863d98bd5b5b7371f34e677f1ed579b4e7c82c6283ff7e7a1dbac39da8525b5dd520f7b09adb112972db063411ba3c6d6277b277b960f3f303fa63ceb35beff9453f597ebb3ed814bedb35ca724fe9f2456f452f5062aa09342c74c69cceaabacdd5999f6fe0bba32c4b57472eeedf418f4960951353db0dc5b-0b9ad5052e", "/ Voix/ ambiguë/ d'un/ cœur/ qui,/ au/ zéphyr,/ préfère/ les/ jattes/ de/ kiwis.png", 0, 1536285675},
	{"d2bac3505a1cc3f9bd871cd2044c66d5-598b436c3c", "", 0, 1536285675},
	{"07f4706075bada774db5d4773a71cb34-57913242e420c5bddd35c2e72d2143b04", "/example_file.png", 900, 1536285675},
	{"ac9096061a3517afe43ece4c80d39dbf-0efee946a824518618f734064e92e241d", "/example_file.png", 900, 1536285676},
	{"2204f56944c658059bf19cda283ad3a5-388f2aad2e98b0e94288756cf555b0df7", "/example_file.png", 901, 1536285676},
}

func TestPathEncodeExpirable(t *testing.T) {
	GlobSalt = []byte("example_hash_salt")
	GlobUrlList = []byte("list.url.example.com")
	GlobUrlDown = []byte("down.url.example.com")

	for _, test := range pathEncodeExpirableTests {
		if res := PathEncodeExpirable(test.Input, test.Duration, test.Since); res != test.Expected {
			t.Errorf("expected output '%s' for input '%s', got '%s'", test.Expected, test.Input, res)
		}
	}
}
