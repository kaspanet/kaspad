package server

import (
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

type reservedOutpoints map[externalapi.DomainOutpoint]int64
type mempoolOutpoints map[externalapi.DomainOutpoint]bool
type sentTransactions map[externalapi.DomainTransactionID]bool

//Tracker tracks wallet server utxos via outpoints.
type Tracker struct {
	reservedOutpoints        reservedOutpoints
	mempoolOutpoints         mempoolOutpoints
	sentTransactions         sentTransactions
	expiryDurationInSecounds int64
}

//NewTracker intializes and returns a new Tracker
func NewTracker() *Tracker {
	return &Tracker{
		reservedOutpoints:        make(reservedOutpoints),
		mempoolOutpoints:         make(mempoolOutpoints),
		sentTransactions:         make(sentTransactions),
		expiryDurationInSecounds: 1, // 1 corrosponds to a sync ticker interval, as well as current average BPS
	}
}

func (tr *Tracker) isOutpointAvailable(outpoint *externalapi.DomainOutpoint) bool {
	if tr.isOutpointReserved(outpoint) || tr.isOutpointInMempool(outpoint) {
		return false
	}

	return true
}

func (tr *Tracker) isTransactionTracked(transaction *externalapi.DomainTransaction) bool {
	_, found := tr.sentTransactions[*consensushashing.TransactionID(transaction)]
	return found
}

func (tr *Tracker) isOutpointInMempool(outpoint *externalapi.DomainOutpoint) bool {
	_, found := tr.mempoolOutpoints[*outpoint]
	return found
}

func (tr *Tracker) isOutpointReserved(outpoint *externalapi.DomainOutpoint) bool {
	_, found := tr.reservedOutpoints[*outpoint]
	return found
}

func (tr *Tracker) untrackExpiredOutpointsAsReserved() {
	currentTimestamp := time.Now().Unix()
	for outpoint, reserveTimestamp := range tr.reservedOutpoints {
		if currentTimestamp-reserveTimestamp >= tr.expiryDurationInSecounds {
			delete(tr.reservedOutpoints, outpoint)
		}

	}
	for outpoint, sentTimestamp := range tr.reservedOutpoints {
		if currentTimestamp-sentTimestamp >= tr.expiryDurationInSecounds {
			delete(tr.mempoolOutpoints, outpoint)
		}
	}
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
	for sentOutpoint := range tr.mempoolOutpoints {
		if _, found := validOutpoints[sentOutpoint]; !found {
			delete(tr.mempoolOutpoints, sentOutpoint)
		}
	}
}

func (tr *Tracker) untrackTransactionDifference(transactions []*externalapi.DomainTransaction) {

	validTransactionIDs := make(sentTransactions, len(transactions))

	for _, transaction := range transactions {
		validTransactionIDs[*consensushashing.TransactionID(transaction)] = true
	}
	for sentTransaction := range tr.sentTransactions {
		if _, found := validTransactionIDs[sentTransaction]; !found {
			delete(tr.sentTransactions, sentTransaction)
		}
	}
}

func (tr *Tracker) trackOutpointAsReserved(outpoint *externalapi.DomainOutpoint) {
	tr.reservedOutpoints[*outpoint] = time.Now().Unix()
}

func (tr *Tracker) trackOutpointAsSent(outpoint *externalapi.DomainOutpoint) {
	tr.mempoolOutpoints[*outpoint] = true
}

func (tr *Tracker) trackTransaction(transaction *externalapi.DomainTransaction) {
	tr.sentTransactions[*consensushashing.TransactionID(transaction)] = true
}

func (tr *Tracker) untrackOutpointAsReserved(outpoint externalapi.DomainOutpoint) {
	delete(tr.reservedOutpoints, outpoint)
}
