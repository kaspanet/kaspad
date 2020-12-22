package bech32

import (
	"strings"
	"testing"
)

func TestBech32(t *testing.T) {
	tests := []struct {
		str   string
		valid bool
	}{
		{"prefix:x64nx6hz", true},
		{"p:gpf8m4h7", true},
		{"bitcoincash:qpzry9x8gf2tvdw0s3jn54khce6mua7lcw20ayyn", true},
		{"bchtest:testnetaddress4d6njnut", true},
		{"bchreg:555555555555555555555555555555555555555555555udxmlmrz", true},
		{"A:3X3DXU9W", true},
		{"an83characterlonghumanreadablepartthatcontainscharctr:andtheexcludedcharactersbio:pk68j20a", true},
		{"abcdef:qpzry9x8gf2tvdw0s3jn54khce6mua7:nw2t26kg", true},
		{"::qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq40ku0e3z", true},
		{"split:checkupstagehandshakeupstreamerranterredcaperred3za27wc5", true},
		{"aaa:bbb", false}, // too short
		{"split:checkupstagehandshakeupstreamerranterredCaperred3za27wc5", false},                         // mixed uppercase and lowercase
		{"split:checkupstagehandshakeupstreamerranterredcaperred3za28wc5", false},                         // invalid checksum
		{"s lit:checkupstagehandshakeupstreamerranterredcaperred3za27wc5", false},                         // invalid character (space) in prefix
		{"spl" + string(rune(127)) + "t:checkupstagehandshakeupstreamerranterredcaperred3za27wc5", false}, // invalid character (DEL) in prefix
		{"split:cheosgds2s3c", false}, // invalid character (o) in data part
		{"split:te5peu7", false},      // too short data part
		{":checkupstagehandshakeupstreamerranterredcaperred3za27wc5", false},                                   // empty prefix
		{"::qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq40ku0e3z", false}, // too long
		{"bitcoincash:qr6m7j9njldwwzlg9v7v53unlr4jkmx6eylep8ekg2", true},
		{"bchtest:pr6m7j9njldwwzlg9v7v53unlr4jkmx6eyvwc0uz5t", true},
		{"prefix:0r6m7j9njldwwzlg9v7v53unlr4jkmx6ey3qnjwsrf", true},
	}

	for _, test := range tests {
		str := test.str
		prefix, decoded, err := decode(str)
		if !test.valid {
			// Invalid string decoding should result in error.
			if err == nil {
				t.Error("expected decoding to fail for "+
					"invalid string %v", test.str)
			}
			continue
		}

		// Valid string decoding should result in no error.
		if err != nil {
			t.Errorf("expected string to be valid bech32: %v", err)
		}

		// Check that it encodes to the same string
		encoded := encode(prefix, decoded)
		if encoded != strings.ToLower(str) {
			t.Errorf("expected data to encode to %v, but got %v",
				str, encoded)
		}

		// Flip a bit in the string an make sure it is caught.
		pos := strings.LastIndexAny(str, "1")
		flipped := str[:pos+1] + string((str[pos+1] ^ 1)) + str[pos+2:]
		_, _, err = decode(flipped)
		if err == nil {
			t.Error("expected decoding to fail")
		}
	}
}

func TestEncodeToBech32NotUInt5(t *testing.T) {
	encoded := encodeToBase32([]byte("™"))
	if encoded != "" {
		t.Errorf("encodeToBase32 unexpectedly succeeded")
	}
}
