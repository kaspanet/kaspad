// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addressmanager

import (
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
)

// KnownAddress tracks information about a known network address that is used
// to determine how viable an address is.
type KnownAddress struct {
	netAddress     *appmessage.NetAddress
	sourceAddress  *appmessage.NetAddress
	attempts       int
	lastAttempt    mstime.Time
	lastSuccess    mstime.Time
	tried          bool
	referenceCount int // reference count of new buckets
	subnetworkID   *externalapi.DomainSubnetworkID
	isBanned       bool
	bannedTime     mstime.Time
}

// NetAddress returns the underlying appmessage.NetAddress associated with the
// known address.
func (ka *KnownAddress) NetAddress() *appmessage.NetAddress {
	return ka.netAddress
}

// SubnetworkID returns the subnetwork ID of the known address.
func (ka *KnownAddress) SubnetworkID() *externalapi.DomainSubnetworkID {
	return ka.subnetworkID
}

// LastAttempt returns the last time the known address was attempted.
func (ka *KnownAddress) LastAttempt() mstime.Time {
	return ka.lastAttempt
}

// chance returns the selection probability for a known address. The priority
// depends upon how recently the address has been seen, how recently it was last
// attempted and how often attempts to connect to it have failed.
func (ka *KnownAddress) chance() float64 {
	now := mstime.Now()
	lastAttempt := now.Sub(ka.lastAttempt)

	if lastAttempt < 0 {
		lastAttempt = 0
	}

	c := 1.0

	// Very recent attempts are less likely to be retried.
	if lastAttempt < 10*time.Minute {
		c *= 0.01
	}

	// Failed attempts deprioritise.
	for i := ka.attempts; i > 0; i-- {
		c /= 1.5
	}

	return c
}

// isBad returns true if the address in question has not been tried in the last
// minute and meets one of the following criteria:
// 1) It claims to be from the future
// 2) It hasn't been seen in over a month
// 3) It has failed at least three times and never succeeded
// 4) It has failed ten times in the last week
// All addresses that meet these criteria are assumed to be worthless and not
// worth keeping hold of.
func (ka *KnownAddress) isBad() bool {
	if ka.lastAttempt.After(mstime.Now().Add(-1 * time.Minute)) {
		return false
	}

	// From the future?
	if ka.netAddress.Timestamp.After(mstime.Now().Add(10 * time.Minute)) {
		return true
	}

	// Over a month old?
	if ka.netAddress.Timestamp.Before(mstime.Now().Add(-1 * numMissingDays * time.Hour * 24)) {
		return true
	}

	// Never succeeded?
	if ka.lastSuccess.IsZero() && ka.attempts >= numRetries {
		return true
	}

	// Hasn't succeeded in too long?
	if !ka.lastSuccess.After(mstime.Now().Add(-1*minBadDays*time.Hour*24)) &&
		ka.attempts >= maxFailures {
		return true
	}

	return false
}
