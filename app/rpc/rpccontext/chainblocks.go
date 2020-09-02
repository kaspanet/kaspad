package rpccontext

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

func (ctx *Context) CollectChainBlocks(hashes []*daghash.Hash) ([]*appmessage.ChainChangedChainBlock, error) {
	chainBlocks := make([]*appmessage.ChainChangedChainBlock, 0, len(hashes))
	for _, hash := range hashes {
		acceptanceData, err := ctx.AcceptanceIndex.TxsAcceptanceData(hash)
		if err != nil {
			return nil, errors.Errorf("could not retrieve acceptance data for block %s", hash)
		}

		acceptedBlocks := make([]*appmessage.ChainChangedAcceptedBlock, 0, len(acceptanceData))
		for _, blockAcceptanceData := range acceptanceData {
			acceptedTxIds := make([]string, 0, len(blockAcceptanceData.TxAcceptanceData))
			for _, txAcceptanceData := range blockAcceptanceData.TxAcceptanceData {
				if txAcceptanceData.IsAccepted {
					acceptedTxIds = append(acceptedTxIds, txAcceptanceData.Tx.ID().String())
				}
			}
			acceptedBlock := &appmessage.ChainChangedAcceptedBlock{
				Hash:          blockAcceptanceData.BlockHash.String(),
				AcceptedTxIDs: acceptedTxIds,
			}
			acceptedBlocks = append(acceptedBlocks, acceptedBlock)
		}

		chainBlock := &appmessage.ChainChangedChainBlock{
			Hash:           hash.String(),
			AcceptedBlocks: acceptedBlocks,
		}
		chainBlocks = append(chainBlocks, chainBlock)
	}
	return chainBlocks, nil
}
