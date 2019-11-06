// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcec

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

type signatureTest struct {
	name    string
	sig     []byte
	der     bool
	isValid bool
}

// decodeHex decodes the passed hex string and returns the resulting bytes.  It
// panics if an error occurs.  This is only used in the tests as a helper since
// the only way it can fail is if there is an error in the test source code.
func decodeHex(hexStr string) []byte {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		panic("invalid hex string in test source: err " + err.Error() +
			", hex: " + hexStr)
	}

	return b
}

func TestRFC6979(t *testing.T) {
	// Test vectors matching Trezor and CoreBitcoin implementations.
	// - https://github.com/trezor/trezor-crypto/blob/9fea8f8ab377dc514e40c6fd1f7c89a74c1d8dc6/tests.c#L432-L453
	// - https://github.com/oleganza/CoreBitcoin/blob/e93dd71207861b5bf044415db5fa72405e7d8fbc/CoreBitcoin/BTCKey%2BTests.m#L23-L49
	tests := []struct {
		key       string
		msg       string
		nonce     string
		signature string
	}{
		{
			"cca9fbcc1b41e5a95d369eaa6ddcff73b61a4efaa279cfc6567e8daa39cbaf50",
			"sample",
			"2df40ca70e639d89528a6b670d9d48d9165fdc0febc0974056bdce192b8e16a3",
			"e8d6af55a53a1ac260c89aa247cc677a6219132f0d031530c75c22bffde442d894a5edc9cfffa4362f0b9f2af84f4cd13ae67b0b742d18c46eb85dac8eea6171",
		},
		{
			// This signature hits the case when S is higher than halforder.
			// If S is not canonicalized (lowered by halforder), this test will fail.
			"0000000000000000000000000000000000000000000000000000000000000001",
			"Satoshi Nakamoto",
			"8f8a276c19f4149656b280621e358cce24f5f52542772691ee69063b74f15d15",
			"2f78a0720cf85bef9a24aef691fce02002c59381133ee543055d24222e2797cc78531f684122d541bbb4afe536e0b19236ad83e2e97a6277626aa5e0d1428f64",
		},
		{
			"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140",
			"Satoshi Nakamoto",
			"33a19b60e25fb6f4435af53a3d42d493644827367e6453928554f43e49aa6f90",
			"bebecd554a4c22ad202d074db855af4c974a55a5b9ef5ecf009eff232c806c63b9b37ef1d984feb2aa500777562d66f2ba468fab0491202ede12e76aa69d194a",
		},
		{
			"f8b8af8ce3c7cca5e300d33939540c10d45ce001b8f252bfbc57ba0342904181",
			"Alan Turing",
			"525a82b70e67874398067543fd84c83d30c175fdc45fdeee082fe13b1d7cfdf1",
			"864487c21bc3a7823a9dbc99fc2a6de67cf7d7cf5f2feb6585319256393ff0f2192c3292df9979a01bf26f96017bc7e5f4a2e01fa5b9e1007af1a19a58f767d0",
		},
		{
			"0000000000000000000000000000000000000000000000000000000000000001",
			"All those moments will be lost in time, like tears in rain. Time to die...",
			"38aa22d72376b4dbc472e06c3ba403ee0a394da63fc58d88686c611aba98d6b3",
			"66e3a89b2842c05cf57e5315bc2d063b1054979a4f0e5718ec5e0d68bde268296f855a1595aaa0625c822b526226a0e248fc0a1eae6f160cc97a783b8646b755",
		},
		{
			"e91671c46231f833a6406ccbea0e3e392c76c167bac1cb013f6f1013980455c2",
			"There is a computer disease that anybody who works with computers knows about. It's a very serious disease and it interferes completely with the work. The trouble with computers is that you 'play' with them!",
			"1f4b84c23a86a221d233f2521be018d9318639d5b8bbd6374a8a59232d16ad3d",
			"c5290b3ecadc0ed0d674501ffc880f2f44b2ed829bfd101676f6fb55def15f23236ace4b651a52e9d590f421668e2b1b495c9039d27db848eb3edf1d1eb60d50",
		},
	}

	for i, test := range tests {
		privKey, _ := PrivKeyFromBytes(S256(), decodeHex(test.key))
		hash := sha256.Sum256([]byte(test.msg))

		// Ensure deterministically generated nonce is the expected value.
		gotNonce := nonceRFC6979(privKey.D, hash[:], nil).Bytes()
		wantNonce := decodeHex(test.nonce)
		if !bytes.Equal(gotNonce, wantNonce) {
			t.Errorf("NonceRFC6979 #%d (%s): Nonce is incorrect: "+
				"%x (expected %x)", i, test.msg, gotNonce,
				wantNonce)
			continue
		}

		// Ensure deterministically generated signature is the expected value.
		gotSig, err := privKey.Sign(hash[:])
		if err != nil {
			t.Errorf("Sign #%d (%s): unexpected error: %v", i,
				test.msg, err)
			continue
		}
		gotSigBytes := gotSig.Serialize()
		wantSigBytes := decodeHex(test.signature)
		if !bytes.Equal(gotSigBytes, wantSigBytes) {
			t.Errorf("Sign #%d (%s): mismatched signature: %x "+
				"(expected %x)", i, test.msg, gotSigBytes,
				wantSigBytes)
			continue
		}
	}
}

