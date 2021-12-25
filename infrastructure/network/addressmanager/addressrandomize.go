package addressmanager

import (
	"math/rand"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
)

// AddressRandomize implement addressRandomizer interface
type AddressRandomize struct {
	random *rand.Rand
}

// NewAddressRandomize returns a new RandomizeAddress.
func NewAddressRandomize() *AddressRandomize {
	return &AddressRandomize{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RandomAddresses returns count addresses at random from input list
func (amc *AddressRandomize) RandomAddresses(addresses []*address, count int) []*appmessage.NetAddress {
	lenAddresses := len(addresses)
	if lenAddresses < count {
		count = lenAddresses
	}
	const numLevels = maxLevel + 1
	var addressesByLevels [numLevels][]*address
	var lengthsByLevels [numLevels]struct{ lower, length int }
	for i := range addressesByLevels {
		addressesByLevels[i] = make([]*address, 0, lenAddresses/int(numLevels))
	}
	for _, address := range addresses {
		addressesByLevels[address.level] = append(addressesByLevels[address.level], address)
	}
	upper := 0
	for i, addressesOfLevel := range addressesByLevels {
		length := len(addressesOfLevel)
		lengthsByLevels[i].length, lengthsByLevels[i].lower = length, upper
		upper += (1 << i) * length
	}

	result := make([]*appmessage.NetAddress, 0, count)

	for count > 0 {
		random := rand.Intn(upper)
		for i := range lengthsByLevels {
			e := lengthsByLevels[numLevels-uint8(i)-1]
			if random > e.lower {
				index := (random - e.lower) / e.length
				address := addresses[index]
				if address != nil {
					result = append(result, address.netAddress)
					addresses[index] = nil // use address only once
					count--
				}
				break
			}
		}
	}
	return result
}
