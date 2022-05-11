package server

import (
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

type reservedOutpoints map[externalapi.DomainOutpoint]int64
type sentTransactions map[string][]*externalapi.DomainOutpoint

//Tracker tracks wallet server utxos via outpoints.
type Tracker struct {
	reservedOutpoints        reservedOutpoints
	sentTransactions         sentTransactions
	expiryDurationInSecounds int64
}

//NewTracker intializes and returns a new Tracker
func NewTracker() *Tracker {
	return &Tracker{
		reservedOutpoints:        make(reservedOutpoints),
		sentTransactions:         make(sentTransactions),
		expiryDurationInSecounds: 14, // TO DO: better expiry mechanisim 14 secounds current corrosponds to about two ticks in server sync.
	}
}

func (tr *Tracker) isOutpointAvailable(outpoint *externalapi.DomainOutpoint) bool {
	if tr.isOutpointReserved(outpoint) || tr.isOutpointSent(outpoint) {
		return false
	}

	return true
}

func (tr *Tracker) isOutpointSent(outpoint *externalapi.DomainOutpoint) bool {
	for _, outpoints := range tr.sentTransactions {
		for _, sentOutpoint := range outpoints {
			if outpoint.Equal(sentOutpoint) {
				return true
			}
		}
	}
	return false
}

func (tr *Tracker) untrackOutpointDifferenceViaWalletUTXOs(utxos []*walletUTXO) {

	validOutpoints := make(map[externalapi.DomainOutpoint]bool, len(utxos))
	for _, utxo := range utxos {
		validOutpoints[*utxo.Outpoint] = true
	}
	for reservedOutpoint := range tr.reservedOutpoints {
		if _, found := validOutpoints[reservedOutpoint]; !found {
			delete(tr.reservedOutpoints, reservedOutpoint)
		}
	}
}

func (tr *Tracker) isOutpointReserved(outpoint *externalapi.DomainOutpoint) bool {
	_, found := tr.reservedOutpoints[*outpoint]
	return found
}

func (tr *Tracker) isTransactionIDTracked(transactionID string) bool {
	_, found := tr.sentTransactions[transactionID]
	return found
}

func (tr *Tracker) untrackExpiredOutpointsAsReserved() {
	currentTimestamp := time.Now().Unix()
	for outpoint, reserveTimestamp := range tr.reservedOutpoints {
		if currentTimestamp-reserveTimestamp >= tr.expiryDurationInSecounds {
			delete(tr.reservedOutpoints, outpoint)
		}

	}
}

func (tr *Tracker) untrackSentTransactionID(transactionID string) {
	delete(tr.sentTransactions, transactionID)
}

func (tr *Tracker) trackOutpointAsReserved(outpoint *externalapi.DomainOutpoint) {
	tr.reservedOutpoints[*outpoint] = time.Now().Unix()
}

func (tr *Tracker) trackTransaction(transaction *externalapi.DomainTransaction) {
	transactionOutpoints := make([]*externalapi.DomainOutpoint, len(transaction.Inputs))
	for i, transactionInput := range transaction.Inputs {
		transactionOutpoints[i] = &transactionInput.PreviousOutpoint
	}
	tr.sentTransactions[consensushashing.TransactionID(transaction).String()] = transactionOutpoints
}

func (tr *Tracker) untrackOutpointAsReserved(outpoint externalapi.DomainOutpoint) {
	delete(tr.reservedOutpoints, outpoint)
}

func (tr *Tracker) countOutpointsInmempool() int {
	numOfOutpoints := 0
	for _, value := range tr.sentTransactions {
		numOfOutpoints = numOfOutpoints + len(value)
	}
	return numOfOutpoints
}