func TestSignatureIsEqual(t *testing.T) {
	sig1 := &Signature{
		R: fromHex("0082235e21a2300022738dabb8e1bbd9d19cfb1e7ab8c30a23b0afbb8d178abcf3"),
		S: fromHex("24bf68e256c534ddfaf966bf908deb944305596f7bdcc38d69acad7f9c868724"),
	}
	sig2 := &Signature{
		R: fromHex("4e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41"),
		S: fromHex("181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d09"),
	}

	if !sig1.IsEqual(sig1) {
		t.Fatalf("value of IsEqual is incorrect, %v is "+
			"equal to %v", sig1, sig1)
	}

	if sig1.IsEqual(sig2) {
		t.Fatalf("value of IsEqual is incorrect, %v is not "+
			"equal to %v", sig1, sig2)
	}
}

func TestSchnorrSignatureVerify(t *testing.T) {
	// Test vectors taken from https://github.com/sipa/bips/blob/bip-schnorr/bip-schnorr/test-vectors.csv
	tests := []struct {
		pubKey    []byte
		message   []byte
		signature []byte
		valid     bool
	}{
		{
			decodeHex("0279BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"),
			decodeHex("0000000000000000000000000000000000000000000000000000000000000000"),
			decodeHex("787A848E71043D280C50470E8E1532B2DD5D20EE912A45DBDD2BD1DFBF187EF67031A98831859DC34DFFEEDDA86831842CCD0079E1F92AF177F7F22CC1DCED05"),
			true,
		},
		{
			decodeHex("02DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659"),
			decodeHex("243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89"),
			decodeHex("2A298DACAE57395A15D0795DDBFD1DCB564DA82B0F269BC70A74F8220429BA1D1E51A22CCEC35599B8F266912281F8365FFC2D035A230434A1A64DC59F7013FD"),
			true,
		},
		{
			decodeHex("03FAC2114C2FBB091527EB7C64ECB11F8021CB45E8E7809D3C0938E4B8C0E5F84B"),
			decodeHex("5E2D58D8B3BCDF1ABADEC7829054F90DDA9805AAB56C77333024B9D0A508B75C"),
			decodeHex("00DA9B08172A9B6F0466A2DEFD817F2D7AB437E0D253CB5395A963866B3574BE00880371D01766935B92D2AB4CD5C8A2A5837EC57FED7660773A05F0DE142380"),
			true,
		},
		{
			decodeHex("03DEFDEA4CDB677750A420FEE807EACF21EB9898AE79B9768766E4FAA04A2D4A34"),
			decodeHex("4DF3C3F68FCC83B27E9D42C90431A72499F17875C81A599B566C9889B9696703"),
			decodeHex("00000000000000000000003B78CE563F89A0ED9414F5AA28AD0D96D6795F9C6302A8DC32E64E86A333F20EF56EAC9BA30B7246D6D25E22ADB8C6BE1AEB08D49D"),
			true,
		},
		{
			decodeHex("031B84C5567B126440995D3ED5AABA0565D71E1834604819FF9C17F5E9D5DD078F"),
			decodeHex("0000000000000000000000000000000000000000000000000000000000000000"),
			decodeHex("52818579ACA59767E3291D91B76B637BEF062083284992F2D95F564CA6CB4E3530B1DA849C8E8304ADC0CFE870660334B3CFC18E825EF1DB34CFAE3DFC5D8187"),
			true,
		},
		{
			decodeHex("03FAC2114C2FBB091527EB7C64ECB11F8021CB45E8E7809D3C0938E4B8C0E5F84B"),
			decodeHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
			decodeHex("570DD4CA83D4E6317B8EE6BAE83467A1BF419D0767122DE409394414B05080DCE9EE5F237CBD108EABAE1E37759AE47F8E4203DA3532EB28DB860F33D62D49BD"),
			true,
		},
		{
			decodeHex("02DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659"),
			decodeHex("243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89"),
			decodeHex("2A298DACAE57395A15D0795DDBFD1DCB564DA82B0F269BC70A74F8220429BA1DFA16AEE06609280A19B67A24E1977E4697712B5FD2943914ECD5F730901B4AB7"),
			false,
		},
		{
			decodeHex("03FAC2114C2FBB091527EB7C64ECB11F8021CB45E8E7809D3C0938E4B8C0E5F84B"),
			decodeHex("5E2D58D8B3BCDF1ABADEC7829054F90DDA9805AAB56C77333024B9D0A508B75C"),
			decodeHex("00DA9B08172A9B6F0466A2DEFD817F2D7AB437E0D253CB5395A963866B3574BED092F9D860F1776A1F7412AD8A1EB50DACCC222BC8C0E26B2056DF2F273EFDEC"),
			false,
		},
		{
			decodeHex("0279BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"),
			decodeHex("0000000000000000000000000000000000000000000000000000000000000000"),
			decodeHex("787A848E71043D280C50470E8E1532B2DD5D20EE912A45DBDD2BD1DFBF187EF68FCE5677CE7A623CB20011225797CE7A8DE1DC6CCD4F754A47DA6C600E59543C"),
			false,
		},
		{
			decodeHex("03DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659"),
			decodeHex("243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89"),
			decodeHex("2A298DACAE57395A15D0795DDBFD1DCB564DA82B0F269BC70A74F8220429BA1D1E51A22CCEC35599B8F266912281F8365FFC2D035A230434A1A64DC59F7013FD"),
			false,
		},
		{
			decodeHex("03DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659"),
			decodeHex("243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89"),
			decodeHex("00000000000000000000000000000000000000000000000000000000000000009E9D01AF988B5CEDCE47221BFA9B222721F3FA408915444A4B489021DB55775F"),
			false,
		},
		{
			decodeHex("03DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659"),
			decodeHex("243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89"),
			decodeHex("0000000000000000000000000000000000000000000000000000000000000001D37DDF0254351836D84B1BD6A795FD5D523048F298C4214D187FE4892947F728"),
			false,
		},
		{
			decodeHex("03DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659"),
			decodeHex("243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89"),
			decodeHex("4A298DACAE57395A15D0795DDBFD1DCB564DA82B0F269BC70A74F8220429BA1D1E51A22CCEC35599B8F266912281F8365FFC2D035A230434A1A64DC59F7013FD"),
			false,
		},
		{
			decodeHex("03DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659"),
			decodeHex("243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89"),
			decodeHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFC2F1E51A22CCEC35599B8F266912281F8365FFC2D035A230434A1A64DC59F7013FD"),
			false,
		},
		{
			decodeHex("03DFF1D77F2A671C5F36183726DB2341BE58FEAE1DA2DECED843240F7B502BA659"),
			decodeHex("243F6A8885A308D313198A2E03707344A4093822299F31D0082EFA98EC4E6C89"),
			decodeHex("2A298DACAE57395A15D0795DDBFD1DCB564DA82B0F269BC70A74F8220429BA1DFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"),
			false,
		},
	}

	for i, test := range tests {
		pubkey, err := ParsePubKey(test.pubKey, S256())
		if err != nil {
			t.Fatal(err)
		}
		sig, err := ParseSignature(test.signature)
		if err != nil {
			t.Fatal(err)
		}
		valid := sig.Verify(test.message, pubkey)
		if valid != test.valid {
			t.Errorf("Schnorr test vector %d didn't produce correct result", i)
		}
	}
}

func TestDeterministicSchnorrSignatureGen(t *testing.T) {
	// Test vector from Bitcoin-ABC
	privKeyBytes := decodeHex("12b004fff7f4b69ef8650e767f18f11ede158148b425660723b9f9a66e61f747")
	privKey, _ := PrivKeyFromBytes(S256(), privKeyBytes)

	h1 := sha256.Sum256([]byte("Very deterministic message"))
	h2 := sha256.Sum256(h1[:])
	sig, err := privKey.Sign(h2[:])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sig.R.Bytes(), decodeHex("2c56731ac2f7a7e7f11518fc7722a166b02438924ca9d8b4d111347b81d07175")) ||
		!bytes.Equal(sig.S.Bytes(), decodeHex("71846de67ad3d913a8fdf9d8f3f73161a4c48ae81cb183b214765feb86e255ce")) {
		t.Error("Failed to generate deterministic schnorr signature")
	}
}
