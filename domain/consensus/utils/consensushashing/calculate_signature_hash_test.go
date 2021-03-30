package consensushashing_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/kaspanet/go-secp256k1"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// shortened versions of SigHash types to fit in single line of test case
const (
	all                = consensushashing.SigHashAll
	none               = consensushashing.SigHashNone
	single             = consensushashing.SigHashSingle
	allAnyoneCanPay    = consensushashing.SigHashAll | consensushashing.SigHashAnyOneCanPay
	noneAnyoneCanPay   = consensushashing.SigHashNone | consensushashing.SigHashAnyOneCanPay
	singleAnyoneCanPay = consensushashing.SigHashSingle | consensushashing.SigHashAnyOneCanPay
)

func modifyOutput(outputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		clone.Outputs[outputIndex].Value = 100
		return clone
	}
}

func modifyInput(inputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		clone.Inputs[inputIndex].PreviousOutpoint.Index = 2
		return clone
	}
}

func modifyAmountSpent(inputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		utxoEntry := clone.Inputs[inputIndex].UTXOEntry
		clone.Inputs[inputIndex].UTXOEntry = utxo.NewUTXOEntry(666, utxoEntry.ScriptPublicKey(), false, 100)
		return clone
	}
}

func modifyScriptPublicKey(inputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		utxoEntry := clone.Inputs[inputIndex].UTXOEntry
		scriptPublicKey := utxoEntry.ScriptPublicKey()
		scriptPublicKey.Script = append(scriptPublicKey.Script, 1, 2, 3)
		clone.Inputs[inputIndex].UTXOEntry = utxo.NewUTXOEntry(utxoEntry.Amount(), scriptPublicKey, false, 100)
		return clone
	}
}

func modifySequence(inputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		clone.Inputs[inputIndex].Sequence = 12345
		return clone
	}
}

func modifyPayload(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	clone := tx.Clone()
	clone.Payload = []byte{6, 6, 6, 4, 2, 0, 1, 3, 3, 7}
	return clone
}

func modifyGas(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	clone := tx.Clone()
	clone.Gas = 1234
	return clone
}

func modifySubnetworkID(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	clone := tx.Clone()
	clone.SubnetworkID = externalapi.DomainSubnetworkID{6, 6, 6, 4, 2, 0, 1, 3, 3, 7}
	return clone
}

