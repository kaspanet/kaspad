package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/daghash"
)

// HandleGetBlockDAGInfo handles the respectively named RPC command
func HandleGetBlockDAGInfo(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	dag := context.DAG
	params := dag.Params

	response := appmessage.NewGetBlockDAGInfoResponseMessage()
	response.NetworkName = params.Name
	response.BlockCount = dag.BlockCount()
	response.TipHashes = daghash.Strings(dag.TipHashes())
	response.Difficulty = context.GetDifficultyRatio(dag.CurrentBits(), params)
	response.PastMedianTime = dag.CalcPastMedianTime().UnixMilliseconds()
	return response, nil
}
