package addressmanager

import (
	"math"
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

// Help function which returns a random index in the range [0, len(weights)-1] with probability weighted by `weights`
func weightedRand(weights []float32) int {
	sum := float32(0)
	for _, weight := range weights {
		sum += weight
	}
	randPoint := rand.Float32()
	scanPoint := float32(0)
	for i, weight := range weights {
		normalizedWeight := weight/sum
		scanPoint += normalizedWeight
		if randPoint <= scanPoint {
			return i
		}
	}
	return len(weights) - 1
}

// RandomAddresses returns count addresses at random from input list
func (amc *AddressRandomize) RandomAddresses(addresses []*address, count int) []*appmessage.NetAddress {
	if len(addresses) < count {
		count = len(addresses)
	}
	weights := make([]float32, 0, len(addresses))
	for _, addr := range addresses {
		weights = append(weights, float32(math.Pow(64, float64(addr.level))))
	}
	result := make([]*appmessage.NetAddress, 0, count)
	for count > 0 {
		i := weightedRand(weights)
		result = append(result, addresses[i].netAddress)
		// Delete entry i from both arrays
		addresses[i] = addresses[len(addresses)-1]
		weights[i] = weights[len(weights)-1]
		addresses = addresses[:len(addresses)-1]
		weights = weights[:len(weights)-1]
		// Update count
		count--
	}
	return result
}

