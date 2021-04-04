// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"bytes"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
)

// mustParseShortForm parses the passed short form script and returns the
// resulting bytes. It panics if an error occurs. This is only used in the
// tests as a helper since the only way it can fail is if there is an error in
// the test source code.
func mustParseShortForm(script string, version uint16) []byte {
	s, err := parseShortForm(script, version)
	if err != nil {
		panic("invalid short form script in test source: err " +
			err.Error() + ", script: " + script)
	}

	return s
}

// newAddressPubKey returns a new util.AddressPubKey from the
// provided public key. It panics if an error occurs. This is only used in the tests
// as a helper since the only way it can fail is if there is an error in the
// test source code.
func newAddressPubKey(publicKey []byte) util.Address {
	addr, err := util.NewAddressPubKey(publicKey, util.Bech32PrefixKaspa)
	if err != nil {
		panic("invalid public key in test source")
	}

	return addr
}

// newAddressScriptHash returns a new util.AddressScriptHash from the
// provided hash. It panics if an error occurs. This is only used in the tests
// as a helper since the only way it can fail is if there is an error in the
// test source code.
func newAddressScriptHash(scriptHash []byte) util.Address {
	addr, err := util.NewAddressScriptHashFromHash(scriptHash,
		util.Bech32PrefixKaspa)
	if err != nil {
		panic("invalid script hash in test source")
	}

	return addr
}

// TestExtractScriptPubKeyAddrs ensures that extracting the type, addresses, and
// number of required signatures from scriptPubKeys works as intended.
func TestExtractScriptPubKeyAddrs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script *externalapi.ScriptPublicKey
		addr   util.Address
		class  ScriptClass
	}{
		{
			name: "standard p2pk",
			script: &externalapi.ScriptPublicKey{
				Script:  hexToBytes("202454a285d8566b0cb2792919536ee0f1b6f69b58ba59e9850ecbc91eef722daeac"),
				Version: 0,
			},
			addr:  newAddressPubKey(hexToBytes("2454a285d8566b0cb2792919536ee0f1b6f69b58ba59e9850ecbc91eef722dae")),
			class: PubKeyTy,
		},
		{
			name: "standard p2sh",
			script: &externalapi.ScriptPublicKey{
				Script: hexToBytes("aa2063bcc565f9e68ee0189dd5cc67f1b" +
					"0e5f02f45cbad06dd6ddee55cbca9a9e37187"),
				Version: 0,
			},
			addr: newAddressScriptHash(hexToBytes("63bcc565f9e6" +
				"8ee0189dd5cc67f1b0e5f02f45cbad06dd6ddee55cbca9a9e371")),
			class: ScriptHashTy,
		},

		// The below are nonstandard script due to things such as
		// invalid pubkeys, failure to parse, and not being of a
		// standard form.

		{
			name: "p2pk with uncompressed pk missing OP_CHECKSIG",
			script: &externalapi.ScriptPublicKey{
				Script: hexToBytes("410411db93e1dcdb8a016b49840f8c53b" +
					"c1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddf" +
					"b84ccf9744464f82e160bfa9b8b64f9d4c03f999b864" +
					"3f656b412a3"),
				Version: 0,
			},
			addr:  nil,
			class: NonStandardTy,
		},
		{
			name: "valid signature from a sigscript - no addresses",
			script: &externalapi.ScriptPublicKey{
				Script: hexToBytes("47304402204e45e16932b8af514961a1d" +
					"3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41022" +
					"0181522ec8eca07de4860a4acdd12909d831cc56cbba" +
					"c4622082221a8768d1d0901"),
				Version: 0,
			},
			addr:  nil,
			class: NonStandardTy,
		},
		// Note the technically the pubkey is the second item on the
		// stack, but since the address extraction intentionally only
		// works with standard scriptPubKeys, this should not return any
		// addresses.
		{
			name: "valid sigscript to reedeem p2pk - no addresses",
			script: &externalapi.ScriptPublicKey{
				Script: hexToBytes("493046022100ddc69738bf2336318e4e0" +
					"41a5a77f305da87428ab1606f023260017854350ddc0" +
					"22100817af09d2eec36862d16009852b7e3a0f6dd765" +
					"98290b7834e1453660367e07a014104cd4240c198e12" +
					"523b6f9cb9f5bed06de1ba37e96a1bbd13745fcf9d11" +
					"c25b1dff9a519675d198804ba9962d3eca2d5937d58e" +
					"5a75a71042d40388a4d307f887d"),
				Version: 0,
			},
			addr:  nil,
			class: NonStandardTy,
		},
		{
			name: "empty script",
			script: &externalapi.ScriptPublicKey{
				Script:  []byte{},
				Version: 0,
			},
			addr:  nil,
			class: NonStandardTy,
		},
		{
			name: "script that does not parse",
			script: &externalapi.ScriptPublicKey{
				Script:  []byte{OpData45},
				Version: 0,
			},
			addr:  nil,
			class: NonStandardTy,
		},
	}

	t.Logf("Running %d tests.", len(tests))
	for i, test := range tests {
		class, addr, _ := ExtractScriptPubKeyAddress(
			test.script, &dagconfig.MainnetParams)

		if !reflect.DeepEqual(addr, test.addr) {
			t.Errorf("ExtractScriptPubKeyAddress #%d (%s) unexpected "+
				"address\ngot  %v\nwant %v", i, test.name,
				addr, test.addr)
			continue
		}

		if class != test.class {
			t.Errorf("ExtractScriptPubKeyAddress #%d (%s) unexpected "+
				"script type - got %s, want %s", i, test.name,
				class, test.class)
			continue
		}
	}
}