func TestCalculateSignatureHash(t *testing.T) {
	nativeTx, subnetworkTx, err := generateTxs()
	if err != nil {
		t.Fatalf("Error from generateTxs: %+v", err)
	}

	// Note: Expected values were generated by the same code that they test,
	// As long as those were not verified using 3rd-party code they only check for regression, not correctness
	tests := []struct {
		name                  string
		tx                    *externalapi.DomainTransaction
		hashType              consensushashing.SigHashType
		inputIndex            int
		modificationFunction  func(*externalapi.DomainTransaction) *externalapi.DomainTransaction
		expectedSignatureHash string
	}{
		// native transactions

		// sigHashAll
		{name: "native-all-0", tx: nativeTx, hashType: all, inputIndex: 0,
			expectedSignatureHash: "c899a5ea7414f0bbfd77e50674f46da34ce8722b928d4362a4b4b727c69c6499"},
		{name: "native-all-0-modify-input-1", tx: nativeTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyInput(1), // should change the hash
			expectedSignatureHash: "faf3b9db2e07b1c14b2df02002d3e40f1e430f177ac5cd3354c84dad8fbe72ce"},
		{name: "native-all-0-modify-output-1", tx: nativeTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyOutput(1), // should change the hash
			expectedSignatureHash: "3a557c5b873aab72dcb81649642e1d7a63b75dcdcc74e19d340964a9e0eac76c"},
		{name: "native-all-0-modify-sequence-1", tx: nativeTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifySequence(1), // should change the hash
			expectedSignatureHash: "2dd5fe8f9fa4bf551ea2f080a26e07b2462083e12d3b2ed01cb9369a61920665"},
		{name: "native-all-anyonecanpay-0", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			expectedSignatureHash: "19fe2e0db681017f318fda705a39bbbad9c1085514cfbcff6fac01e1725f758b"},
		{name: "native-all-anyonecanpay-0-modify-input-0", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyInput(0), // should change the hash
			expectedSignatureHash: "5b21d492560a1c794595f769b3ae3c151775b9cfc4029d17c53f1856e1005da4"},
		{name: "native-all-anyonecanpay-0-modify-input-1", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyInput(1), // shouldn't change the hash
			expectedSignatureHash: "19fe2e0db681017f318fda705a39bbbad9c1085514cfbcff6fac01e1725f758b"},
		{name: "native-all-anyonecanpay-0-modify-sequence", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifySequence(1), // shouldn't change the hash
			expectedSignatureHash: "19fe2e0db681017f318fda705a39bbbad9c1085514cfbcff6fac01e1725f758b"},

		// sigHashNone
		{name: "native-none-0", tx: nativeTx, hashType: none, inputIndex: 0,
			expectedSignatureHash: "fafabaabf6349fee4e18626b4eff015472f2317576a8f4bf7b0eea1df6f3e32b"},
		{name: "native-none-0-modify-output-1", tx: nativeTx, hashType: none, inputIndex: 0,
			modificationFunction:  modifyOutput(1), // shouldn't change the hash
			expectedSignatureHash: "fafabaabf6349fee4e18626b4eff015472f2317576a8f4bf7b0eea1df6f3e32b"},
		{name: "native-none-0-modify-sequence-0", tx: nativeTx, hashType: none, inputIndex: 0,
			modificationFunction:  modifySequence(0), // should change the hash
			expectedSignatureHash: "daee0700e0ed4ab9f50de24d83e0bfce62999474ec8ceeb537ea35980662b601"},
		{name: "native-none-0-modify-sequence-1", tx: nativeTx, hashType: none, inputIndex: 0,
			modificationFunction:  modifySequence(1), // shouldn't change the hash
			expectedSignatureHash: "fafabaabf6349fee4e18626b4eff015472f2317576a8f4bf7b0eea1df6f3e32b"},
		{name: "native-none-anyonecanpay-0", tx: nativeTx, hashType: noneAnyoneCanPay, inputIndex: 0,
			expectedSignatureHash: "4e5c2d895f9711dc89c19d49ba478e9c8f4be0d82c9bd6b60d0361eb9b5296bc"},
		{name: "native-none-anyonecanpay-0-modify-amount-spent", tx: nativeTx, hashType: noneAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyAmountSpent(0), // should change the hash
			expectedSignatureHash: "9ce2f75eafc85b8e19133942c3143d14b61f2e7cc479fbc6d2fca026e50897f1"},
		{name: "native-none-anyonecanpay-0-modify-script-public-key", tx: nativeTx, hashType: noneAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyScriptPublicKey(0), // should change the hash
			expectedSignatureHash: "c6c364190520fe6c0419c2f45e25bf084356333b03ac7aaec28251126398bda3"},

		// sigHashSingle
		{name: "native-single-0", tx: nativeTx, hashType: single, inputIndex: 0,
			expectedSignatureHash: "6ff01d5d7cd82e24bc9ca0edec8bd6931ffb5aa1d303f07ca05dc89757343a92"},
		{name: "native-single-0-modify-output-0", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifyOutput(0), // should change the hash
			expectedSignatureHash: "d62af956aea369365bacc7e7f1aac106836994f1648311e82dd38da822c8771e"},
		{name: "native-single-0-modify-output-1", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifyOutput(1), // shouldn't change the hash
			expectedSignatureHash: "6ff01d5d7cd82e24bc9ca0edec8bd6931ffb5aa1d303f07ca05dc89757343a92"},
		{name: "native-single-0-modify-sequence-0", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifySequence(0), // should change the hash
			expectedSignatureHash: "46692229d45bf2ceacb18960faba29753e325c0ade26ecf94495b91daacb828d"},
		{name: "native-single-0-modify-sequence-1", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifySequence(1), // shouldn't change the hash
			expectedSignatureHash: "6ff01d5d7cd82e24bc9ca0edec8bd6931ffb5aa1d303f07ca05dc89757343a92"},
		{name: "native-single-2-no-corresponding-output", tx: nativeTx, hashType: single, inputIndex: 2,
			expectedSignatureHash: "d3cc385082a7f272ec2c8aae7f3a96ab2f49a4a4e1ed44d61af34058a7721281"},
		{name: "native-single-2-no-corresponding-output-modify-output-1", tx: nativeTx, hashType: single, inputIndex: 2,
			modificationFunction:  modifyOutput(1), // shouldn't change the hash
			expectedSignatureHash: "d3cc385082a7f272ec2c8aae7f3a96ab2f49a4a4e1ed44d61af34058a7721281"},
		{name: "native-single-anyonecanpay-0", tx: nativeTx, hashType: singleAnyoneCanPay, inputIndex: 0,
			expectedSignatureHash: "408fcfd8ceca135c0f54569ccf8ac727e1aa6b5a15f87ccca765f1d5808aa4ea"},
		{name: "native-single-anyonecanpay-2-no-corresponding-output", tx: nativeTx, hashType: singleAnyoneCanPay, inputIndex: 2,
			expectedSignatureHash: "685fac0d0b9dd3c5556f266714c4f7f93475d49fa12befb18e8297bc062aeaba"},

		// subnetwork transaction
		{name: "subnetwork-all-0", tx: subnetworkTx, hashType: all, inputIndex: 0,
			expectedSignatureHash: "0e8b1433b761a220a61c0dc1f0fda909d49cef120d98d9f87344fef52dac0d8b"},
		{name: "subnetwork-all-modify-payload", tx: subnetworkTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyPayload, // should change the hash
			expectedSignatureHash: "087315acb9193eaa14929dbe3d0ace80238aebe13eab3bf8db6c0a0d7ddb782e"},
		{name: "subnetwork-all-modify-gas", tx: subnetworkTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyGas, // should change the hash
			expectedSignatureHash: "07a90408ef45864ae8354b07a74cf826a4621391425ba417470a6e680af4ce70"},
		{name: "subnetwork-all-subnetwork-id", tx: subnetworkTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifySubnetworkID, // should change the hash
			expectedSignatureHash: "4ca44c2e35729ae5efe831a77027f1a58a41dbdd853459c26cbfe7d6c88783fb"},
	}

	for _, test := range tests {
		tx := test.tx
		if test.modificationFunction != nil {
			tx = test.modificationFunction(tx)
		}

		actualSignatureHash, err := consensushashing.CalculateSignatureHash(
			tx, test.inputIndex, test.hashType, &consensushashing.SighashReusedValues{})
		if err != nil {
			t.Errorf("%s: Error from CalculateSignatureHash: %+v", test.name, err)
			continue
		}

		if actualSignatureHash.String() != test.expectedSignatureHash {
			t.Errorf("%s: expected signature hash: '%s'; but got: '%s'",
				test.name, test.expectedSignatureHash, actualSignatureHash)
		}
	}
}

