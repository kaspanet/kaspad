package consensushashing

import (
	"fmt"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
)

func TestTransactionHash(t *testing.T) {
	tx := externalapi.DomainTransaction{0, []*externalapi.DomainTransactionInput{}, []*externalapi.DomainTransactionOutput{}, 0,
		externalapi.DomainSubnetworkID{}, 0, []byte{}, 0, 0,
		nil}
	id := TransactionID(&tx)
	fmt.Printf("%s\n", id)
	tx_hash := TransactionHash(&tx)
	fmt.Printf("%s\n\n", tx_hash)

	inputs := []*externalapi.DomainTransactionInput{&externalapi.DomainTransactionInput{
		PreviousOutpoint: externalapi.DomainOutpoint{
			TransactionID: externalapi.DomainTransactionID{},
			Index:         2,
		},
		SignatureScript: []byte{1, 2},
		Sequence:        7,
		SigOpCount:      5,
		UTXOEntry:       nil,
	}}

	tx = externalapi.DomainTransaction{1, inputs, []*externalapi.DomainTransactionOutput{}, 0,
		externalapi.DomainSubnetworkID{}, 0, []byte{}, 0, 0,
		nil}
	id = TransactionID(&tx)
	fmt.Printf("%s\n", id)
	tx_hash = TransactionHash(&tx)
	fmt.Printf("%s\n\n", tx_hash)

	outputs := []*externalapi.DomainTransactionOutput{&externalapi.DomainTransactionOutput{
		Value:           1564,
		ScriptPublicKey: &externalapi.ScriptPublicKey{
			Script:  []byte{1, 2, 3, 4, 5},
			Version: 7,
		},
	}}

	tx = externalapi.DomainTransaction{1, inputs, outputs, 0,
		externalapi.DomainSubnetworkID{}, 0, []byte{}, 0, 0,
		nil}
	id = TransactionID(&tx)
	fmt.Printf("%s\n", id)
	tx_hash = TransactionHash(&tx)
	fmt.Printf("%s\n\n", tx_hash)

	tx = externalapi.DomainTransaction{2, inputs, outputs, 54,
		externalapi.DomainSubnetworkID{}, 3, []byte{}, 4, 7,
		nil}
	id = TransactionID(&tx)
	fmt.Printf("%s\n", id)
	tx_hash = TransactionHash(&tx)
	fmt.Printf("%s\n\n", tx_hash)

	transactionId, err := externalapi.NewDomainHashFromString("59b3d6dc6cdc660c389c3fdb5704c48c598d279cdf1bab54182db586a4c95dd5")
	if err != nil {
		t.Fatalf("%s", err)
	}

	inputs = []*externalapi.DomainTransactionInput{&externalapi.DomainTransactionInput{
		PreviousOutpoint: externalapi.DomainOutpoint{
			TransactionID: externalapi.DomainTransactionID(*transactionId),
			Index:         2,
		},
		SignatureScript: []byte{1, 2},
		Sequence:        7,
		SigOpCount:      5,
		UTXOEntry:       nil,
	}}

	tx = externalapi.DomainTransaction{2, inputs, outputs, 54,
		externalapi.DomainSubnetworkID{}, 3, []byte{}, 4, 7,
		nil}
	id = TransactionID(&tx)
	fmt.Printf("%s\n", id)
	tx_hash = TransactionHash(&tx)
	fmt.Printf("%s\n\n", tx_hash)

	tx = externalapi.DomainTransaction{2, inputs, outputs, 54,
		subnetworks.SubnetworkIDCoinbase, 3, []byte{}, 4, 7,
		nil}
	id = TransactionID(&tx)
	fmt.Printf("%s\n", id)
	tx_hash = TransactionHash(&tx)
	fmt.Printf("%s\n\n", tx_hash)

	tx = externalapi.DomainTransaction{2, inputs, outputs, 54,
		subnetworks.SubnetworkIDRegistry, 3, []byte{}, 4, 7,
		nil}
	id = TransactionID(&tx)
	fmt.Printf("%s\n", id)
	tx_hash = TransactionHash(&tx)
	fmt.Printf("%s\n\n", tx_hash)
}
