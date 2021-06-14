package prefixmanager

import (
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

var activePrefixKey = database.MakeBucket(nil).Key([]byte("active-prefix"))
var inactivePrefixKey = database.MakeBucket(nil).Key([]byte("inactive-prefix"))

// Prefix is a database prefix that is used to manage more than one database at once.
type Prefix struct {
	value byte
}

// Serialize serializes the prefix into a byte slice
func (p *Prefix) Serialize() []byte {
	return []byte{p.value}
}

// Equal returns whether p equals to other
func (p *Prefix) Equal(other *Prefix) bool {
	return p.value == other.value
}

// NewPrefix returns a new prefix with the given value
func NewPrefix(value byte) *Prefix {
	return &Prefix{value: value}
}

// ActivePrefix returns the current active database prefix, and whether it exists
func ActivePrefix(dataAccessor database.DataAccessor) (*Prefix, bool, error) {
	prefixBytes, err := dataAccessor.Get(activePrefixKey)
	if database.IsNotFoundError(err) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	prefix, err := parsePrefix(prefixBytes)
	if err != nil {
		return nil, false, err
	}

	return prefix, true, nil
}

// InactivePrefix returns the current inactive database prefix, and whether it exists
func InactivePrefix(dataAccessor database.DataAccessor) (*Prefix, bool, error) {
	prefixBytes, err := dataAccessor.Get(inactivePrefixKey)
	if database.IsNotFoundError(err) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	prefix, err := parsePrefix(prefixBytes)
	if err != nil {
		return nil, false, err
	}

	return prefix, true, nil
}

func parsePrefix(prefixBytes []byte) (*Prefix, error) {
	if len(prefixBytes) > 1 {
		return nil, errors.Errorf("invalid length %d for prefix", len(prefixBytes))
	}

	return NewPrefix(prefixBytes[0]), nil
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

	prefix, err := parsePrefix(prefixBytes)
	if err != nil {
		return err
	}

	err = deletePrefix(dataAccessor, prefix)
	if err != nil {
		return err
	}

	return dataAccessor.Delete(inactivePrefixKey)
}

func deletePrefix(dataAccessor database.DataAccessor, prefix *Prefix) error {
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
func SetPrefixAsActive(dataAccessor database.DataAccessor, prefix *Prefix) error {
	return dataAccessor.Put(activePrefixKey, prefix.Serialize())
}

// SetPrefixAsInactive sets the given prefix as the inactive prefix
func SetPrefixAsInactive(dataAccessor database.DataAccessor, prefix *Prefix) error {
	return dataAccessor.Put(inactivePrefixKey, prefix.Serialize())
}
