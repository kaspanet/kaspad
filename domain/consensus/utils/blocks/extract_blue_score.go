package blocks

import (
	"errors"

	"github.com/kaspanet/kaspad/domain/consensus/utils/coinbase"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

// ExtractBlueScore extracts the block's blue score out of it's coinbase transaction's payload
func ExtractBlueScore(block *externalapi.DomainBlock) (uint64, error) {
	if len(block.Transactions) < transactionhelper.CoinbaseTransactionIndex+1 {
		return 0, errors.New("Block has no coinbase transaction")
	}

	coinbaseTransaction := block.Transactions[transactionhelper.CoinbaseTransactionIndex]

	blueScore, _, err := coinbase.ExtractCoinbaseDataAndBlueScore(coinbaseTransaction)
	return blueScore, err
}
