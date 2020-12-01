package ibd

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestMaxHeaders(t *testing.T) {
	testutils.ForAllNets(t, false, func(t *testing.T, params *dagconfig.Params) {
		if params.FinalityDepth() > maxHeaders {
			t.Errorf("FinalityDepth() in %s should be lower or equal to appmessage.MaxInvPerMsg", params.Name)
		}
	})
}
