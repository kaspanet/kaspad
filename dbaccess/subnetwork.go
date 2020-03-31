package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

var subnetworkBucket = database2.MakeBucket([]byte("subnetworks"))

// FetchSubnetworkData returns the subnetwork data by its ID.
func FetchSubnetworkData(context Context, subnetworkID *subnetworkid.SubnetworkID) ([]byte, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	key := subnetworkKey(subnetworkID)
	return accessor.Get(key)
}

// RegisterSubnetwork stores mappings from ID of the subnetwork to the subnetwork data.
func RegisterSubnetwork(context Context, subnetworkID *subnetworkid.SubnetworkID, subnetworkData []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := subnetworkKey(subnetworkID)
	return accessor.Put(key, subnetworkData)
}

// SubnetworkExists returns whether the subnetwork exists in the database.
func SubnetworkExists(context Context, subnetworkID *subnetworkid.SubnetworkID) (bool, error) {
	accessor, err := context.accessor()
	if err != nil {
		return false, err
	}

	key := subnetworkKey(subnetworkID)
	return accessor.Has(key)
}

func subnetworkKey(subnetworkID *subnetworkid.SubnetworkID) []byte {
	return subnetworkBucket.Key(subnetworkID[:])
}
