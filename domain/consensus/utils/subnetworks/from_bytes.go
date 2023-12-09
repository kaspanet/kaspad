package subnetworks

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
)

// FromBytes creates a DomainSubnetworkID from the given byte slice
func FromBytes(subnetworkIDBytes []byte) (*externalapi.DomainSubnetworkID, error) {
	if len(subnetworkIDBytes) != externalapi.DomainSubnetworkIDSize {
		return nil, errors.Errorf("invalid hash size. Want: %d, got: %d",
			externalapi.DomainSubnetworkIDSize, len(subnetworkIDBytes))
	}
	var domainSubnetworkID externalapi.DomainSubnetworkID
	copy(domainSubnetworkID[:], subnetworkIDBytes)
	return &domainSubnetworkID, nil
}
