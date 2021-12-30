package addressmanager

import (
	"math"
	"math/rand"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
)

// AddressRandomize implement addressRandomizer interface
type AddressRandomize struct {
	random         *rand.Rand
	maxFailedCount uint64
}

// NewAddressRandomize returns a new RandomizeAddress.
func NewAddressRandomize(maxFailedCount uint64) *AddressRandomize {
	return &AddressRandomize{
		random:         rand.New(rand.NewSource(time.Now().UnixNano())),
		maxFailedCount: maxFailedCount,
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
		normalizedWeight := weight / sum
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
		weights = append(weights, float32(math.Pow(64, float64(amc.maxFailedCount-addr.connectionFailedCount))))
	}
	result := make([]*appmessage.NetAddress, 0, count)
	for count > 0 {
		i := weightedRand(weights)
		result = append(result, addresses[i].netAddress)
		// Zero entry i to avoid re-selection
		weights[i] = 0
		// Update count
		count--
	}
	return result
}
