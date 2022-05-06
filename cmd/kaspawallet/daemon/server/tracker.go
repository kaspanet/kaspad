package server

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"time"
)

type reservedOutpoints map[externalapi.DomainOutpoint]int64
type sentOutpoints map[externalapi.DomainOutpoint]int64

//Tracker tracks wallet server utxos via outpoints.
type Tracker struct {
	reservedOutpoints        reservedOutpoints
	sentOutpoints            sentOutpoints
	expiryDurationInSecounds int64
}

//NewTracker intializes and returns a new Tracker
func NewTracker() *Tracker {
	return &Tracker{
		reservedOutpoints:        make(reservedOutpoints),
		sentOutpoints:            make(sentOutpoints),
		expiryDurationInSecounds: 3, // 3 is somewhat aribitary
	}
}

func (tr *Tracker) isOutpointAvailable(outpoint *externalapi.DomainOutpoint) bool {
	var found bool

	_, found = tr.reservedOutpoints[*outpoint]
	if found {
		return false
	}
	_, found = tr.sentOutpoints[*outpoint]
	if found {
		return false
	}

	return true
}

func (tr *Tracker) untrackExpiredOutpointsAsResrved() {
	currentTimestamp := time.Now().Unix()
	for outpoint, reserveTimestamp := range tr.reservedOutpoints {
		if currentTimestamp-reserveTimestamp >= tr.expiryDurationInSecounds {
			delete(tr.reservedOutpoints, outpoint)
		}

	}
}

func (tr *Tracker) untrackOutpointDifferenceViaWalletUTXOs(utxos []*walletUTXO) {

	for trackedOutpoint := range tr.sentOutpoints {
		for _, utxo := range utxos {
			outpoint := externalapi.DomainOutpoint{
				TransactionID: utxo.Outpoint.TransactionID,
				Index:         utxo.Outpoint.Index,
			}
			if trackedOutpoint == outpoint {
				break
			}
			delete(tr.sentOutpoints, trackedOutpoint)

		}
	}

	for trackedOutpoint := range tr.reservedOutpoints {
		for _, utxo := range utxos {
			outpoint := externalapi.DomainOutpoint{
				TransactionID: utxo.Outpoint.TransactionID,
				Index:         utxo.Outpoint.Index,
			}
			if trackedOutpoint == outpoint {
				break
			}
			delete(tr.reservedOutpoints, trackedOutpoint)
		}
	}

}

func (tr *Tracker) trackOutpointAsReserved(outpoint externalapi.DomainOutpoint) {
	tr.reservedOutpoints[outpoint] = time.Now().Unix()
}

func (tr *Tracker) trackOutpointAsSent(outpoint externalapi.DomainOutpoint) {
	tr.sentOutpoints[outpoint] = time.Now().Unix()
}

func (tr *Tracker) untrackOutpointAsReserved(outpoint externalapi.DomainOutpoint) {
	delete(tr.reservedOutpoints, outpoint)
}
