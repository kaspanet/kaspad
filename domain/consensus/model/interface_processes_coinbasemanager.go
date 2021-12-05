package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// CoinbaseManager exposes methods for handling blocks'
// coinbase transactions
type CoinbaseManager interface {
	ExpectedCoinbaseTransaction(stagingArea *StagingArea, blockHash *externalapi.DomainHash,
		coinbaseData *externalapi.DomainCoinbaseData, blockPruningPoint *externalapi.DomainHash) (*externalapi.DomainTransaction, error)
	CalcBlockSubsidy(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (uint64, error)
	ExtractCoinbaseDataBlueScoreAndSubsidy(coinbaseTx *externalapi.DomainTransaction) (blueScore uint64, coinbaseData *externalapi.DomainCoinbaseData, subsidy uint64, err error)
}