func generateTxs() (nativeTx, subnetworkTx *externalapi.DomainTransaction, err error) {
	genesisCoinbase := dagconfig.SimnetParams.GenesisBlock.Transactions[0]
	genesisCoinbaseTransactionID := consensushashing.TransactionID(genesisCoinbase)

	address1Str := "kaspasim:qzpj2cfa9m40w9m2cmr8pvfuqpp32mzzwsuw6ukhfduqpp32mzzws59e8fapc"
	address1, err := util.DecodeAddress(address1Str, util.Bech32PrefixKaspaSim)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding address1: %+v", err)
	}
	address1ToScript, err := txscript.PayToAddrScript(address1)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating script: %+v", err)
	}

	address2Str := "kaspasim:qr7w7nqsdnc3zddm6u8s9fex4ysk95hm3v30q353ymuqpp32mzzws59e8fapc"
	address2, err := util.DecodeAddress(address2Str, util.Bech32PrefixKaspaSim)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding address2: %+v", err)
	}
	address2ToScript, err := txscript.PayToAddrScript(address2)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating script: %+v", err)
	}

	txIns := []*externalapi.DomainTransactionInput{
		{
			PreviousOutpoint: *externalapi.NewDomainOutpoint(genesisCoinbaseTransactionID, 0),
			Sequence:         0,
			UTXOEntry:        utxo.NewUTXOEntry(100, address1ToScript, false, 0),
		},
		{
			PreviousOutpoint: *externalapi.NewDomainOutpoint(genesisCoinbaseTransactionID, 1),
			Sequence:         1,
			UTXOEntry:        utxo.NewUTXOEntry(200, address2ToScript, false, 0),
		},
		{
			PreviousOutpoint: *externalapi.NewDomainOutpoint(genesisCoinbaseTransactionID, 2),
			Sequence:         2,
			UTXOEntry:        utxo.NewUTXOEntry(300, address2ToScript, false, 0),
		},
	}

	txOuts := []*externalapi.DomainTransactionOutput{
		{
			Value:           300,
			ScriptPublicKey: address2ToScript,
		},
		{
			Value:           300,
			ScriptPublicKey: address1ToScript,
		},
	}

	nativeTx = &externalapi.DomainTransaction{
		Version:      0,
		Inputs:       txIns,
		Outputs:      txOuts,
		LockTime:     1615462089000,
		SubnetworkID: externalapi.DomainSubnetworkID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	subnetworkTx = &externalapi.DomainTransaction{
		Version:      0,
		Inputs:       txIns,
		Outputs:      txOuts,
		LockTime:     1615462089000,
		SubnetworkID: externalapi.DomainSubnetworkID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Gas:          250,
		Payload:      []byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}

	return nativeTx, subnetworkTx, nil
}

