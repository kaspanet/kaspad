package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// CoinbaseManager exposes methods for handling blocks'
// coinbase transactions
type CoinbaseManager interface {
	ExpectedCoinbaseTransaction(blockHash *externalapi.DomainHash,
		coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransaction, error)
	ValidateCoinbaseTransactionInContext(blockHash *externalapi.DomainHash, coinbaseTransaction *externalapi.DomainTransaction) error
	ValidateCoinbaseTransactionInIsolation(coinbaseTransaction *externalapi.DomainTransaction) error
}