// TestCalcScriptInfo ensures the CalcScriptInfo provides the expected results
// for various valid and invalid script pairs.
func TestCalcScriptInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		sigScript    string
		scriptPubKey string

		isP2SH bool

		scriptInfo    ScriptInfo
		scriptInfoErr error
	}{
		{
			// Invented scripts, the hashes do not match
			// Truncated version of test below:
			name: "scriptPubKey doesn't parse",
			sigScript: "1 81 DATA_8 2DUP EQUAL NOT VERIFY ABS " +
				"SWAP ABS EQUAL",
			scriptPubKey: "BLAKE2B DATA_32 0xfe441065b6532231de2fac56" +
				"3152205ec4f59cfe441065b6532231de2fac56",
			isP2SH:        true,
			scriptInfoErr: scriptError(ErrMalformedPush, ""),
		},
		{
			name: "sigScript doesn't parse",
			// Truncated version of p2sh script below.
			sigScript: "1 81 DATA_8 2DUP EQUAL NOT VERIFY ABS " +
				"SWAP ABS",
			scriptPubKey: "BLAKE2B DATA_32 0xfe441065b6532231de2fac56" +
				"3152205ec4f59c74fe441065b6532231de2fac56 EQUAL",
			isP2SH:        true,
			scriptInfoErr: scriptError(ErrMalformedPush, ""),
		},
		{
			// Invented scripts, the hashes do not match
			name: "p2sh standard script",
			sigScript: "1 81 DATA_34 DATA_32 0x010203" +
				"0405060708090a0b0c0d0e0f1011121314fe441065b6532231de2fac56 " +
				"CHECKSIG",
			scriptPubKey: "BLAKE2B DATA_32 0xfe441065b6532231de2fac56" +
				"3152205ec4f59c74fe441065b6532231de2fac56 EQUAL",
			isP2SH: true,
			scriptInfo: ScriptInfo{
				ScriptPubKeyClass: ScriptHashTy,
				NumInputs:         3,
				ExpectedInputs:    2, // nonstandard p2sh.
				SigOps:            1,
			},
		},
		{
			name: "p2sh nonstandard script",
			sigScript: "1 81 DATA_8 2DUP EQUAL NOT VERIFY ABS " +
				"SWAP ABS EQUAL",
			scriptPubKey: "BLAKE2B DATA_32 0xfe441065b6532231de2fac56" +
				"3152205ec4f59c74fe441065b6532231de2fac56 EQUAL",
			isP2SH: true,
			scriptInfo: ScriptInfo{
				ScriptPubKeyClass: ScriptHashTy,
				NumInputs:         3,
				ExpectedInputs:    -1, // nonstandard p2sh.
				SigOps:            0,
			},
		},
	}

	for _, test := range tests {
		sigScript := mustParseShortForm(test.sigScript, 0)
		scriptPubKey := mustParseShortForm(test.scriptPubKey, 0)

		si, err := CalcScriptInfo(sigScript, scriptPubKey, test.isP2SH)
		if e := checkScriptError(err, test.scriptInfoErr); e != nil {
			t.Errorf("scriptinfo test %q: %v", test.name, e)
			continue
		}
		if err != nil {
			continue
		}

		if *si != test.scriptInfo {
			t.Errorf("%s: scriptinfo doesn't match expected. "+
				"got: %q expected %q", test.name, *si,
				test.scriptInfo)
			continue
		}
	}
}

