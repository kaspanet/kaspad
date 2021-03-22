package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetBlockDAGInfo handles the respectively named RPC command
func HandleGetBlockDAGInfo(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	params := context.Config.ActiveNetParams
	consensus := context.Domain.Consensus()

	response := appmessage.NewGetBlockDAGInfoResponseMessage()
	response.NetworkName = params.Name

	syncInfo, err := consensus.GetSyncInfo(nil)
	if err != nil {
		return nil, err
	}
	response.BlockCount = syncInfo.BlockCount
	response.HeaderCount = syncInfo.HeaderCount

	tipHashes, err := consensus.Tips()
	if err != nil {
		return nil, err
	}
	response.TipHashes = hashes.ToStrings(tipHashes)

	virtualInfo, err := consensus.GetVirtualInfo()
	if err != nil {
		return nil, err
	}
	response.VirtualParentHashes = hashes.ToStrings(virtualInfo.ParentHashes)
	response.Difficulty = context.GetDifficultyRatio(virtualInfo.Bits, context.Config.ActiveNetParams)
	response.PastMedianTime = virtualInfo.PastMedianTime

	pruningPoint, err := context.Domain.Consensus().PruningPoint()
	if err != nil {
		return nil, err
	}
	response.PruningPointHash = pruningPoint.String()

	return response, nil
}