func BenchmarkCalculateSignatureHash(b *testing.B) {
	sigHashTypes := []consensushashing.SigHashType{
		consensushashing.SigHashAll,
		consensushashing.SigHashNone,
		consensushashing.SigHashSingle,
		consensushashing.SigHashAll | consensushashing.SigHashAnyOneCanPay,
		consensushashing.SigHashNone | consensushashing.SigHashAnyOneCanPay,
		consensushashing.SigHashSingle | consensushashing.SigHashAnyOneCanPay}

	for _, size := range []int{10, 100, 1000} {
		tx := generateTransaction(b, sigHashTypes, size)

		b.Run(fmt.Sprintf("%d-inputs-and-outputs", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reusedValues := &consensushashing.SighashReusedValues{}
				for inputIndex := range tx.Inputs {
					sigHashType := sigHashTypes[inputIndex%len(sigHashTypes)]
					_, err := consensushashing.CalculateSignatureHash(tx, inputIndex, sigHashType, reusedValues)
					if err != nil {
						b.Fatalf("Error from CalculateSignatureHash: %+v", err)
					}
				}
			}
		})
	}
}

func generateTransaction(b *testing.B, sigHashTypes []consensushashing.SigHashType, inputAndOutputSizes int) *externalapi.DomainTransaction {
	sourceScript := getSourceScript(b)
	tx := &externalapi.DomainTransaction{
		Version:      0,
		Inputs:       generateInputs(inputAndOutputSizes, sourceScript),
		Outputs:      generateOutputs(inputAndOutputSizes, sourceScript),
		LockTime:     123456789,
		SubnetworkID: externalapi.DomainSubnetworkID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Gas:          125,
		Payload:      []byte{9, 8, 7, 6, 5, 4, 3, 2, 1},
		Fee:          0,
		Mass:         0,
		ID:           nil,
	}
	signTx(b, tx, sigHashTypes)
	return tx
}

func signTx(b *testing.B, tx *externalapi.DomainTransaction, sigHashTypes []consensushashing.SigHashType) {
	sourceAddressPKStr := "a4d85b7532123e3dd34e58d7ce20895f7ca32349e29b01700bb5a3e72d2570eb"
	privateKeyBytes, err := hex.DecodeString(sourceAddressPKStr)
	if err != nil {
		b.Fatalf("Error parsing private key hex: %+v", err)
	}
	keyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		b.Fatalf("Error deserializing private key: %+v", err)
	}
	for i, txIn := range tx.Inputs {
		signatureScript, err := txscript.SignatureScript(
			tx, i, sigHashTypes[i%len(sigHashTypes)], keyPair, &consensushashing.SighashReusedValues{})
		if err != nil {
			b.Fatalf("Error from SignatureScript: %+v", err)
		}
		txIn.SignatureScript = signatureScript
	}

}

func generateInputs(size int, sourceScript *externalapi.ScriptPublicKey) []*externalapi.DomainTransactionInput {
	inputs := make([]*externalapi.DomainTransactionInput, size)

	for i := 0; i < size; i++ {
		inputs[i] = &externalapi.DomainTransactionInput{
			PreviousOutpoint: *externalapi.NewDomainOutpoint(
				externalapi.NewDomainTransactionIDFromByteArray(&[32]byte{12, 3, 4, 5}), 1),
			SignatureScript: nil,
			Sequence:        uint64(i),
			UTXOEntry:       utxo.NewUTXOEntry(uint64(i), sourceScript, false, 12),
		}
	}

	return inputs
}

func getSourceScript(b *testing.B) *externalapi.ScriptPublicKey {
	sourceAddressStr := "kaspasim:qz6f9z6l3x4v3lf9mgf0t934th4nx5kgzu663x9yjh"

	sourceAddress, err := util.DecodeAddress(sourceAddressStr, util.Bech32PrefixKaspaSim)
	if err != nil {
		b.Fatalf("Error from DecodeAddress: %+v", err)
	}

	sourceScript, err := txscript.PayToAddrScript(sourceAddress)
	if err != nil {
		b.Fatalf("Error from PayToAddrScript: %+v", err)
	}
	return sourceScript
}

func generateOutputs(size int, script *externalapi.ScriptPublicKey) []*externalapi.DomainTransactionOutput {
	outputs := make([]*externalapi.DomainTransactionOutput, size)

	for i := 0; i < size; i++ {
		outputs[i] = &externalapi.DomainTransactionOutput{
			Value:           uint64(i),
			ScriptPublicKey: script,
		}
	}

	return outputs
}
