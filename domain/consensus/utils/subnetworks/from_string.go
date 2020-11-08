package subnetworks

import (
	"encoding/hex"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// FromString creates a DomainSubnetworkID from the given byte slice
func FromString(str string) (*externalapi.DomainSubnetworkID, error) {
	runes := []rune(str)
	for i := 0; i < externalapi.DomainSubnetworkIDSize*2; i++ {
		runes[i], runes[externalapi.DomainSubnetworkIDSize-1-i] = runes[externalapi.DomainSubnetworkIDSize-1-i], runes[i]
	}
	subnetworkIDBytes, err := hex.DecodeString(string(runes))
	if err != nil {
		return nil, err
	}
	return FromBytes(subnetworkIDBytes)
}
