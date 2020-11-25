package addressindex

import (
	"bytes"
	"errors"
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

// getSizeOfUTXOEntryAndOutpoint returns estimated size of UTXOEntry & Outpoint in bytes
func getSizeOfUTXOEntryAndOutpoint(entry *externalapi.UTXOEntry) uint64 {
	const staticSize = uint64(unsafe.Sizeof(externalapi.DomainOutpoint{})) + uint64(unsafe.Sizeof(externalapi.UTXOEntry{}))
	return staticSize + uint64(len(entry.ScriptPublicKey))
}

// add adds a new UTXO entry associated with an address to this Index
func (in *Index) add(address string, outpoint *externalapi.DomainOutpoint, entry *externalapi.UTXOEntry) {
	in.utxoMapCache.Add(address, outpoint, entry)
	in.estimatedSize += getSizeOfUTXOEntryAndOutpoint(entry)
	in.checkAndCleanCachedData()
}

// remove removes an UTXO associated with an address from this Index
func (in *Index) remove(address string, outpoint *externalapi.DomainOutpoint) bool {
	utxosOfAddress, ok := in.utxoMapCache.Get(address)
	if !ok {
		return false
	}
	entry, ok := utxosOfAddress.Get(outpoint)
	if !ok {
		return false
	}
	utxosOfAddress.Remove(outpoint)
	in.estimatedSize -= getSizeOfUTXOEntryAndOutpoint(entry)
	return true
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
	if address != nil {
		addressStr := address.EncodeAddress()
		return addressStr, nil
	}
	return "", err
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
			toRemove.Add(address, &txIn.PreviousOutpoint, txIn.UTXOEntry)
		}

		isCoinbase := transactionhelper.IsCoinBase(transaction)
		for i, txOut := range transaction.Outputs {
			address, err := GetAddress(txOut.ScriptPublicKey, prefix)
			if err != nil {
				return nil, err
			}
			txID := consensusserialization.TransactionID(transaction)
			outpoint := externalapi.NewDomainOutpoint(txID, uint32(i))
			entry := externalapi.NewUTXOEntry(txOut.Value, txOut.ScriptPublicKey, isCoinbase, blueScore)
			toAdd.Add(address, outpoint, entry)
		}
	}

	return in.Update(toAdd, toRemove)
}

// Update updates the Index in the database
func (in *Index) Update(toAdd UTXOMap, toRemove UTXOMap) (UTXOMap, error) {
	changedAddresses := make(UTXOMap)
	for address, utxosToRemove := range toRemove {
		utxosOfAddress, err := in.GetUTXOsByAddress(address)
		if err != nil {
			return nil, err
		}
		if utxosOfAddress == nil {
			return nil, errors.New("address was not found")
		}
		for outpoint := range utxosToRemove {
			in.remove(address, &outpoint)
			utxosOfAddress.Remove(&outpoint)
		}
		changedAddresses[address] = utxosOfAddress
	}

	for address, utxosToAdd := range toAdd {
		utxosOfAddress, ok := changedAddresses.Get(address)
		if !ok {
			var err error
			utxosOfAddress, err = in.GetUTXOsByAddress(address)
			if err != nil && !database.IsNotFoundError(err) {
				return nil, err
			}
			if utxosOfAddress == nil {
				utxosOfAddress = make(UTXOCollection)
			}
			changedAddresses[address] = utxosOfAddress
		}

		for outpoint, utxoEntry := range utxosToAdd {
			in.add(address, &outpoint, utxoEntry)
			utxosOfAddress.Add(&outpoint, utxoEntry)
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
