package pow

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/bits"
)

type xoShiRo256PlusPlus struct {
	s0 uint64
	s1 uint64
	s2 uint64
	s3 uint64
}

func newxoShiRo256PlusPlus(hash *externalapi.DomainHash) *xoShiRo256PlusPlus {
	hashArray := hash.ByteArray()
	return &xoShiRo256PlusPlus{
		s0: binary.LittleEndian.Uint64(hashArray[:8]),
		s1: binary.LittleEndian.Uint64(hashArray[8:16]),
		s2: binary.LittleEndian.Uint64(hashArray[16:24]),
		s3: binary.LittleEndian.Uint64(hashArray[24:32]),
	}
}

func (x *xoShiRo256PlusPlus) Uint64() uint64 {
	res := bits.RotateLeft64(x.s0+x.s3, 23) + x.s0
	t := x.s1 << 17
	x.s2 ^= x.s0
	x.s3 ^= x.s1
	x.s1 ^= x.s2
	x.s0 ^= x.s3

	x.s2 ^= t
	x.s3 = bits.RotateLeft64(x.s3, 45)
	return res
}
