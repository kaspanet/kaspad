package bip32

import "github.com/pkg/errors"

var BitcoinMainnetPrivate = [4]byte{
	0x04,
	0x88,
	0xad,
	0xe4,
}

var BitcoinMainnetPublic = [4]byte{
	0x04,
	0x88,
	0xb2,
	0x1e,
}

var KaspaMainnetPrivate = [4]byte{
	0x01,
	0x02,
	0x03,
	0x04,
}

var KaspaMainnetPublic = [4]byte{
	0x01,
	0x02,
	0xfe,
	0xff,
}

func toPublicVersion(version [4]byte) ([4]byte, error) {
	switch version {
	case BitcoinMainnetPrivate:
		return BitcoinMainnetPublic, nil
	case KaspaMainnetPrivate:
		return KaspaMainnetPublic, nil
	}

	return [4]byte{}, errors.Errorf("unknown version %x", version)
}
