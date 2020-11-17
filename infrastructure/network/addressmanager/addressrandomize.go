package addressmanager

import (
	"math/rand"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
)

// AddressRandomize implement AddressRandomizer interface
type AddressRandomize struct {
	random *rand.Rand
}

// NewAddressRandomize returns a new RandomizeAddress.
func NewAddressRandomize() *AddressRandomize {
	return &AddressRandomize{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RandomAddress returns a random address from input list
func (amc *AddressRandomize) RandomAddress(addresses []*appmessage.NetAddress) *appmessage.NetAddress {
	if len(addresses) > 0 {
		randomIndex := rand.Intn(len(addresses))
		return addresses[randomIndex]
	}

	return nil
}

// RandomAddresses returns count addresses at random from input list
func (amc *AddressRandomize) RandomAddresses(addresses []*appmessage.NetAddress, count int) []*appmessage.NetAddress {
	result := make([]*appmessage.NetAddress, 0, count)
	if len(addresses) < count {
		count = len(addresses)
	}

	randomIndexes := rand.Perm(len(addresses))
	for i := 0; i < count; i++ {
		result = append(result, addresses[randomIndexes[i]])
	}

	return result
}
