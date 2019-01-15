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

// registerSubnetworks scans a list of accepted transactions, singles out
// subnetwork registry transactions, validates them, and registers a new
// subnetwork based on it.
// This function returns an error if one or more transactions are invalid
func registerSubnetworks(dbTx database.Tx, txs []*TxWithBlockHash) error {
	validSubnetworkRegistryTxs := make([]*wire.MsgTx, 0)
	for _, txData := range txs {
		tx := txData.Tx.MsgTx()
		if tx.SubnetworkID == wire.SubnetworkRegistry {
			err := validateSubnetworkRegistryTransaction(tx)
			if err != nil {
				return err
			}
			validSubnetworkRegistryTxs = append(validSubnetworkRegistryTxs, tx)
		}

		if subnetworkid.Less(&wire.SubnetworkRegistry, &tx.SubnetworkID) {
			// Transactions are ordered by subnetwork, so we can safely assume
			// that the rest of the transactions will not be subnetwork registry
			// transactions.
			break
		}
	}

	for _, registryTx := range validSubnetworkRegistryTxs {
		subnetworkID, err := txToSubnetworkID(registryTx)
		if err != nil {
			return err
		}
		sNet, err := dbGetSubnetwork(dbTx, subnetworkID)
		if err != nil {
			return err
		}
		if sNet == nil {
			createdSubnetwork := newSubnetwork(registryTx)
			err := dbRegisterSubnetwork(dbTx, subnetworkID, createdSubnetwork)
			if err != nil {
				return fmt.Errorf("failed registering subnetwork"+
					"for tx '%s': %s", registryTx.TxHash(), err)
			}
		}
	}

	return nil
}

// validateSubnetworkRegistryTransaction makes sure that a given subnetwork registry
// transaction is valid. Such a transaction is valid iff:
// - Its entire payload is a uint64 (8 bytes)
func validateSubnetworkRegistryTransaction(tx *wire.MsgTx) error {
	if len(tx.Payload) != 8 {
		return ruleError(ErrSubnetworkRegistry, fmt.Sprintf("validation failed: subnetwork registry"+
			"tx '%s' has an invalid payload", tx.TxHash()))
	}

	return nil
}

// txToSubnetworkID creates a subnetwork ID from a subnetwork registry transaction
func txToSubnetworkID(tx *wire.MsgTx) (*subnetworkid.SubnetworkID, error) {
	txHash := tx.TxHash()
	return subnetworkid.New(util.Hash160(txHash[:]))
}

// subnetwork returns a registered subnetwork. If the subnetwork does not exist
// this method returns an error.
func (dag *BlockDAG) subnetwork(subnetworkID *subnetworkid.SubnetworkID) (*subnetwork, error) {
	var sNet *subnetwork
	var err error
	dbErr := dag.db.View(func(dbTx database.Tx) error {
		sNet, err = dbGetSubnetwork(dbTx, subnetworkID)
		return nil
	})
	if dbErr != nil {
		return nil, fmt.Errorf("could not retrieve subnetwork '%d': %s", subnetworkID, dbErr)
	}
	if err != nil {
		return nil, fmt.Errorf("could not retrieve subnetwork '%d': %s", subnetworkID, err)
	}

	return sNet, nil
}

// GasLimit returns the gas limit of a registered subnetwork. If the subnetwork does not
// exist this method returns an error.
func (dag *BlockDAG) GasLimit(subnetworkID *subnetworkid.SubnetworkID) (uint64, error) {
	sNet, err := dag.subnetwork(subnetworkID)
	if err != nil {
		return 0, err
	}
	if sNet == nil {
		return 0, fmt.Errorf("subnetwork '%s' not found", subnetworkID)
	}

	return sNet.gasLimit, nil
}

// dbRegisterSubnetwork stores mappings from ID of the subnetwork to the subnetwork data.
func dbRegisterSubnetwork(dbTx database.Tx, subnetworkID *subnetworkid.SubnetworkID, network *subnetwork) error {
	// Serialize the subnetwork
	serializedSubnetwork, err := serializeSubnetwork(network)
	if err != nil {
		return fmt.Errorf("failed to serialize sub-netowrk '%s': %s", subnetworkID, err)
	}

	// Store the subnetwork
	subnetworksBucket := dbTx.Metadata().Bucket(subnetworksBucketName)
	err = subnetworksBucket.Put(subnetworkID[:], serializedSubnetwork)
	if err != nil {
		return fmt.Errorf("failed to write sub-netowrk '%s': %s", subnetworkID, err)
	}

	return nil
}

// dbGetSubnetwork returns the subnetwork associated with subnetworkID or nil if the subnetwork was not found.
func dbGetSubnetwork(dbTx database.Tx, subnetworkID *subnetworkid.SubnetworkID) (*subnetwork, error) {
	bucket := dbTx.Metadata().Bucket(subnetworksBucketName)
	serializedSubnetwork := bucket.Get(subnetworkID[:])
	if serializedSubnetwork == nil {
		return nil, nil
	}

	return deserializeSubnetwork(serializedSubnetwork)
}

type subnetwork struct {
	gasLimit uint64
}

func newSubnetwork(tx *wire.MsgTx) *subnetwork {
	gasLimit := binary.LittleEndian.Uint64(tx.Payload[:8])

	return &subnetwork{
		gasLimit: gasLimit,
	}
}

// serializeSubnetwork serializes a subnetwork into the following binary format:
// | gasLimit (8 bytes) |
func serializeSubnetwork(sNet *subnetwork) ([]byte, error) {
	serializedSNet := bytes.NewBuffer(make([]byte, 0, 8))

	// Write the gas limit
	err := binary.Write(serializedSNet, byteOrder, sNet.gasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize subnetwork: %s", err)
	}

	return serializedSNet.Bytes(), nil
}

// deserializeSubnetwork deserializes a byte slice into a subnetwork.
// See serializeSubnetwork for the binary format.
func deserializeSubnetwork(serializedSNetBytes []byte) (*subnetwork, error) {
	serializedSNet := bytes.NewBuffer(serializedSNetBytes)

	// Read the gas limit
	var gasLimit uint64
	err := binary.Read(serializedSNet, byteOrder, &gasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize subnetwork: %s", err)
	}

	return &subnetwork{
		gasLimit: gasLimit,
	}, nil
}
