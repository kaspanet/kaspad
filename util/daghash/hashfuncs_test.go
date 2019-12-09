// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package daghash

import (
	"fmt"
	"testing"
)

// TestHashFuncs ensures the hash functions which perform hash(b) work as
// expected.
func TestHashFuncs(t *testing.T) {
	tests := []struct {
		out string
		in  string
	}{
		{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", ""},
		{"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb", "a"},
		{"fb8e20fc2e4c3f248c60c39bd652f3c1347298bb977b8b4d5903b85055620603", "ab"},
		{"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", "abc"},
		{"88d4266fd4e6338d13b845fcf289579d209c897823b9217da3e161936f031589", "abcd"},
		{"36bbe50ed96841d10443bcb670d6554f0a34b761be67ec9c4a8ad2c0c44ca42c", "abcde"},
		{"bef57ec7f53a6d40beb640a780a639c83bc29ac8a9816f1fc6c5c6dcd93c4721", "abcdef"},
		{"7d1a54127b222502f5b79b5fb0803061152a44f92b37e23c6527baf665d4da9a", "abcdefg"},
		{"9c56cc51b374c3ba189210d5b6d4bf57790d351c96c47c02190ecf1e430635ab", "abcdefgh"},
		{"19cc02f26df43cc571bc9ed7b0c4d29224a3ec229529221725ef76d021c8326f", "abcdefghi"},
		{"72399361da6a7754fec986dca5b7cbaf1c810a28ded4abaf56b2106d06cb78b0", "abcdefghij"},
		{"a144061c271f152da4d151034508fed1c138b8c976339de229c3bb6d4bbb4fce", "Discard medicine more than two years old."},
		{"6dae5caa713a10ad04b46028bf6dad68837c581616a1589a265a11288d4bb5c4", "He who has a shady past knows that nice guys finish last."},
		{"ae7a702a9509039ddbf29f0765e70d0001177914b86459284dab8b348c2dce3f", "I wouldn't marry him with a ten foot pole."},
		{"6748450b01c568586715291dfa3ee018da07d36bb7ea6f180c1af6270215c64f", "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
		{"92fb79584704377a7556e8516707162983804dbe92b0c440ee9715bcb9c7e46f", "The days of the digital watch are numbered. -Tom Stoppard"},
		{"7102cfd76e2e324889eece5d6c41921b1e142a4ac5a2692be78803097f6a48d8", "Nepal premier won't resign."},
		{"23b1018cd81db1d67983c5f7417c44da9deb582459e378d7a068552ea649dc9f", "For every action there is an equal and opposite government program."},
		{"8001f190dfb527261c4cfcab70c98e8097a7a1922129bc4096950e57c7999a5a", "His money is twice tainted: 'taint yours and 'taint mine."},
		{"8c87deb65505c3993eb24b7a150c4155e82eee6960cf0c3a8114ff736d69cad5", "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
		{"bfb0a67a19cdec3646498b2e0f751bddc41bba4b7f30081b0b932aad214d16d7", "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
		{"56229cec0edb6b1cd72bf3f6a407c954af974b01d6e64d5b06b926820a645a79", "size: a.out: bad magic"},
		{"f058498afd2bccc48d3060ca6e1de5301dfc6c479ab2c5547aa256d116c3ed62", "The major problem is with sendmail. -Mark Horton"},
		{"e66b6d3e83f81384ae9b52d664d7639e40481321adc7a8c6d442dad1f98a1c73", "Give me a rock, paper and scissors and I will move the world. CCFestoon"},
		{"049d5e26d4f10222cd841a119e38bd8d2e0d1129728688449575d4ff42b842c1", "If the enemy is within range, then so are you."},
		{"0e116838e3cc1c1a14cd045397e29b4d087aa11b0853fc69ec82e90330d60949", "It's well we cannot hear the screams/That we create in others' dreams."},
		{"4f7d8eb5bcf11de2a56b971021a444aa4eafd6ecd0f307b5109e4e776cd0fe46", "You remind me of a TV show, but that's all right: I watch it anyway."},
		{"61c0cc4c4bd8406d5120b3fb4ebc31ce87667c162f29468b3c779675a85aebce", "C is as portable as Stonehedge!!"},
		{"1fb2eb3688093c4a3f80cd87a5547e2ce940a4f923243a79a2a1e242220693ac", "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
		{"6ee3a23a4afe025102fe5382a5fedebc383f2bd081cdb252e9422041fe32956e", "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction. Lewis-Randall Rule"},
		{"0ac1e12a1a25b7a986580dc95b64e30b0849cd6c5f5945f8680570d81a28771d", "How can you write a big system without C++? -Paul Glick"},
	}

	// Ensure the hash function which returns a byte slice returns the
	// expected result.
	for _, test := range tests {
		h := fmt.Sprintf("%x", HashB([]byte(test.in)))
		if h != test.out {
			t.Errorf("HashB(%q) = %s, want %s", test.in, h, test.out)
			continue
		}
	}

	// Ensure the hash function which returns a Hash returns the expected
	// result.
	for _, test := range tests {
		hash := HashH([]byte(test.in))
		h := fmt.Sprintf("%x", hash[:])
		if h != test.out {
			t.Errorf("HashH(%q) = %s, want %s", test.in, h, test.out)
			continue
		}
	}
}

// TestDoubleHashFuncs ensures the hash functions which perform hash(hash(b))
// work as expected.
func TestDoubleHashFuncs(t *testing.T) {
	tests := []struct {
		out string
		in  string
	}{
		{"5df6e0e2761359d30a8275058e299fcc0381534545f55cf43e41983f5d4c9456", ""},
		{"bf5d3affb73efd2ec6c36ad3112dd933efed63c4e1cbffcfa88e2759c144f2d8", "a"},
		{"a1ff8f1856b5e24e32e3882edd4a021f48f28a8b21854b77fdef25a97601aace", "ab"},
		{"4f8b42c22dd3729b519ba6f68d2da7cc5b2d606d05daed5ad5128cc03e6c6358", "abc"},
		{"7e9c158ecd919fa439a7a214c9fc58b85c3177fb1613bdae41ee695060e11bc6", "abcd"},
		{"1d72b6eb7ba8b9709c790b33b40d8c46211958e13cf85dbcda0ed201a99f2fb9", "abcde"},
		{"ce65d4756128f0035cba4d8d7fae4e9fa93cf7fdf12c0f83ee4a0e84064bef8a", "abcdef"},
		{"dad6b965ad86b880ceb6993f98ebeeb242de39f6b87a458c6510b5a15ff7bbf1", "abcdefg"},
		{"b9b12e7125f73fda20b8c4161fb9b4b146c34cf88595a1e0503ca2cf44c86bc4", "abcdefgh"},
		{"546db09160636e98405fbec8464a84b6464b32514db259e235eae0445346ffb7", "abcdefghi"},
		{"27635cf23fdf8a10f4cb2c52ade13038c38718c6d7ca716bfe726111a57ad201", "abcdefghij"},
		{"ae0d8e0e7c0336f0c3a72cefa4f24b625a6a460417a921d066058a0b81e23429", "Discard medicine more than two years old."},
		{"eeb56d02cf638f87ea8f11ebd5b0201afcece984d87be458578d3cfb51978f1b", "He who has a shady past knows that nice guys finish last."},
		{"dc640bf529608a381ea7065ecbcd0443b95f6e4c008de6e134aff1d36bd4b9d8", "I wouldn't marry him with a ten foot pole."},
		{"42e54375e60535eb07fc15c6350e10f2c22526f84db1d6f6bba925e154486f33", "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
		{"dc468358256e446bf153fa12be5e2474e66e91b63f3d0ab19ca1ffd3ac90935d", "The days of the digital watch are numbered. -Tom Stoppard"},
		{"590c24d1877c1919fad12fe01a8796999e9d20cfbf9bc9bc72fa0bd69f0b04dd", "Nepal premier won't resign."},
		{"37d270687ee8ebafcd3c1a32f56e1e1304b3c93f252cb637d57a66d59c475eca", "For every action there is an equal and opposite government program."},
		{"306828fd89278838bb1c544c3032a1fd25ea65c40bba586437568828a5fbe944", "His money is twice tainted: 'taint yours and 'taint mine."},
		{"49965777eac71faf1e2fb0f6b239ba2fae770977940fd827bcbfe15def6ded53", "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
		{"df99ee4e87dd3fb07922dee7735997bbae8f26db20c86137d4219fc4a37b77c3", "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
		{"3416e4d22bebefd3be926a39d11db493c77f245df8092275a963281d301245a9", "size: a.out: bad magic"},
		{"7e2a6bcda29be82a048ee6966a1a3b8f8d01199e3e2172c35f8f3362682e90c0", "The major problem is with sendmail. -Mark Horton"},
		{"cfbf69a14b29a43fb227209160f8d89781e75a9c1fbe3f37158e14b948f23bf0", "Give me a rock, paper and scissors and I will move the world. CCFestoon"},
		{"2e7aa1b362c94efdbff582a8bd3f7f61c8ce4c25bbde658ef1a7ae1010e2126f", "If the enemy is within range, then so are you."},
		{"e6729d51240b1e1da76d822fd0c55c75e409bcb525674af21acae1f11667c8ca", "It's well we cannot hear the screams/That we create in others' dreams."},
		{"09945e4d2743eb669f85e4097aa1cc39ea680a0b2ae2a65a42a5742b3b809610", "You remind me of a TV show, but that's all right: I watch it anyway."},
		{"1018d8b2870a974887c5174360f0fbaf27958eef15b24522a605c5dae4ae0845", "C is as portable as Stonehedge!!"},
		{"97c76b83c6645c78c261dcdc55d44af02d9f1df8057f997fd08c310c903624d5", "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
		{"ff0f1be0bbadb42baa244e4f71f6942d17ad4b0147405f3f56c695307f3dc55c", "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction. Lewis-Randall Rule"},
		{"5d1ba8dd3f81ac07a5c0065db708b91db1befe9a145d60db3bd2217a6b63641e", "How can you write a big system without C++? -Paul Glick"},
	}

	// Ensure the hash function which returns a byte slice returns the
	// expected result.
	for _, test := range tests {
		h := fmt.Sprintf("%x", DoubleHashB([]byte(test.in)))
		if h != test.out {
			t.Errorf("DoubleHashB(%q) = %s, want %s", test.in, h,
				test.out)
			continue
		}
	}

	// Ensure the hash function which returns a Hash returns the expected
	// result.
	for _, test := range tests {
		hash := DoubleHashH([]byte(test.in))
		h := fmt.Sprintf("%x", hash[:])
		if h != test.out {
			t.Errorf("DoubleHashH(%q) = %s, want %s", test.in, h,
				test.out)
			continue
		}
	}
}
