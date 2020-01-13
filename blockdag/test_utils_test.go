package blockdag

import (
	"testing"
)

func TestIsSupportedDbType(t *testing.T) {
	if !isSupportedDbType("ffldb") {
		t.Errorf("ffldb should be a supported DB driver")
	}
	if isSupportedDbType("madeUpDb") {
		t.Errorf("madeUpDb should not be a supported DB driver")
	}
}
