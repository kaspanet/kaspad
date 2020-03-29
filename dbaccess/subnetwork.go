package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

var subnetworkBucket = database2.MakeBucket([]byte("subnetworks"))

// FetchSubnetworkData returns the subnetwork data by its ID.
func FetchSubnetworkData(context Context, subnetworkID *subnetworkid.SubnetworkID) (subnetworkData []byte, found bool, err error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, false, err
	}

	key := subnetworkBucket.Key(subnetworkID[:])
	return accessor.Get(key)
}

// RegisterSubnetwork stores mappings from ID of the subnetwork to the subnetwork data.
func RegisterSubnetwork(context Context, subnetworkID *subnetworkid.SubnetworkID, subnetworkData []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := subnetworkBucket.Key(subnetworkID[:])
	return accessor.Put(key, subnetworkData)
}

// SubnetworkExists returns whether the subnetwork exists in the database.
func SubnetworkExists(context Context, subnetworkID *subnetworkid.SubnetworkID) (bool, error) {
	accessor, err := context.accessor()
	if err != nil {
		return false, err
	}

	key := subnetworkBucket.Key(subnetworkID[:])
	_, exists, err := accessor.Get(key)
	if err != nil {
		return false, err
	}

	return exists, nil
}
