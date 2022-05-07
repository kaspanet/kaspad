package server

import (
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type reservedOutpoints map[externalapi.DomainOutpoint]int64
type sentOutpoints map[externalapi.DomainOutpoint]int64

//Tracker tracks wallet server utxos via outpoints.
type Tracker struct {
	reservedOutpoints        reservedOutpoints
	sentOutpoints	         sentOutpoints
	expiryDurationInSecounds int64
}

//NewTracker intializes and returns a new Tracker
func NewTracker() *Tracker {
	return &Tracker{
		reservedOutpoints:        make(reservedOutpoints),
		sentOutpoints:            make(sentOutpoints),
		expiryDurationInSecounds: 1, // 1 corrosponds to a sync ticker interval, as well as current average BPS
	}
}

func (tr *Tracker) isOutpointAvailable(outpoint *externalapi.DomainOutpoint) bool {
	var found bool

	if _, found = tr.reservedOutpoints[*outpoint]; found {
		return false
	}

	if _, found = tr.sentOutpoints[*outpoint]; found {
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
	for outpoint, sentTimestamp := range tr.reservedOutpoints {
		if currentTimestamp-sentTimestamp >= tr.expiryDurationInSecounds {
			delete(tr.sentOutpoints, outpoint)
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
	for sentOutpoint := range tr.sentOutpoints {
		if _, found := validOutpoints[sentOutpoint]; !found {
			delete(tr.sentOutpoints, sentOutpoint)
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
