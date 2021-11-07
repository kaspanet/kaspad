package pow

import "testing"

// Test vectors are from here: https://github.com/rust-random/rngs/blob/17aa826cc38d3e8408c9489ac859fa9397acd479/rand_xoshiro/src/xoshiro256plusplus.rs#L121
func TestXoShiRo256PlusPlus_Uint64(t *testing.T) {
	state := xoShiRo256PlusPlus{1, 2, 3, 4}
	expected := []uint64{41943041, 58720359, 3588806011781223, 3591011842654386,
		9228616714210784205, 9973669472204895162, 14011001112246962877,
		12406186145184390807, 15849039046786891736, 10450023813501588000}
	for _, ex := range expected {
		val := state.Uint64()
		if val != ex {
			t.Errorf("expected: %d, found: %d", ex, val)
		}
	}
}