// bogusAddress implements the util.Address interface so the tests can ensure
// unsupported address types are handled properly.
type bogusAddress struct{}

// EncodeAddress simply returns an empty string. It exists to satisfy the
// util.Address interface.
func (b *bogusAddress) EncodeAddress() string {
	return ""
}

// ScriptAddress simply returns an empty byte slice. It exists to satisfy the
// util.Address interface.
func (b *bogusAddress) ScriptAddress() []byte {
	return nil
}

// IsForPrefix lies blatantly to satisfy the util.Address interface.
func (b *bogusAddress) IsForPrefix(prefix util.Bech32Prefix) bool {
	return true // why not?
}

// String simply returns an empty string. It exists to satisfy the
// util.Address interface.
func (b *bogusAddress) String() string {
	return ""
}

func (b *bogusAddress) Prefix() util.Bech32Prefix {
	return util.Bech32PrefixUnknown
}

// TestPayToAddrScript ensures the PayToAddrScript function generates the
// correct scripts for the various types of addresses.
func TestPayToAddrScript(t *testing.T) {
	t.Parallel()

	p2pkMain, err := util.NewAddressPubKey(hexToBytes("e34cce70c86"+
		"373273efcc54ce7d2a491bb4a0e84e34cce70c86373273efcc54c"), util.Bech32PrefixKaspa)
	if err != nil {
		t.Fatalf("Unable to create public key address: %v", err)
	}

	p2shMain, err := util.NewAddressScriptHashFromHash(hexToBytes("e8c300"+
		"c87986efa84c37c0519929019ef86eb5b4e34cce70c86373273efcc54c"), util.Bech32PrefixKaspa)
	if err != nil {
		t.Fatalf("Unable to create script hash address: %v", err)
	}

	// Errors used in the tests below defined here for convenience and to
	// keep the horizontal test size shorter.
	errUnsupportedAddress := scriptError(ErrUnsupportedAddress, "")

	tests := []struct {
		in              util.Address
		expectedScript  string
		expectedVersion uint16
		err             error
	}{
		// pay-to-pubkey address on mainnet
		{
			p2pkMain,
			"DATA_32 0xe34cce70c86373273efcc54ce7d2a4" +
				"91bb4a0e84e34cce70c86373273efcc54c CHECKSIG",
			0,
			nil,
		},
		// pay-to-script-hash address on mainnet
		{
			p2shMain,
			"BLAKE2B DATA_32 0xe8c300c87986efa84c37c0519929019ef8" +
				"6eb5b4e34cce70c86373273efcc54c EQUAL",
			0,
			nil,
		},

		// Supported address types with nil pointers.
		{(*util.AddressPubKey)(nil), "", 0, errUnsupportedAddress},
		{(*util.AddressScriptHash)(nil), "", 0, errUnsupportedAddress},

		// Unsupported address type.
		{&bogusAddress{}, "", 0, errUnsupportedAddress},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		scriptPublicKey, err := PayToAddrScript(test.in)
		if e := checkScriptError(err, test.err); e != nil {
			t.Errorf("PayToAddrScript #%d unexpected error - "+
				"got %v, want %v", i, err, test.err)
			continue
		}

		var scriptPublicKeyScript []byte
		var scriptPublicKeyVersion uint16
		if scriptPublicKey != nil {
			scriptPublicKeyScript = scriptPublicKey.Script
			scriptPublicKeyVersion = scriptPublicKey.Version
		}

		expectedVersion := test.expectedVersion
		expectedScript := mustParseShortForm(test.expectedScript, test.expectedVersion)
		if !bytes.Equal(scriptPublicKeyScript, expectedScript) {
			t.Errorf("PayToAddrScript #%d got: %x\nwant: %x",
				i, scriptPublicKey, expectedScript)
			continue
		}
		if scriptPublicKeyVersion != expectedVersion {
			t.Errorf("PayToAddrScript #%d got version: %d\nwant: %d",
				i, scriptPublicKeyVersion, expectedVersion)
			continue
		}
	}
}

