package crypt

import (
	"math/rand"
	"nfe_3.0_go/nfe/vfs"
	"testing"
	"time"
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
	Duration time.Duration
	Since    time.Time
}{
	// Tests avec une durée de 0 (= même comportement que le PHP)
	{"23d4bbb3a8851890209491c7fc3fcfe9-8a6c7ce73e", "/example_file.png", 0, time.Unix(1536285675, 0)},
	{"759d6437dc87734e9f5f63f149775365-32ea9e80ea", "/example_file.png", 0, time.Unix(1536285674, 0)},
	{"bf810b0b3f3145e2f36b97f03b25fd2123d4bbb3a8851890209491c7fc3fcfe9-1815f99690", "/dir/example_file.png", 0, time.Unix(1536285675, 0)},
	{"7745e7cb613bc63537d7408a728b26f6a6bf206fbe0935ca6dd40f21ad13e8ceb4ac71f5d8ff81fc5cef196d3672a40e23d4bbb3a8851890209491c7fc3fcfe9-c49a7cf7a6", "/double//slash/example_file.png", 0, time.Unix(1536285675, 0)},
	{"17e02de9e3f0f3015b40c614d85b6b21bdfcb9385cd0df4a5b22ecb243e52df8-0a07d84fdf", " no / trim.png ", 0, time.Unix(1536285675, 0)},
	{"8934e198e0db4bba33356f4a4370e334944f44a7457682cee3b946bc27cc67b9617e0a0fc6352e362e289a828181291badc880784c92c697891a70d9344138a7c05560985ae3c17421466b3b471cb894fff6f6d5a6a0fe24c79db50162258b8b23d4bbb3a8851890209491c7fc3fcfe9-20753697cc", "/very/long/chain/of/subdirs/everywhere/example_file.png", 0, time.Unix(1536285675, 0)},
	{"3b3d0dc83f62c65406959818155bcf0d6d14ee6c6793c22861326e17b7df70e14252aa4c3802eaef085d57152e470cb50729787138b561ac20d8cc71bd935e83609c6bb507525863d98bd5b5b7371f34e677f1ed579b4e7c82c6283ff7e7a1dbac39da8525b5dd520f7b09adb112972db063411ba3c6d6277b277b960f3f303fa63ceb35beff9453f597ebb3ed814bedb35ca724fe9f2456f452f5062aa09342c74c69cceaabacdd5999f6fe0bba32c4b57472eeedf418f4960951353db0dc5b-0b9ad5052e", "/ Voix/ ambiguë/ d'un/ cœur/ qui,/ au/ zéphyr,/ préfère/ les/ jattes/ de/ kiwis.png", 0, time.Unix(1536285675, 0)},
	{"d2bac3505a1cc3f9bd871cd2044c66d5-598b436c3c", "", 0, time.Unix(1536285675, 0)},

	// Tests avec une durée > 0 (= comportement de la nouvelle version, plus sécure
	{"3ddbf1108305b4cf91b13582a62d2b86-85d91ffc4ebfb9c4f31de0a8d24567646", "/example_file.png", 900 * time.Second, time.Unix(1536285671, 0)},
	{"7b02670cc200872f59924e789c6bb0bb-a5c76828ea90f0e5ae8f389b3076329d3", "/example_file.png", 900 * time.Second, time.Unix(1536285672, 0)},
	{"9c69c6f88a826ae2fdffa7b9dcdcb3e1-9fe2b38214a40b3b851f8e39a71b13e47", "/example_file.png", 901 * time.Second, time.Unix(1536285672, 0)},
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

var checkHashTests = []struct {
	Expected bool
	Input    string
}{
	{true, "3ddbf1108305b4cf91b13582a62d2b86-85d91ffc4ebfb9c4f31de0a8d24567646"},
	{true, "7b02670cc200872f59924e789c6bb0bb-a5c76828ea90f0e5ae8f389b3076329d3"},
	{true, "9c69c6f88a826ae2fdffa7b9dcdcb3e1-9fe2b38214a40b3b851f8e39a71b13e47"},
	{true, "d4056d1f64fd72d1b05df35fa9d782f9c5"},
	{true, "60b2bd67fba9af238324f988e8cdb031d4056d1f64fd72d1b05df35fa9d782f927"},
	{true, "287699272da32076c790a2122f23e90657e0d2cb7a719f0bfd9d61b95abbabde65dd23519467eb3deca87bf5e31a671ed4056d1f64fd72d1b05df35fa9d782f941"},
	{true, "c811df45af685d42eb0928ac85f32e316e2d6b941848398bebeb4e4af08de0081e"},
	{true, "3a6593f4ac43a5fbc3fec1d2f018a6444570f60301eeec0f7372a844d4642ac912afbc6b82ad8877bee1fc1a3e29ec2b5ef932d4080a20d819d3d261e1e9fbb7718612f4165b2bb5b10fcdc3f4b47ba4a027a83162185865575617991fcd4e9bd4056d1f64fd72d1b05df35fa9d782f9b3"},
	{true, "ec6ebf24fbda2095965efaa0c2f3821d1e4590c8230b2c69f1fbc0af647733f1f3835ca8f47a44209816b9addbefcfc5b85a2addf42dcbedb0912e096a3b119311cd1d11c3cab2a46944374d64dfd24497a8a3491303a8bd128f8ac7a48f64eb5d6a8ce1e12d37939f346b356eba5a3d6194f3776f3e30680be0dd2ebcd7f34f576d9d917a67fe9485504d4b9a290efd648d5980ba078e97841b579ed7485652787d1b28a613061ee9525886b852f5d466a5244aa96c723526c2b3cdea589f6b52"},
	{true, "83eb75bc16842d3a4d407e6ab1e429e525"},

	{false, "3ddbf1108305b4cf91b13582a62c2b86-85d91ffc4ebfb9c4f31de0a8d24567646"},
	{false, "7b02670cc200872f59924e789c6bb0bb-a5c76828ea90f0e5ae8f389b3076329d4"},
	{false, "9c69c6f88a826ae2fdffa7b9dcdb3e1-9fe2b38214a40b3b851f8e39a71b13e47"},
	{false, "d4056d1f64fd72d1b05df35fa9d782f9c6"},
	{false, "60b2bd67fba9af238324f988e8adb031d4056d1f64fd72d1b05df35fa9d782f927"},
	{false, "28769927ada32076c790a2122f23e90657e0d2cb7a719f0bfd9d61b95abbabde65dd23519467eb3deca87bf5e31a671ed4056d1f64fd72d1b05df35fa9d782f941"},
	{false, "c811df45bf685d42eb0928ac85f32e316e2d6b941848398bebeb4e4af08de0081e"},
	{false, "3a6593f4cc43a5fbc3fec1d2f018a6444570f60301eeec0f7372a844d4642ac912afbc6b82ad8877bee1fc1a3e29ec2b5ef932d4080a20d819d3d261e1e9fbb7718612f4165b2bb5b10fcdc3f4b47ba4a027a83162185865575617991fcd4e9bd4056d1f64fd72d1b05df35fa9d782f9b3"},
	{false, "ec6ebf24dbda2095965efaa0c2f3821d1e4590c8230b2c69f1fbc0af647733f1f3835ca8f47a44209816b9addbefcfc5b85a2addf42dcbedb0912e096a3b119311cd1d11c3cab2a46944374d64dfd24497a8a3491303a8bd128f8ac7a48f64eb5d6a8ce1e12d37939f346b356eba5a3d6194f3776f3e30680be0dd2ebcd7f34f576d9d917a67fe9485504d4b9a290efd648d5980ba078e97841b579ed7485652787d1b28a613061ee9525886b852f5d466a5244aa96c723526c2b3cdea589f6b52"},
	{false, "83eb75bc16842d4a4d407e6ab1e429e525"},
}

func TestCheckHash(t *testing.T) {
	GlobSalt = []byte("example_hash_salt")
	GlobUrlList = []byte("list.url.example.com")
	GlobUrlDown = []byte("down.url.example.com")

	for _, test := range checkHashTests {
		if res := CheckHash(test.Input); res != test.Expected {
			t.Errorf("expected output '%t' for input '%s', got '%t'", test.Expected, test.Input, res)
		}
	}
}

var checkFind = map[string]bool{
	"/movies":                               true,
	"/music":                                true,
	"/books":                                true,
	"/other":                                true,
	"/software":                             true,
	"/games":                                true,
	"/software/OpenOffice":                  true,
	"/software/LibreOffice":                 true,
	"/software/Firefox":                     true,
	"/software/Thunderbird":                 true,
	"/software/Audacity":                    true,
	"/software/The Gimp":                    true,
	"/software/VLC":                         true,
	"/software/Handbrake":                   true,
	"/software/Notepad++":                   true,
	"/software/Audacity/audacity_1.0.0.exe": true,

	"":                          false,
	"/":                         false,
	"/nonexisting":              false,
	"/software/":                false,
	"/software/Audacity/qsdqsd": false,
}

var vfsFake = vfs.Fake{
	Structure: map[string][]string{
		"/":                  {"movies", "music", "books", "other", "software", "games"},
		"/software":          {"OpenOffice", "LibreOffice", "Firefox", "Thunderbird", "Audacity", "The Gimp", "VLC", "Handbrake", "Notepad++"},
		"/software/Audacity": {"audacity_1.0.0.exe"},
	},
}

func TestFindWithoutExpiration(t *testing.T) {
	for path, expected := range checkFind {
		findPath, _, err := Find(PathEncode(path), time.Unix(0, 0), &vfsFake)
		if err != nil {
			if expected {
				t.Error(path, " => ", err)
			}
		} else if findPath == path != expected {
			t.Errorf("expected findPath to be '%s', got '%s'", path, findPath)
		}
	}
}

func TestFindWithExpiration(t *testing.T) {
	timestamp := time.Unix(1536285675, 0)
	for path, expected := range checkFind {
		findPath, _, err := Find(PathEncodeExpirable(path, 10, timestamp), timestamp, &vfsFake)
		if err != nil {
			if expected {
				t.Error(path, " => ", err)
			}
		} else if findPath == path != expected {
			t.Errorf("expected findPath to be '%s', got '%s'", path, findPath)
		}
	}
}

func TestFindExpirationReached(t *testing.T) {
	timestamp := time.Unix(1536285675, 0)
	var duration = 10 * time.Second
	for path, expected := range checkFind {
		if !expected {
			continue
		}

		findPath, _, err := Find(PathEncodeExpirable(path, duration, timestamp), timestamp.Add(duration).Add(time.Second), &vfsFake)
		if err == nil {
			t.Errorf("err should not be nil")
		} else if findPath != "" {
			t.Errorf("findPath should be empty, got '%s'", findPath)
		}
	}
}

func TestAddBandwidthLimit(t *testing.T) {
	rand.Seed(4063542065746508042)
	var expectedLimit int64 = 12345678987654321

	for path := range checkFind {
		encodedPath := PathEncode(path)
		bandwidthPath := AddBandwidthLimit(encodedPath, expectedLimit)

		decodedLink, limit, err := GetBandwidthLimit(bandwidthPath)
		if err != nil {
			t.Errorf("err should be nil: %s", err)
		}
		if limit != expectedLimit {
			t.Errorf("expected limit to be '%d', got '%d'", expectedLimit, limit)
		}
		if encodedPath != decodedLink {
			t.Errorf("expected limit to be '%s', got '%s'", encodedPath, decodedLink)
		}
	}
}

func TestFindWithBandwidthLimitNotExpirable(t *testing.T) {
	timestamp := time.Unix(1536285675, 0)
	var duration = 10 * time.Second
	rand.Seed(4063542065746508042)
	var expectedLimit int64 = 12345678987654321

	for path, expected := range checkFind {
		if !expected {
			continue
		}

		encodedPath := PathEncode(path)
		bandwidthPath := AddBandwidthLimit(encodedPath, expectedLimit)

		vfsPath, bandwidthLimit, err := Find(bandwidthPath, timestamp.Add(duration), &vfsFake)
		if err != nil {
			t.Errorf("err should be nil: %s", err)
		}
		if expectedLimit != bandwidthLimit {
			t.Errorf("expected limit to be '%d', got '%d'", expectedLimit, bandwidthLimit)
		}
		if path != vfsPath {
			t.Errorf("expected limit to be '%s', got '%s'", vfsPath, path)
		}
	}
}

func TestFindWithBandwidthLimitExpirable(t *testing.T) {
	timestamp := time.Unix(1536285675, 0)
	var duration = 10 * time.Second
	rand.Seed(4063542065746508042)
	var expectedLimit int64 = 123456789

	for path, expected := range checkFind {
		if !expected {
			continue
		}

		encodedPath := PathEncodeExpirable(path, duration, timestamp)
		bandwidthPath := AddBandwidthLimit(encodedPath, expectedLimit)

		vfsPath, bandwidthLimit, err := Find(bandwidthPath, timestamp, &vfsFake)
		if err != nil {
			t.Errorf("err should be nil: %s", err)
		}
		if expectedLimit != bandwidthLimit {
			t.Errorf("expected limit to be '%d', got '%d'", expectedLimit, bandwidthLimit)
		}
		if path != vfsPath {
			t.Errorf("expected limit to be '%s', got '%s'", vfsPath, path)
		}
	}
}
