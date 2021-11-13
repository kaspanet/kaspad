package pow

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"math"
)

const eps float64 = 1e-9

type matrix [64][64]uint16

func generateMatrix(hash *externalapi.DomainHash) *matrix {
	var mat matrix
	generator := newxoShiRo256PlusPlus(hash)
	for {
		for i := range mat {
			for j := 0; j < 64; j += 16 {
				val := generator.Uint64()
				for shift := 0; shift < 16; shift++ {
					mat[i][j+shift] = uint16(val >> (4 * shift) & 0x0F)
				}
			}
		}
		if mat.computeRank() == 64 {
			return &mat
		}
	}
}

func (mat *matrix) computeRank() int {
	var B [64][64]float64
	for i := range B {
		for j := range B[0] {
			B[i][j] = float64(mat[i][j])
		}
	}
	var rank int
	var rowSelected [64]bool
	for i := 0; i < 64; i++ {
		var j int
		for j = 0; j < 64; j++ {
			if !rowSelected[j] && math.Abs(B[j][i]) > eps {
				break
			}
		}
		if j != 64 {
			rank++
			rowSelected[j] = true
			for p := i + 1; p < 64; p++ {
				B[j][p] /= B[j][i]
			}
			for k := 0; k < 64; k++ {
				if k != j && math.Abs(B[k][i]) > eps {
					for p := i + 1; p < 64; p++ {
						B[k][p] -= B[j][p] * B[k][i]
					}
				}
			}
		}
	}
	return rank
}

func (mat *matrix) HeavyHash(hash *externalapi.DomainHash) *externalapi.DomainHash {
	hashBytes := hash.ByteArray()
	var vector [64]uint16
	var product [64]uint16
	for i := 0; i < 32; i++ {
		vector[2*i] = uint16(hashBytes[i] >> 4)
		vector[2*i+1] = uint16(hashBytes[i] & 0x0F)
	}
	// Matrix-vector multiplication, and convert to 4 bits.
	for i := 0; i < 64; i++ {
		var sum uint16
		for j := 0; j < 64; j++ {
			sum += mat[i][j] * vector[j]
		}
		product[i] = sum >> 10
	}

	// Concatenate 4 LSBs back to 8 bit xor with sum1
	var res [32]byte
	for i := range res {
		res[i] = hashBytes[i] ^ (byte(product[2*i]<<4) | byte(product[2*i+1]))
	}
	// Hash again
	writer := hashes.NewHeavyHashWriter()
	writer.InfallibleWrite(res[:])
	return writer.Finalize()
}
