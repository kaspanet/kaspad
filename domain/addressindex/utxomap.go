package addressindex

import (
	"bytes"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"unsafe"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/addressindex/dbaccess"
)

// OutpointCollection represents a set of UTXOs outpoints
type OutpointCollection map[appmessage.Outpoint]struct{}

// Add adds a new UTXO to this collection
func (uc OutpointCollection) Add(outpoint appmessage.Outpoint) {
	uc[outpoint] = struct{}{}
}

// Remove removes a UTXO from this collection if it exists
func (uc OutpointCollection) Remove(outpoint appmessage.Outpoint) {
	delete(uc, outpoint)
}

// UTXOMap represents a set of UTXOs sets indexed by their addresses
type UTXOMap map[string]OutpointCollection

// Get returns the OutpointCollection represented by provided address,
// and a boolean value indicating if said address is in the set or not
func (um UTXOMap) Get(address string) (OutpointCollection, bool) {
	if collection, ok := um[address]; ok {
		return collection, true
	}

	return nil, false
}

// Add adds a new address to this utxoMap
func (um UTXOMap) Add(address string, outpoint appmessage.Outpoint) {
	if collection, ok := um[address]; ok {
		collection[outpoint] = struct{}{}
	}
}

// Remove removes a UTXO by address from this utxoMap if it exists
func (um UTXOMap) Remove(address string, outpoint appmessage.Outpoint) {
	if collection, ok := um[address]; ok {
		delete(collection, outpoint)
	}
}

// Addresses returns all addresses
func (um UTXOMap) Addresses() []string {
	addresses := make([]string, 0, len(um))
	for key := range um {
		addresses = append(addresses, key)
	}
	return addresses
}

// FullUTXOMap represents a full list of transaction outputs indexed by addresses
type FullUTXOMap struct {
	utxoMapCache  UTXOMap
	dbContext     database.Database
	estimatedSize uint64
	outpointBuff  *bytes.Buffer
	maxCacheSize  uint64
}

// FullUTXOMap creates a new FullUTXOMap and map the data context with caching
func NewFullUTXOMap(context database.Database, maxCacheSize uint64) *FullUTXOMap {
	return &FullUTXOMap{
		dbContext:    context,
		maxCacheSize: maxCacheSize,
		utxoMapCache: make(UTXOMap),
	}
}

// GetUTXOsByAddress returns a set of UTXOs associated with an address
func (fum *FullUTXOMap) GetUTXOsByAddress(address string) (OutpointCollection, error) {
	collection, ok := fum.utxoMapCache[address]
	if ok {
		return collection, nil
	}

	value, err := dbaccess.GetFromUTXOMap(fum.dbContext, []byte(address))
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	collection, err = DeserializeOutpointCollection(bytes.NewReader(value))
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	fum.utxoMapCache[address] = collection
	return collection, nil
}

// GetUTXOsByAddresses returns a map of UTXO associated with addresses
func (fum *FullUTXOMap) GetUTXOsByAddresses(addresses []string) (UTXOMap, error) {
	result := make(UTXOMap, len(addresses))
	for _, address := range addresses {
		collection, err := fum.GetUTXOsByAddress(address)
		if err != nil {
			return nil, err
		}
		result[address] = collection
	}

	return result, nil
}

// getSizeOfAddressAndOutpoint returns estimated size of UTXOEntry & Outpoint in bytes
func getSizeOfAddressAndOutpoint(address string) uint64 {
	const staticSize = uint64(unsafe.Sizeof(appmessage.Outpoint{}))
	return staticSize + uint64(len(address))
}

// add adds a new UTXO associated with an address to this FullUTXOMap
func (fum *FullUTXOMap) add(address string, outpoint appmessage.Outpoint) {
	fum.utxoMapCache.Add(address, outpoint)
	fum.estimatedSize += getSizeOfAddressAndOutpoint(address)
	fum.checkAndCleanCachedData()
}

// remove removes a new UTXO associated with an address to this FullUTXOMap
func (fum *FullUTXOMap) remove(address string, outpoint appmessage.Outpoint) {
	fum.utxoMapCache.Remove(address, outpoint)
	fum.checkAndCleanCachedData()
}

// checkAndCleanCachedData checks the FullUTXOMap estimated size and clean it if it reaches the limit
func (fum *FullUTXOMap) checkAndCleanCachedData() {
	if fum.estimatedSize > fum.maxCacheSize {
		fum.utxoMapCache = make(UTXOMap)
		fum.estimatedSize = 0
	}
}

// Update updates the FullUTXOMap in the database
func (fum *FullUTXOMap) Update(toAdd UTXOMap, toRemove UTXOMap) (UTXOMap, error) {
	changedAddresses := make(UTXOMap)

	for address, outpoints := range toRemove {
		collection, ok := changedAddresses[address]
		if !ok {
			collection, err := fum.GetUTXOsByAddress(address)
			if err != nil {
				return nil, err
			}
			changedAddresses[address] = collection
		}

		for outpoint := range outpoints {
			fum.remove(address, outpoint)
			collection.Remove(outpoint)
		}
	}

	for address, outpoints := range toAdd {
		collection, ok := changedAddresses[address]
		if !ok {
			collection, err := fum.GetUTXOsByAddress(address)
			if err != nil && !database.IsNotFoundError(err) {
				return nil, err
			}

			if collection == nil {
				collection = make(OutpointCollection)
			}

			changedAddresses[address] = collection
		}

		for outpoint := range outpoints {
			fum.add(address, outpoint)
			collection.Add(outpoint)
		}
	}

	buffer := &bytes.Buffer{}
	for address, utxoCollection := range changedAddresses {
		buffer.Reset()
		err := SerializeOutpointCollection(buffer, utxoCollection)
		if err != nil {
			return nil, err
		}

		serializedUTXOCollection := buffer.Bytes()
		err = dbaccess.AddToUTXOMap(fum.dbContext, []byte(address), serializedUTXOCollection)
		if err != nil {
			return nil, err
		}
	}

	return changedAddresses, nil
}
