package mining

import (
	"fmt"

	"github.com/kaspanet/kaspad/util"
)

type testTxDescDefinition struct {
	fee  uint64
	mass uint64
	gas  uint64

	expectedMinSelectedTimes uint64
	expectedMaxSelectedTimes uint64

	tx *util.Tx
}

func (dd testTxDescDefinition) String() string {
	return fmt.Sprintf("[fee: %d, gas: %d, mass: %d]", dd.fee, dd.gas, dd.mass)
}