// scriptClassTests houses several test scripts used to ensure various class
// determination is working as expected. It's defined as a test global versus
// inside a function scope since this spans both the standard tests and the
// consensus tests (pay-to-script-hash is part of consensus).
var scriptClassTests = []struct {
	name   string
	script string
	class  ScriptClass
}{
	// p2pk
	{
		name:   "Pay Pubkey",
		script: "DATA_32 0x89ac24ea10bb751af4939623ccc5e550d96842b64e8fca0f63e94b4373fd555e CHECKSIG",
		class:  PubKeyTy,
	},
	{
		name: "Pay PubkeyHash",
		script: "DUP BLAKE2B DATA_32 0x660d4ef3a743e3e696ad990364e55543e3e696ad990364e555e555" +
			"c271ad504b EQUALVERIFY CHECKSIG",
		class: NonStandardTy,
	},
	// mutlisig
	{
		name: "multisig",
		script: "1 DATA_33 0x0232abdc893e7f0631364d7fd01cb33d24da4" +
			"5329a00357b3a7886211ab414d55a 1 CHECKMULTISIG",
		class: NonStandardTy,
	},
	{
		name: "P2SH",
		script: "BLAKE2B DATA_32 0x433ec2ac1ffa1b7b7d027f564529c57197fa1b7b7d027f564529c57197f" +
			"9ae88 EQUAL",
		class: ScriptHashTy,
	},

	{
		// Nulldata. It is standard in Bitcoin but not in Kaspa
		name:   "nulldata",
		script: "RETURN 0",
		class:  NonStandardTy,
	},

	// The next few are almost multisig (it is the more complex script type)
	// but with various changes to make it fail.
	{
		// Multisig but invalid nsigs.
		name: "strange 1",
		script: "DUP DATA_33 0x0232abdc893e7f0631364d7fd01cb33d24da45" +
			"329a00357b3a7886211ab414d55a 1 CHECKMULTISIG",
		class: NonStandardTy,
	},
	{
		// Multisig but invalid pubkey.
		name:   "strange 2",
		script: "1 1 1 CHECKMULTISIG",
		class:  NonStandardTy,
	},
	{
		// Multisig but no matching npubkeys opcode.
		name: "strange 3",
		script: "1 DATA_33 0x0232abdc893e7f0631364d7fd01cb33d24da4532" +
			"9a00357b3a7886211ab414d55a DATA_33 0x0232abdc893e7f0" +
			"631364d7fd01cb33d24da45329a00357b3a7886211ab414d55a " +
			"CHECKMULTISIG",
		class: NonStandardTy,
	},
	{
		// Multisig but with multisigverify.
		name: "strange 4",
		script: "1 DATA_33 0x0232abdc893e7f0631364d7fd01cb33d24da4532" +
			"9a00357b3a7886211ab414d55a 1 CHECKMULTISIGVERIFY",
		class: NonStandardTy,
	},
	{
		// Multisig but wrong length.
		name:   "strange 5",
		script: "1 CHECKMULTISIG",
		class:  NonStandardTy,
	},
	{
		name:   "doesn't parse",
		script: "DATA_5 0x01020304",
		class:  NonStandardTy,
	},
	{
		name: "multisig script with wrong number of pubkeys",
		script: "2 " +
			"DATA_33 " +
			"0x027adf5df7c965a2d46203c781bd4dd8" +
			"21f11844136f6673af7cc5a4a05cd29380 " +
			"DATA_33 " +
			"0x02c08f3de8ee2de9be7bd770f4c10eb0" +
			"d6ff1dd81ee96eedd3a9d4aeaf86695e80 " +
			"3 CHECKMULTISIG",
		class: NonStandardTy,
	},
}

// TestScriptClass ensures all the scripts in scriptClassTests have the expected
// class.
func TestScriptClass(t *testing.T) {
	t.Parallel()

	for _, test := range scriptClassTests {
		script := mustParseShortForm(test.script, 0)
		class := GetScriptClass(script)
		if class != test.class {
			t.Errorf("%s: expected %s got %s (script %x)", test.name,
				test.class, class, script)
			continue
		}
	}
}

// TestStringifyClass ensures the script class string returns the expected
// string for each script class.
func TestStringifyClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		class    ScriptClass
		stringed string
	}{
		{
			name:     "nonstandardty",
			class:    NonStandardTy,
			stringed: "nonstandard",
		},
		{
			name:     "pubkey",
			class:    PubKeyTy,
			stringed: "pubkey",
		},
		{
			name:     "scripthash",
			class:    ScriptHashTy,
			stringed: "scripthash",
		},
		{
			name:     "broken",
			class:    ScriptClass(255),
			stringed: "Invalid",
		},
	}

	for _, test := range tests {
		typeString := test.class.String()
		if typeString != test.stringed {
			t.Errorf("%s: got %#q, want %#q", test.name,
				typeString, test.stringed)
		}
	}
}
