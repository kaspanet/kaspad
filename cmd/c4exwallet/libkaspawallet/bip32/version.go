package bip32

import "github.com/pkg/errors"

// BitcoinMainnetPrivate is the version that is used for
// bitcoin mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var BitcoinMainnetPrivate = [4]byte{
	0x04,
	0x88,
	0xad,
	0xe4,
}

// BitcoinMainnetPublic is the version that is used for
// bitcoin mainnet bip32 public extended keys.
// Ecnodes to xpub in base58.
var BitcoinMainnetPublic = [4]byte{
	0x04,
	0x88,
	0xb2,
	0x1e,
}

// C4exMainnetPrivate is the version that is used for
// c4ex mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var C4exMainnetPrivate = [4]byte{
	0x03,
	0x8f,
	0x2e,
	0xf4,
}

// C4exMainnetPublic is the version that is used for
// c4ex mainnet bip32 public extended keys.
// Ecnodes to kpub in base58.
var C4exMainnetPublic = [4]byte{
	0x03,
	0x8f,
	0x33,
	0x2e,
}

// C4exTestnetPrivate is the version that is used for
// c4ex testnet bip32 public extended keys.
// Ecnodes to ktrv in base58.
var C4exTestnetPrivate = [4]byte{
	0x03,
	0x90,
	0x9e,
	0x07,
}

// C4exTestnetPublic is the version that is used for
// c4ex testnet bip32 public extended keys.
// Ecnodes to ktub in base58.
var C4exTestnetPublic = [4]byte{
	0x03,
	0x90,
	0xa2,
	0x41,
}

// C4exDevnetPrivate is the version that is used for
// c4ex devnet bip32 public extended keys.
// Ecnodes to kdrv in base58.
var C4exDevnetPrivate = [4]byte{
	0x03,
	0x8b,
	0x3d,
	0x80,
}

// C4exDevnetPublic is the version that is used for
// c4ex devnet bip32 public extended keys.
// Ecnodes to xdub in base58.
var C4exDevnetPublic = [4]byte{
	0x03,
	0x8b,
	0x41,
	0xba,
}

// C4exSimnetPrivate is the version that is used for
// c4ex simnet bip32 public extended keys.
// Ecnodes to ksrv in base58.
var C4exSimnetPrivate = [4]byte{
	0x03,
	0x90,
	0x42,
	0x42,
}

// C4exSimnetPublic is the version that is used for
// c4ex simnet bip32 public extended keys.
// Ecnodes to xsub in base58.
var C4exSimnetPublic = [4]byte{
	0x03,
	0x90,
	0x46,
	0x7d,
}

func toPublicVersion(version [4]byte) ([4]byte, error) {
	switch version {
	case BitcoinMainnetPrivate:
		return BitcoinMainnetPublic, nil
	case C4exMainnetPrivate:
		return C4exMainnetPublic, nil
	case C4exTestnetPrivate:
		return C4exTestnetPublic, nil
	case C4exDevnetPrivate:
		return C4exDevnetPublic, nil
	case C4exSimnetPrivate:
		return C4exSimnetPublic, nil
	}

	return [4]byte{}, errors.Errorf("unknown version %x", version)
}

func isPrivateVersion(version [4]byte) bool {
	switch version {
	case BitcoinMainnetPrivate:
		return true
	case C4exMainnetPrivate:
		return true
	case C4exTestnetPrivate:
		return true
	case C4exDevnetPrivate:
		return true
	case C4exSimnetPrivate:
		return true
	}

	return false
}
