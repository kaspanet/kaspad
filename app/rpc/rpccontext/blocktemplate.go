package rpccontext

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/mining"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"sync"
)

// BlockTemplateGenerator houses state that is used in between multiple RPC invocations to
// getBlockTemplate.
type BlockTemplateGenerator struct {
	sync.Mutex

	context *Context

	lastTxUpdate  mstime.Time
	lastGenerated mstime.Time
	tipHashes     []*daghash.Hash
	minTimestamp  mstime.Time
	template      *mining.BlockTemplate
	notifyMap     map[string]map[int64]chan struct{}
	payAddress    util.Address
}

// NewBlockTemplateGenerator returns a new instance of a BlockTemplateGenerator with all internal
// fields initialized and ready to use.
func NewBlockTemplateGenerator(context *Context) *BlockTemplateGenerator {
	return &BlockTemplateGenerator{
		context:   context,
		notifyMap: make(map[string]map[int64]chan struct{}),
	}
}

func (bt *BlockTemplateGenerator) Update(payAddress util.Address) error {
	return nil
}

func (bt *BlockTemplateGenerator) Response() *appmessage.GetBlockTemplateResponseMessage {
	return appmessage.NewGetBlockTemplateResponseMessage()
}
