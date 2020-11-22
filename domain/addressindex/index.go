package addressindex

import (
	"bytes"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/util"
	"unsafe"
)

// Index represents a full list of transaction outputs indexed by addresses
type Index struct {
	utxoMapCache  UTXOMap
	dbContext     database.Database
	estimatedSize uint64
	outpointBuff  *bytes.Buffer
	maxCacheSize  uint64
}

// NewIndex creates a new Index and map the data from database with caching
func NewIndex(context database.Database, maxCacheSize uint64) *Index {
	return &Index{
		dbContext:    context,
		maxCacheSize: maxCacheSize,
		utxoMapCache: make(UTXOMap),
	}
}

// GetUTXOsByAddress returns a set of UTXOs associated with an address
func (in *Index) GetUTXOsByAddress(address string) (UTXOCollection, error) {
	collection, ok := in.utxoMapCache[address]
	if ok {
		return collection, nil
	}

	value, err := getFromUTXOMap(in.dbContext, []byte(address))
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	collection, err = DeserializeUTXOCollection(bytes.NewReader(value))
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	in.utxoMapCache[address] = collection
	return collection, nil
}

// GetUTXOsByAddresses returns a map of UTXO associated with addresses
func (in *Index) GetUTXOsByAddresses(addresses []string) (UTXOMap, error) {
	result := make(UTXOMap, len(addresses))
	for _, address := range addresses {
		collection, err := in.GetUTXOsByAddress(address)
		if err != nil {
			return nil, err
		}
		result[address] = collection
	}

	return result, nil
}

// getSizeOfAddressAndOutpoint returns estimated size of UTXOEntry & Outpoint in bytes
func getSizeOfAddressAndOutpoint(address string) uint64 {
	const staticSize = uint64(unsafe.Sizeof(externalapi.DomainOutpoint{}))
	return staticSize + uint64(len(address))
}

// add adds a new UTXO entry associated with an address to this Index
func (in *Index) add(address string, outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) {
	in.utxoMapCache.Add(address, outpoint, entry)
	in.estimatedSize += getSizeOfAddressAndOutpoint(address)
	in.checkAndCleanCachedData()
}

// remove removes a new UTXO associated with an address to this Index
func (in *Index) remove(address string, outpoint *externalapi.DomainOutpoint) {
	in.utxoMapCache.Remove(address, outpoint)
	in.checkAndCleanCachedData()
}

// checkAndCleanCachedData checks the Index estimated size and clean it if it reaches the limit
func (in *Index) checkAndCleanCachedData() {
	if in.estimatedSize > in.maxCacheSize {
		in.utxoMapCache = make(UTXOMap)
		in.estimatedSize = 0
	}
}

// GetAddress extracts addresses from scriptPublicKey
func GetAddress(scriptPublicKey []byte, prefix util.Bech32Prefix) (string, error) {
	_, address, err := txscript.ExtractScriptPubKeyAddress(scriptPublicKey, prefix)
	addressStr := address.EncodeAddress()
	return addressStr, err
}

// AddBlock adds the provided block's content to the Index
func (in *Index) AddBlock(block *externalapi.DomainBlock, blueScore uint64, prefix util.Bech32Prefix) (UTXOMap, error) {
	transactions := block.Transactions
	toAdd := make(UTXOMap)
	toRemove := make(UTXOMap)

	for _, transaction := range transactions {
		for _, txIn := range transaction.Inputs {
			address, err := GetAddress(txIn.UTXOEntry.ScriptPublicKey, prefix)
			if err != nil {
				return nil, err
			}
			toRemove[address].Add(&txIn.PreviousOutpoint, txIn.UTXOEntry)
		}

		isCoinbase := transactionhelper.IsCoinBase(transaction)
		for i, txOut := range transaction.Outputs {
			address, err := GetAddress(txOut.ScriptPublicKey, prefix)
			if err != nil {
				return nil, err
			}
			txID := consensusserialization.TransactionID(transaction)
			outpoint := externalapi.NewDomainOutpoint(txID, uint32(i))
			if err != nil {
				return nil, err
			}
			entry := externalapi.NewUTXOEntry(txOut.Value, txOut.ScriptPublicKey, isCoinbase, blueScore)
			toAdd[address].Add(outpoint, entry)
		}
	}

	return in.Update(toAdd, toRemove)
}

// Update updates the Index in the database
func (in *Index) Update(toAdd UTXOMap, toRemove UTXOMap) (UTXOMap, error) {
	changedAddresses := make(UTXOMap)

	for address, utxos := range toRemove {
		collection, ok := changedAddresses[address]
		if !ok {
			collection, err := in.GetUTXOsByAddress(address)
			if err != nil {
				return nil, err
			}
			changedAddresses[address] = collection
		}

		for outpoint := range utxos {
			in.remove(address, &outpoint)
			collection.Remove(&outpoint)
		}
	}

	for address, utxos := range toAdd {
		collection, ok := changedAddresses[address]
		if !ok {
			collection, err := in.GetUTXOsByAddress(address)
			if err != nil && !database.IsNotFoundError(err) {
				return nil, err
			}

			if collection == nil {
				collection = make(UTXOCollection)
			}

			changedAddresses[address] = collection
		}

		for outpoint, utxoEntry := range utxos {
			in.add(address, &outpoint, utxoEntry)
			collection.Add(&outpoint, utxoEntry)
		}
	}

	buffer := &bytes.Buffer{}
	for address, utxoCollection := range changedAddresses {
		buffer.Reset()
		err := SerializeUTXOCollection(buffer, utxoCollection)
		if err != nil {
			return nil, err
		}

		serializedUTXOCollection := buffer.Bytes()
		err = addToUTXOMap(in.dbContext, []byte(address), serializedUTXOCollection)
		if err != nil {
			return nil, err
		}
	}

	return changedAddresses, nil
}

// GetAddressesAndUTXOsFromTransaction extracts addresses and utxos from transaction
func GetAddressesAndUTXOsFromTransaction(transaction *externalapi.DomainTransaction, blueScore uint64, prefix util.Bech32Prefix) ([]string, UTXOCollection, error) {
	addressMap := make(map[string]struct{})
	utxos := make(UTXOCollection)

	for _, txIn := range transaction.Inputs {
		address, err := GetAddress(txIn.UTXOEntry.ScriptPublicKey, prefix)
		if err != nil {
			return nil, nil, err
		}
		addressMap[address] = struct{}{}
	}

	for i, txOut := range transaction.Outputs {
		address, err := GetAddress(txOut.ScriptPublicKey, prefix)
		if err != nil {
			return nil, nil, err
		}
		txID := consensusserialization.TransactionID(transaction)
		outpoint := externalapi.NewDomainOutpoint(txID, uint32(i))
		isCoinbase := transactionhelper.IsCoinBase(transaction)
		entry := externalapi.NewUTXOEntry(txOut.Value, txOut.ScriptPublicKey, isCoinbase, blueScore)
		utxos.Add(outpoint, entry)
		addressMap[address] = struct{}{}
	}

	addresses := make([]string, 0, len(addressMap))
	for address := range addressMap {
		addresses = append(addresses, address)
	}

	return addresses, utxos, nil
}
