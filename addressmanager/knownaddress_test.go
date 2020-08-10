// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addressmanager_test

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"math"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/addressmanager"
	"github.com/kaspanet/kaspad/domainmessage"
)

func TestChance(t *testing.T) {
	now := mstime.Now()
	var tests = []struct {
		addr     *addressmanager.KnownAddress
		expected float64
	}{
		{
			//Test normal case
			addressmanager.TstNewKnownAddress(&domainmessage.NetAddress{Timestamp: now.Add(-35 * time.Second)},
				0, mstime.Now().Add(-30*time.Minute), mstime.Now(), false, 0),
			1.0,
		}, {
			//Test case in which lastseen < 0
			addressmanager.TstNewKnownAddress(&domainmessage.NetAddress{Timestamp: now.Add(20 * time.Second)},
				0, mstime.Now().Add(-30*time.Minute), mstime.Now(), false, 0),
			1.0,
		}, {
			//Test case in which lastAttempt < 0
			addressmanager.TstNewKnownAddress(&domainmessage.NetAddress{Timestamp: now.Add(-35 * time.Second)},
				0, mstime.Now().Add(30*time.Minute), mstime.Now(), false, 0),
			1.0 * .01,
		}, {
			//Test case in which lastAttempt < ten minutes
			addressmanager.TstNewKnownAddress(&domainmessage.NetAddress{Timestamp: now.Add(-35 * time.Second)},
				0, mstime.Now().Add(-5*time.Minute), mstime.Now(), false, 0),
			1.0 * .01,
		}, {
			//Test case with several failed attempts.
			addressmanager.TstNewKnownAddress(&domainmessage.NetAddress{Timestamp: now.Add(-35 * time.Second)},
				2, mstime.Now().Add(-30*time.Minute), mstime.Now(), false, 0),
			1 / 1.5 / 1.5,
		},
	}

	err := .0001
	for i, test := range tests {
		chance := addressmanager.TstKnownAddressChance(test.addr)
		if math.Abs(test.expected-chance) >= err {
			t.Errorf("case %d: got %f, expected %f", i, chance, test.expected)
		}
	}
}

func TestIsBad(t *testing.T) {
	now := mstime.Now()
	future := now.Add(35 * time.Minute)
	monthOld := now.Add(-43 * time.Hour * 24)
	secondsOld := now.Add(-2 * time.Second)
	minutesOld := now.Add(-27 * time.Minute)
	hoursOld := now.Add(-5 * time.Hour)
	zeroTime := mstime.Time{}

	futureNa := &domainmessage.NetAddress{Timestamp: future}
	minutesOldNa := &domainmessage.NetAddress{Timestamp: minutesOld}
	monthOldNa := &domainmessage.NetAddress{Timestamp: monthOld}
	currentNa := &domainmessage.NetAddress{Timestamp: secondsOld}

	//Test addresses that have been tried in the last minute.
	if addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(futureNa, 3, secondsOld, zeroTime, false, 0)) {
		t.Errorf("test case 1: addresses that have been tried in the last minute are not bad.")
	}
	if addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(monthOldNa, 3, secondsOld, zeroTime, false, 0)) {
		t.Errorf("test case 2: addresses that have been tried in the last minute are not bad.")
	}
	if addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(currentNa, 3, secondsOld, zeroTime, false, 0)) {
		t.Errorf("test case 3: addresses that have been tried in the last minute are not bad.")
	}
	if addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(currentNa, 3, secondsOld, monthOld, true, 0)) {
		t.Errorf("test case 4: addresses that have been tried in the last minute are not bad.")
	}
	if addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(currentNa, 2, secondsOld, secondsOld, true, 0)) {
		t.Errorf("test case 5: addresses that have been tried in the last minute are not bad.")
	}

	//Test address that claims to be from the future.
	if !addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(futureNa, 0, minutesOld, hoursOld, true, 0)) {
		t.Errorf("test case 6: addresses that claim to be from the future are bad.")
	}

	//Test address that has not been seen in over a month.
	if !addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(monthOldNa, 0, minutesOld, hoursOld, true, 0)) {
		t.Errorf("test case 7: addresses more than a month old are bad.")
	}

	//It has failed at least three times and never succeeded.
	if !addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(minutesOldNa, 3, minutesOld, zeroTime, true, 0)) {
		t.Errorf("test case 8: addresses that have never succeeded are bad.")
	}

	//It has failed ten times in the last week
	if !addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(minutesOldNa, 10, minutesOld, monthOld, true, 0)) {
		t.Errorf("test case 9: addresses that have not succeeded in too long are bad.")
	}

	//Test an address that should work.
	if addressmanager.TstKnownAddressIsBad(addressmanager.TstNewKnownAddress(minutesOldNa, 2, minutesOld, hoursOld, true, 0)) {
		t.Errorf("test case 10: This should be a valid address.")
	}
}
