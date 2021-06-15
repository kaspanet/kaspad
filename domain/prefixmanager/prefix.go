package prefixmanager

import (
	"github.com/kaspanet/kaspad/domain/prefixmanager/prefix"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

var activePrefixKey = database.MakeBucket(nil).Key([]byte("active-prefix"))
var inactivePrefixKey = database.MakeBucket(nil).Key([]byte("inactive-prefix"))

// ActivePrefix returns the current active database prefix, and whether it exists
func ActivePrefix(dataAccessor database.DataAccessor) (*prefix.Prefix, bool, error) {
	prefixBytes, err := dataAccessor.Get(activePrefixKey)
	if database.IsNotFoundError(err) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	prefix, err := prefix.Deserialize(prefixBytes)
	if err != nil {
		return nil, false, err
	}

	return prefix, true, nil
}

// InactivePrefix returns the current inactive database prefix, and whether it exists
func InactivePrefix(dataAccessor database.DataAccessor) (*prefix.Prefix, bool, error) {
	prefixBytes, err := dataAccessor.Get(inactivePrefixKey)
	if database.IsNotFoundError(err) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	prefix, err := prefix.Deserialize(prefixBytes)
	if err != nil {
		return nil, false, err
	}

	return prefix, true, nil
}

// DeleteInactivePrefix deletes all data associated with the inactive database prefix, including itself.
func DeleteInactivePrefix(dataAccessor database.DataAccessor) error {
	prefixBytes, err := dataAccessor.Get(inactivePrefixKey)
	if database.IsNotFoundError(err) {
		return nil
	}

	if err != nil {
		return err
	}

	prefix, err := prefix.Deserialize(prefixBytes)
	if err != nil {
		return err
	}

	err = deletePrefix(dataAccessor, prefix)
	if err != nil {
		return err
	}

	return dataAccessor.Delete(inactivePrefixKey)
}

func deletePrefix(dataAccessor database.DataAccessor, prefix *prefix.Prefix) error {
	log.Infof("Deleting database prefix %x", prefix)
	prefixBucket := database.MakeBucket(prefix.Serialize())
	cursor, err := dataAccessor.Cursor(prefixBucket)
	if err != nil {
		return err
	}

	defer cursor.Close()

	for ok := cursor.First(); ok; ok = cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}

		err = dataAccessor.Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetPrefixAsActive sets the given prefix as the active prefix
func SetPrefixAsActive(dataAccessor database.DataAccessor, prefix *prefix.Prefix) error {
	return dataAccessor.Put(activePrefixKey, prefix.Serialize())
}

// SetPrefixAsInactive sets the given prefix as the inactive prefix
func SetPrefixAsInactive(dataAccessor database.DataAccessor, prefix *prefix.Prefix) error {
	return dataAccessor.Put(inactivePrefixKey, prefix.Serialize())
}
