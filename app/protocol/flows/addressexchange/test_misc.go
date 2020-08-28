package addressexchange

import (
	"math/rand"
	"net"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/util"
)

var (
	calcSequenceFunc = func(tx *util.Tx, utxoSet blockdag.UTXOSet) (*blockdag.SequenceLock, error) {
		return &blockdag.SequenceLock{
			Milliseconds:   -1,
			BlockBlueScore: -1,
		}, nil
	}

	defaultCfg = config.DefaultConfig()
	dagCfg     = &blockdag.Config{
		DAGParams:  &dagconfig.SimnetParams,
		TimeSource: blockdag.NewTimeSource(),
		SigCache:   txscript.NewSigCache(1000),
	}
	memPoolCfg = &mempool.Config{
		Policy: mempool.Policy{
			MaxOrphanTxs:    5,
			MaxOrphanTxSize: 1000,
			MinRelayTxFee:   1000,
			MaxTxVersion:    1,
		},
		CalcSequenceLockNoLock: calcSequenceFunc,
		SigCache:               nil,
	}
)

func generateIPForTest() net.IP {
	ip := make([]byte, 4)
	rand.Read(ip)
	return ip
}

func generateAddressesForTest(number int) []*appmessage.NetAddress {
	addresses := make([]*appmessage.NetAddress, number)
	for i := 0; i < number; i++ {
		addresses[i] = appmessage.NewNetAddressIPPort(
			generateIPForTest(),
			uint16(rand.Intn(65536)),
			appmessage.SFNodeNetwork)
	}
	return addresses
}
