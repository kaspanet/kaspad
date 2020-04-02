package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

var subnetworkBucket = database.MakeBucket([]byte("subnetworks"))

func subnetworkKey(subnetworkID *subnetworkid.SubnetworkID) []byte {
	return subnetworkBucket.Key(subnetworkID[:])
}

// FetchSubnetworkData returns the subnetwork data by its ID.
func FetchSubnetworkData(context Context, subnetworkID *subnetworkid.SubnetworkID) ([]byte, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	key := subnetworkKey(subnetworkID)
	return accessor.Get(key)
}

// StoreSubnetwork stores mappings from ID of the subnetwork to the subnetwork data.
func StoreSubnetwork(context Context, subnetworkID *subnetworkid.SubnetworkID, subnetworkData []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := subnetworkKey(subnetworkID)
	return accessor.Put(key, subnetworkData)
}

// HasSubnetwork returns whether the subnetwork exists in the database.
func HasSubnetwork(context Context, subnetworkID *subnetworkid.SubnetworkID) (bool, error) {
	accessor, err := context.accessor()
	if err != nil {
		return false, err
	}

	key := subnetworkKey(subnetworkID)
	return accessor.Has(key)
}
