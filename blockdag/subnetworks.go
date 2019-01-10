package blockdag

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/daglabs/btcd/util"

	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

// registerSubNetworks scans a list of accepted transactions, singles out
// sub-network registry transactions, validates them, and registers a new
// sub-network based on it.
// This function returns an error if one or more transactions are invalid
func registerSubNetworks(dbTx database.Tx, txs []*TxWithBlockHash) error {
	validSubNetworkRegistryTxs := make([]*wire.MsgTx, 0)
	for _, txData := range txs {
		tx := txData.Tx.MsgTx()
		if tx.SubNetworkID == wire.SubNetworkRegistry {
			err := validateSubNetworkRegistryTransaction(tx)
			if err != nil {
				return err
			}
			validSubNetworkRegistryTxs = append(validSubNetworkRegistryTxs, tx)
		}
	}

	for _, registryTx := range validSubNetworkRegistryTxs {
		subNetworkID, err := buildSubNetworkID(registryTx)
		if err != nil {
			return err
		}
		if _, err := dbGetSubNetwork(dbTx, subNetworkID); err != nil {
			createdSubNetwork := newSubNetwork(registryTx)
			err := dbRegisterSubNetwork(dbTx, subNetworkID, createdSubNetwork)
			if err != nil {
				return fmt.Errorf("failed registering sub-network"+
					"for tx '%s': %s", registryTx.TxHash(), err)
			}
		}
	}

	return nil
}

// validateSubNetworkRegistryTransaction makes sure that a given sub-network registry
// transaction is valid. Such a transaction is valid iff:
// - Its entire payload is a uint64 (8 bytes)
func validateSubNetworkRegistryTransaction(tx *wire.MsgTx) error {
	if len(tx.Payload) != 8 {
		return ruleError(ErrSubNetworkRegistry, fmt.Sprintf("validation failed: subnetwork registry"+
			"tx '%s' has an invalid payload", tx.TxHash()))
	}

	return nil
}

// buildSubNetworkID creates a sub-network ID from a sub-network registry transaction
func buildSubNetworkID(tx *wire.MsgTx) (*subnetworkid.SubNetworkID, error) {
	txHash := tx.TxHash()
	return subnetworkid.New(util.Hash160(txHash[:]))
}

// subNetwork returns a registered sub-network. If the sub-network does not exist
// this method returns an error.
func (dag *BlockDAG) subNetwork(subNetworkID *subnetworkid.SubNetworkID) (*subNetwork, error) {
	var sNet *subNetwork
	var err error
	dbErr := dag.db.View(func(dbTx database.Tx) error {
		sNet, err = dbGetSubNetwork(dbTx, subNetworkID)
		return nil
	})
	if dbErr != nil {
		return nil, fmt.Errorf("could not retrieve sub-network '%d': %s", subNetworkID, dbErr)
	}
	if err != nil {
		return nil, fmt.Errorf("could not retrieve sub-network '%d': %s", subNetworkID, err)
	}

	return sNet, nil
}

// GasLimit returns the gas limit of a registered sub-network. If the sub-network does not
// exist this method returns an error.
func (dag *BlockDAG) GasLimit(subNetworkID *subnetworkid.SubNetworkID) (uint64, error) {
	sNet, err := dag.subNetwork(subNetworkID)
	if err != nil {
		return 0, err
	}

	return sNet.gasLimit, nil
}

// dbRegisterSubNetwork stores mappings from ID of the sub-network to the sub-network data.
func dbRegisterSubNetwork(dbTx database.Tx, subNetworkID *subnetworkid.SubNetworkID, network *subNetwork) error {
	// Serialize the sub-network
	serializedSubNetwork, err := serializeSubNetwork(network)
	if err != nil {
		return fmt.Errorf("failed to serialize sub-netowrk '%s': %s", subNetworkID, err)
	}

	// Store the sub-network
	subNetworksBucket := dbTx.Metadata().Bucket(subNetworksBucketName)
	err = subNetworksBucket.Put(subNetworkID[:], serializedSubNetwork)
	if err != nil {
		return fmt.Errorf("failed to write sub-netowrk '%s': %s", subNetworkID, err)
	}

	return nil
}

func dbGetSubNetwork(dbTx database.Tx, subNetworkID *subnetworkid.SubNetworkID) (*subNetwork, error) {
	// Get the sub-network
	bucket := dbTx.Metadata().Bucket(subNetworksBucketName)
	serializedSubNetwork := bucket.Get(subNetworkID[:])
	if serializedSubNetwork == nil {
		return nil, fmt.Errorf("sub-network '%d' not found", subNetworkID)
	}

	return deserializeSubNetwork(serializedSubNetwork)
}

type subNetwork struct {
	gasLimit uint64
}

func newSubNetwork(tx *wire.MsgTx) *subNetwork {
	gasLimit := binary.LittleEndian.Uint64(tx.Payload[:8])

	return &subNetwork{
		gasLimit: gasLimit,
	}
}

// serializeSubNetwork serializes a subNetwork into the following binary format:
// | gasLimit (8 bytes) |
func serializeSubNetwork(sNet *subNetwork) ([]byte, error) {
	serializedSNet := bytes.NewBuffer(make([]byte, 0, 8))

	// Write the gas limit
	err := binary.Write(serializedSNet, byteOrder, sNet.gasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize sub-network: %s", err)
	}

	return serializedSNet.Bytes(), nil
}

// deserializeSubNetwork deserializes a byte slice into a subNetwork.
// See serializeSubNetwork for the binary format.
func deserializeSubNetwork(serializedSNetBytes []byte) (*subNetwork, error) {
	serializedSNet := bytes.NewBuffer(serializedSNetBytes)

	// Read the gas limit
	var gasLimit uint64
	err := binary.Read(serializedSNet, byteOrder, &gasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize sub-network: %s", err)
	}

	return &subNetwork{
		gasLimit: gasLimit,
	}, nil
}
