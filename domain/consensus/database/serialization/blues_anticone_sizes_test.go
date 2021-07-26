package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"testing"
)

// TestBlueAnticoneSizesSize tests that no data can be loss when converting externalapi.KType to the corresponding type in
// DbBluesAnticoneSizes
func TestKType(t *testing.T) {
	k := externalapi.KType(0)
	k--

	if k < externalapi.KType(0) {
		t.Fatalf("KType must be unsigned")
	}

	// Setting maxKType to maximum value of KType.
	// As we verify above that KType is unsigned we can be sure that maxKType is indeed the maximum value of KType.
	maxKType := ^externalapi.KType(0)
	dbBluesAnticoneSizes := DbBluesAnticoneSizes{
		AnticoneSize: uint32(maxKType),
	}
	if externalapi.KType(dbBluesAnticoneSizes.AnticoneSize) != maxKType {
		t.Fatalf("convert from uint32 to KType losses data")
	}
}
