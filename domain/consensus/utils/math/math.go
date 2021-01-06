package math

import (
	"math/big"
)

var (
	// bigOne is 1 represented as a big.Int. It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// oneLsh256 is 1 shifted left 256 bits. It is defined here to avoid
	// the overhead of creating it multiple times.
	oneLsh256 = new(big.Int).Lsh(bigOne, 256)

	// log2FloorMasks defines the masks to use when quickly calculating
	// floor(log2(x)) in a constant log2(64) = 6 steps, where x is a uint64, using
	// shifts. They are derived from (2^(2^x) - 1) * (2^(2^x)), for x in 5..0.
	log2FloorMasks = []uint64{0xffffffff00000000, 0xffff0000, 0xff00, 0xf0, 0xc, 0x2}
)

// FastLog2Floor calculates and returns floor(log2(x)) in a constant 5 steps.
func FastLog2Floor(n uint64) uint8 {
	rv := uint8(0)
	exponent := uint8(32)
	for i := 0; i < 6; i++ {
		if n&log2FloorMasks[i] != 0 {
			rv += exponent
			n >>= exponent
		}
		exponent >>= 1
	}
	return rv
}

// CompactToBig converts a compact representation of a whole number N to an
// unsigned 32-bit number. The representation is similar to IEEE754 floating
// point numbers.
//
// Like IEEE754 floating point, there are three basic components: the sign,
// the exponent, and the mantissa. They are broken out as follows:
//
//	* the most significant 8 bits represent the unsigned base 256 exponent
// 	* bit 23 (the 24th bit) represents the sign bit
//	* the least significant 23 bits represent the mantissa
//
//	-------------------------------------------------
//	|   Exponent     |    Sign    |    Mantissa     |
//	-------------------------------------------------
//	| 8 bits [31-24] | 1 bit [23] | 23 bits [22-00] |
//	-------------------------------------------------
//
// The formula to calculate N is:
// 	N = (-1^sign) * mantissa * 256^(exponent-3)
func CompactToBig(compact uint32) *big.Int {
	destination := big.NewInt(0)
	CompactToBigWithDestination(compact, destination)
	return destination
}

// CompactToBigWithDestination is a version of CompactToBig that
// takes a destination parameter. This is useful for saving memory,
// as then the destination big.Int can be reused.
// See CompactToBig for further details.
func CompactToBigWithDestination(compact uint32, destination *big.Int) {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number. So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly. This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		destination.SetInt64(int64(mantissa))
	} else {
		destination.SetInt64(int64(mantissa))
		destination.Lsh(destination, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		destination.Neg(destination)
	}
}

// BigToCompact converts a whole number N to a compact representation using
// an unsigned 32-bit number. The compact representation only provides 23 bits
// of precision, so values larger than (2^23 - 1) only encode the most
// significant digits of the number. See CompactToBig for details.
func BigToCompact(n *big.Int) uint32 {
	// No need to do any work if it's zero.
	if n.Sign() == 0 {
		return 0
	}

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes. So, shift the number right or left
	// accordingly. This is equivalent to:
	// mantissa = mantissa / 256^(exponent-3)
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		// Use a copy to avoid modifying the caller's original number.
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	// When the mantissa already has the sign bit set, the number is too
	// large to fit into the available 23-bits, so divide the number by 256
	// and increment the exponent accordingly.
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	// Pack the exponent, sign bit, and mantissa into an unsigned 32-bit
	// int and return it.
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

// CalcWork calculates a work value from difficulty bits. Kaspa increases
// the difficulty for generating a block by decreasing the value which the
// generated hash must be less than. This difficulty target is stored in each
// block header using a compact representation as described in the documentation
// for CompactToBig. Since a lower target difficulty value equates to higher
// actual difficulty, the work value which will be accumulated must be the
// inverse of the difficulty. Also, in order to avoid potential division by
// zero and really small floating point numbers, the result adds 1 to the
// denominator and multiplies the numerator by 2^256.
func CalcWork(bits uint32) *big.Int {
	// Return a work value of zero if the passed difficulty bits represent
	// a negative number. Note this should not happen in practice with valid
	// blocks, but an invalid block could trigger it.
	difficultyNum := CompactToBig(bits)
	if difficultyNum.Sign() <= 0 {
		return big.NewInt(0)
	}

	// (1 << 256) / (difficultyNum + 1)
	denominator := new(big.Int).Add(difficultyNum, bigOne)
	return new(big.Int).Div(oneLsh256, denominator)
}
