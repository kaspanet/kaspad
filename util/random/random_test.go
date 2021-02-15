package random

import (
	"crypto/rand"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"testing"
)

// fakeRandReader implements the io.Reader interface and is used to force
// errors in the RandomUint64 function.
type fakeRandReader struct {
	n   int
	err error
}

// Read returns the fake reader error and the lesser of the fake reader value
// and the length of p.
func (r *fakeRandReader) Read(p []byte) (int, error) {
	n := r.n
	if n > len(p) {
		n = len(p)
	}
	return n, r.err
}

// TestRandomUint64 exercises the randomness of the random number generator on
// the system by ensuring the probability of the generated numbers. If the RNG
// is evenly distributed as a proper cryptographic RNG should be, there really
// should only be 1 number < 2^56 in 2^8 tries for a 64-bit number. However,
// use a higher number of 5 to really ensure the test doesn't fail unless the
// RNG is just horrendous.
func TestRandomUint64(t *testing.T) {
	tries := 1 << 8              // 2^8
	watermark := uint64(1 << 56) // 2^56
	maxHits := 5

	numHits := 0
	for i := 0; i < tries; i++ {
		nonce, err := Uint64()
		if err != nil {
			t.Errorf("RandomUint64 iteration %d failed - err %v",
				i, err)
			return
		}
		if nonce < watermark {
			numHits++
		}
		if numHits > maxHits {
			str := fmt.Sprintf("The random number generator on this system is clearly "+
				"terrible since we got %d values less than %d in %d runs "+
				"when only %d was expected", numHits, watermark, tries, maxHits)
			t.Errorf("Random Uint64 iteration %d failed - %v %v", i,
				str, numHits)
			return
		}
	}
}

// TestRandomUint64Errors uses a fake reader to force error paths to be executed
// and checks the results accordingly.
func TestRandomUint64Errors(t *testing.T) {
	// Test short reads.
	reader := rand.Reader
	rand.Reader = &fakeRandReader{n: 2, err: io.EOF}
	nonce, err := Uint64()
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("Error not expected value of %v [%v]",
			io.ErrUnexpectedEOF, err)
	}
	if nonce != 0 {
		t.Errorf("Nonce is not 0 [%v]", nonce)
	}
	rand.Reader = reader
}
