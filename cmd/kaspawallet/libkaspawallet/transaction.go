package libkaspawallet

import (
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/bip32"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/txmass"
	"github.com/pkg/errors"
)

// Payment contains a recipient payment details
type Payment struct {
	Address util.Address
	Amount  uint64
}

// UTXO is a type that stores a UTXO and meta data
// that is needed in order to sign it and create
// transactions with it.
type UTXO struct {
	Outpoint       *externalapi.DomainOutpoint
	UTXOEntry      externalapi.UTXOEntry
	DerivationPath string
}

// CreateUnsignedTransaction creates an unsigned transaction
func CreateUnsignedTransaction(
	extendedPublicKeys []string,
	minimumSignatures uint32,
	payments []*Payment,
	selectedUTXOs []*UTXO) ([]byte, error) {

	sortPublicKeys(extendedPublicKeys)
	unsignedTransaction, err := createUnsignedTransaction(extendedPublicKeys, minimumSignatures, payments, selectedUTXOs)
	if err != nil {
		return nil, err
	}

	return serialization.SerializePartiallySignedTransaction(unsignedTransaction)
}

func multiSigRedeemScript(extendedPublicKeys []string, minimumSignatures uint32, path string, ecdsa bool) ([]byte, error) {
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddInt64(int64(minimumSignatures))
	for _, key := range extendedPublicKeys {
		extendedKey, err := bip32.DeserializeExtendedKey(key)
		if err != nil {
			return nil, err
		}

		derivedKey, err := extendedKey.DeriveFromPath(path)
		if err != nil {
			return nil, err
		}

		publicKey, err := derivedKey.PublicKey()
		if err != nil {
			return nil, err
		}

		var serializedPublicKey []byte
		if ecdsa {
			serializedECDSAPublicKey, err := publicKey.Serialize()
			if err != nil {
				return nil, err
			}
			serializedPublicKey = serializedECDSAPublicKey[:]
		} else {
			schnorrPublicKey, err := publicKey.ToSchnorr()
			if err != nil {
				return nil, err
			}

			serializedSchnorrPublicKey, err := schnorrPublicKey.Serialize()
			if err != nil {
				return nil, err
			}
			serializedPublicKey = serializedSchnorrPublicKey[:]
		}

		scriptBuilder.AddData(serializedPublicKey)
	}
	scriptBuilder.AddInt64(int64(len(extendedPublicKeys)))

	if ecdsa {
		scriptBuilder.AddOp(txscript.OpCheckMultiSigECDSA)
	} else {
		scriptBuilder.AddOp(txscript.OpCheckMultiSig)
	}

	return scriptBuilder.Script()
}

//sign transaction with private key only
//fee per input currently unused / awaiting fee structure
func CreateUnsignedTransactionWithSchnorrPublicKey(selectedUTXOs []*UTXO, publicKey string, payments []*Payment, changeAdress util.Address, feePerInput uint64) (*serialization.PartiallySignedTransaction, error) {

	if len(payments) == 0 {
		return nil, errors.Errorf("Cannot create a transaction without payment details")
	}

	if len(selectedUTXOs) == 0 {
		return nil, errors.Errorf("Cannot create a transaction without selection of UTXOs")
	}

	totalfee := uint64(len(selectedUTXOs)) * feePerInput
	totalPay := uint64(0)
	totalTransaction := uint64(0)

	for _, payment := range payments {
		totalPay = totalPay + payment.Amount
	}
	if totalPay <= totalfee {
		return nil, errors.Errorf("Total send amount of ", totalPay, " is less then the required fee of ", totalfee)
	}

	//since we know the sizes a priori, do not use commitPaymentToUnsignedTransactionForSchnorrPrivateKey function!
	outputs := make([]*externalapi.DomainTransactionOutput, len(payments)+1) //plus one for change
	for i, payment := range payments[:1] {
		ScriptPublicKey, err := txscript.PayToAddrScript(payment.Address)
		if err != nil {
			return nil, err
		}
		outputs[i] = &externalapi.DomainTransactionOutput{
			Value:           payment.Amount - (payment.Amount/totalPay)*totalfee, // pay fee fractionally
			ScriptPublicKey: ScriptPublicKey,
		}
	}

	//since we know the sizes a priori, do not use the commitUTXOsToUnsignedTransactionForScnorrPrivateKey function!
	inputs := make([]*externalapi.DomainTransactionInput, len(selectedUTXOs))
	partiallySignedInputs := make([]*serialization.PartiallySignedInput, len(selectedUTXOs))
	for i, UTXO := range selectedUTXOs {
		inputs[i] = &externalapi.DomainTransactionInput{
			PreviousOutpoint: *UTXO.Outpoint,
			SigOpCount:       1,
		}

		partiallySignedInputs[i] = &serialization.PartiallySignedInput{
			PrevOutput: &externalapi.DomainTransactionOutput{
				Value:           UTXO.UTXOEntry.Amount(),
				ScriptPublicKey: UTXO.UTXOEntry.ScriptPublicKey(),
			},
			//What do I define as the ExtendedPublicKey for serialization.PartiallySignedInput.PubKeySignaturePairs here,
			//I cannot extract domainTransactions with libkaspawallet FromDe/Ser - ToDe/Ser etc.. functions without it..
		}
		totalTransaction = totalTransaction + UTXO.UTXOEntry.Amount()
	}

	//deal with special case of changeaddress
	ScriptPublicKey, err := txscript.PayToAddrScript(changeAdress)
	if err != nil {
		return nil, err
	}

	if totalTransaction-feePerInput > totalPay {
		outputs[len(outputs)] = &externalapi.DomainTransactionOutput{
			Value:           totalTransaction - totalPay, // pay fee fractionally
			ScriptPublicKey: ScriptPublicKey,
		}
	} else {
		outputs = outputs[:1]
	}

	partiallySignedTransaction := &serialization.PartiallySignedTransaction{
		Tx: &externalapi.DomainTransaction{
			Version:      constants.MaxTransactionVersion,
			Inputs:       inputs,
			Outputs:      outputs,
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
			Gas:          0,
			Payload:      nil,
		},
		PartiallySignedInputs: partiallySignedInputs,
	}

	//Commented out the below: serializing gets rid of the UTXO information, needed to split the transaction
	//return serialization.SerializePartiallySignedTransaction(partiallySignedTransaction)
	return partiallySignedTransaction, nil
}

//for use in compounding where we do not know the size and most expand and add with each iteration
func commitUTXOToUnsignedTransactionForScnorrPrivateKey(
	partiallySignedTransaction *serialization.PartiallySignedTransaction,
	UTXO *UTXO,
) error {
	partiallySignedTransaction.Tx.Inputs = append(
		partiallySignedTransaction.Tx.Inputs,
		&externalapi.DomainTransactionInput{
			PreviousOutpoint: *UTXO.Outpoint,
			SigOpCount:       1,
		},
	)
	partiallySignedTransaction.PartiallySignedInputs = append(
		partiallySignedTransaction.PartiallySignedInputs,
		&serialization.PartiallySignedInput{
			PrevOutput: &externalapi.DomainTransactionOutput{
				Value:           UTXO.UTXOEntry.Amount(),
				ScriptPublicKey: UTXO.UTXOEntry.ScriptPublicKey(),
			},
		},
	)

	return nil
}

//for use in compounding where we do not know the size and most expand and add with each iteration
//this commit increases outputs to
func commitOutputToUnsignedTransactionForSchnorrPrivateKey(
	partiallySignedTransaction *serialization.PartiallySignedTransaction,
	newOutput *externalapi.DomainTransactionOutput) error {
	partiallySignedTransaction.Tx.Outputs = append(
		partiallySignedTransaction.Tx.Outputs,
		newOutput,
	)
	return nil
}

//create a new partiallysignedtransaction with no inputs or outputs, using a template of an already exisiting partiallysignedtransaction
func newDummyTransaction(template *serialization.PartiallySignedTransaction) *serialization.PartiallySignedTransaction {
	return &serialization.PartiallySignedTransaction{
		Tx: &externalapi.DomainTransaction{
			Version:      template.Tx.Version,
			LockTime:     template.Tx.LockTime,
			SubnetworkID: template.Tx.SubnetworkID,
			Gas:          template.Tx.Gas,
			Payload:      template.Tx.Payload,
		},
		PartiallySignedInputs: template.PartiallySignedInputs,
	}
}

//Compounds by iterating each input, and checking for violation of TransactionMass, if it occurs, commits a split with the previous inputs.
//outputs are sent as a whole to the payment address, leftovers are sent to the change address via the last split.
//function is intended to be used for transactions signed with PrivatKeyOnly, untested otherwise.
func CompoundUnsignedTransactionByMaxMassForScnorrPrivateKey(params *dagconfig.Params, partiallySignedTransaction *serialization.PartiallySignedTransaction,
	payments []*Payment, changeAddress util.Address, feePerInput int) ([]*serialization.PartiallySignedTransaction, error) {

	if len(payments) > 1 {
		return nil, errors.Errorf("Compounding is currently only valid for 2 addresses Max")
	}

	var splitTransactions []*serialization.PartiallySignedTransaction

	//TODO: give more room to violate, I think the problem is in txmass calculater, or elsewhere
	massleeway := uint64(7000)
	massCalculater := txmass.NewCalculator(params.MassPerTxByte, params.MassPerScriptPubKeyByte, params.MassPerSigOp)
	if massCalculater.CalculateTransactionMass(partiallySignedTransaction.Tx)+massleeway <= mempool.MaximumStandardTransactionMass {
		return append(splitTransactions, partiallySignedTransaction), nil
	}

	//for refernece this keeps track in respect to dummyWindow[0], not dummyWindow[1]
	currentIdxOfSplit := 0

	totalAmount := uint64(0)

	totalSplitAmount := uint64(0)

	//add change address as payment
	payments = append(
		payments,
		&Payment{
			Address: changeAddress,
			Amount:  0,
		},
	)

	//0 in outputs is for toAddress 1 is for changeAddress
	splitOutputs := make([]*externalapi.DomainTransactionOutput, len(payments))

	//[1] represents the tested unsigned transaction, hence [0] is the last build that didn't violate mass, that can be added as a split
	dummyTransactionWindow := make([]*serialization.PartiallySignedTransaction, 2)

	// This is the actual work, loop through all Inputs and extract the UTXO
	// Create a new "dummy unsigned transaction" in each loop
	// When the dummy unsigned transaction violates transactionmassstandard, commit the former dummyunsignedTransaction as a splitTransaction.
	// For the sake of Effciency the unsignedTransaction is created on-the-fly, within the loop, and commited and reset at split points
	// It might not be the most effecient compounding, but I think it is the most reliable, and exhaustive, and easily malleable to adapt to other needs
	// TO DO: 	1) Map UTXOs according to value to output amount, in a way to use the least amount of available UTXOs per transaction
	//		2) Sign transaction on the fly, for more accurate mass size calcs, unsure if txmass can, or does account for this or not.
	// 		3) deal with multiple payment addresses
	//		4) perhaps incorperate fee payments here, since they are also bound to mass.
	for i, input := range partiallySignedTransaction.Tx.Inputs {

		if currentIdxOfSplit == 0 {
			dummyTransactionWindow[0] = newDummyTransaction(partiallySignedTransaction)
			dummyTransactionWindow[1] = newDummyTransaction(partiallySignedTransaction)

			if massCalculater.CalculateTransactionMass(dummyTransactionWindow[1].Tx)+massleeway >= mempool.MaximumStandardTransactionMass {
				return nil, errors.Errorf("transaction with no inputs or outputs is violating transactionmass")
			}
		}

		currentIdxUTXO := &UTXO{
			Outpoint: &input.PreviousOutpoint,
			UTXOEntry: utxo.NewUTXOEntry(
				partiallySignedTransaction.PartiallySignedInputs[i].PrevOutput.Value-uint64(feePerInput),
				partiallySignedTransaction.PartiallySignedInputs[i].PrevOutput.ScriptPublicKey,
				false,
				constants.UnacceptedDAAScore,
			),
		}
		currentIdxAmount := currentIdxUTXO.UTXOEntry.Amount()
		totalSplitAmount = totalSplitAmount + currentIdxAmount
		totalAmount = totalAmount + currentIdxAmount

		// Below probably isn't needed and can be dealt with when dealing with the last utxo,
		// or early breakage of the loop, if totalAmount exceeds the payment amount
		// None-the-less I think it might be useful for handling multiple payments in the future, and I already wrote it.
		if (totalAmount < payments[0].Amount) && ((totalAmount + currentIdxAmount) > payments[0].Amount) {
			ScriptPublicKey, err := txscript.PayToAddrScript(payments[0].Address)
			if err != nil {
				return nil, err
			}
			splitOutputs[0] = &externalapi.DomainTransactionOutput{
				Value:           (payments[0].Amount - totalAmount),
				ScriptPublicKey: ScriptPublicKey,
			}
			ScriptPublicKey, err = txscript.PayToAddrScript(payments[1].Address)
			if err != nil {
				return nil, err
			}
			splitOutputs[1] = &externalapi.DomainTransactionOutput{
				Value:           currentIdxAmount - (payments[0].Amount - totalAmount),
				ScriptPublicKey: ScriptPublicKey,
			}

			commitOutputToUnsignedTransactionForSchnorrPrivateKey(dummyTransactionWindow[1], splitOutputs[0])
			commitOutputToUnsignedTransactionForSchnorrPrivateKey(dummyTransactionWindow[1], splitOutputs[1])

		} else if totalAmount <= payments[0].Amount {
			ScriptPublicKey, err := txscript.PayToAddrScript(payments[0].Address)
			if err != nil {
				return nil, err
			}
			splitOutputs[0] = &externalapi.DomainTransactionOutput{
				Value:           totalSplitAmount,
				ScriptPublicKey: ScriptPublicKey,
			}

			//we assume output[1] is empty, since we pay back, at the end.
			commitOutputToUnsignedTransactionForSchnorrPrivateKey(dummyTransactionWindow[1], splitOutputs[0])

		} else if totalAmount >= payments[0].Amount {
			ScriptPublicKey, err := txscript.PayToAddrScript(payments[0].Address)
			if err != nil {
				return nil, err
			}
			//we should ideally never enter this scope
			//Only reason is bad selection of UTXOs.

			splitOutputs[1] = &externalapi.DomainTransactionOutput{
				Value:           totalSplitAmount,
				ScriptPublicKey: ScriptPublicKey,
			}
			commitOutputToUnsignedTransactionForSchnorrPrivateKey(dummyTransactionWindow[1], splitOutputs[1])
		}

		//add utxo to unsigned transaction
		commitUTXOToUnsignedTransactionForScnorrPrivateKey(dummyTransactionWindow[1], currentIdxUTXO)

		//commit former dummyTransaction to splitTransactions, if future violates
		if massCalculater.CalculateTransactionMass(dummyTransactionWindow[1].Tx)+massleeway >= mempool.MaximumStandardTransactionMass {
			splitTransactions = append(splitTransactions, dummyTransactionWindow[0])
			currentIdxOfSplit = 0
			totalSplitAmount = 0
			totalAmount = totalAmount + currentIdxAmount
			continue
		}

		//Special case, end of inputs, with no violation, where we can assign dummyWindow[1] to split and break
		if i == len(partiallySignedTransaction.Tx.Inputs)-1 {
			splitTransactions = append(splitTransactions, dummyTransactionWindow[1])
			totalAmount = totalAmount + currentIdxAmount
			break

		}
		totalAmount = totalAmount + currentIdxAmount
		currentIdxOfSplit++

		dummyTransactionWindow[0] = dummyTransactionWindow[1].Clone()
		dummyTransactionWindow[1].Tx.Outputs = nil

	}
	return splitTransactions, nil
}

func createUnsignedTransaction(
	extendedPublicKeys []string,
	minimumSignatures uint32,
	payments []*Payment,
	selectedUTXOs []*UTXO) (*serialization.PartiallySignedTransaction, error) {

	inputs := make([]*externalapi.DomainTransactionInput, len(selectedUTXOs))
	partiallySignedInputs := make([]*serialization.PartiallySignedInput, len(selectedUTXOs))
	for i, utxo := range selectedUTXOs {
		emptyPubKeySignaturePairs := make([]*serialization.PubKeySignaturePair, len(extendedPublicKeys))
		for i, extendedPublicKey := range extendedPublicKeys {
			extendedKey, err := bip32.DeserializeExtendedKey(extendedPublicKey)
			if err != nil {
				return nil, err
			}

			derivedKey, err := extendedKey.DeriveFromPath(utxo.DerivationPath)
			if err != nil {
				return nil, err
			}

			emptyPubKeySignaturePairs[i] = &serialization.PubKeySignaturePair{
				ExtendedPublicKey: derivedKey.String(),
			}
		}

		inputs[i] = &externalapi.DomainTransactionInput{PreviousOutpoint: *utxo.Outpoint}
		partiallySignedInputs[i] = &serialization.PartiallySignedInput{
			PrevOutput: &externalapi.DomainTransactionOutput{
				Value:           utxo.UTXOEntry.Amount(),
				ScriptPublicKey: utxo.UTXOEntry.ScriptPublicKey(),
			},
			MinimumSignatures:    minimumSignatures,
			PubKeySignaturePairs: emptyPubKeySignaturePairs,
			DerivationPath:       utxo.DerivationPath,
		}
	}

	outputs := make([]*externalapi.DomainTransactionOutput, len(payments))
	for i, payment := range payments {
		scriptPublicKey, err := txscript.PayToAddrScript(payment.Address)
		if err != nil {
			return nil, err
		}

		outputs[i] = &externalapi.DomainTransactionOutput{
			Value:           payment.Amount,
			ScriptPublicKey: scriptPublicKey,
		}
	}

	domainTransaction := &externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Gas:          0,
		Payload:      nil,
	}

	return &serialization.PartiallySignedTransaction{
		Tx:                    domainTransaction,
		PartiallySignedInputs: partiallySignedInputs,
	}, nil

}

// IsTransactionFullySigned returns whether the transaction is fully signed and ready to broadcast.
func IsTransactionFullySigned(partiallySignedTransactionBytes []byte) (bool, error) {
	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(partiallySignedTransactionBytes)
	if err != nil {
		return false, err
	}

	return isTransactionFullySigned(partiallySignedTransaction), nil
}

func isTransactionFullySigned(partiallySignedTransaction *serialization.PartiallySignedTransaction) bool {
	for _, input := range partiallySignedTransaction.PartiallySignedInputs {
		numSignatures := 0
		for _, pair := range input.PubKeySignaturePairs {
			if pair.Signature != nil {
				numSignatures++
			}
		}
		if uint32(numSignatures) < input.MinimumSignatures {
			return false
		}
	}
	return true
}

//Extracts a serialized domain transaction from serialized partially signed transaction
func SerializedTransactionFromSerializedPartiallySigned(partiallySignedTransactionBytes []byte, ecda bool) ([]byte, error) {
	deserializedDomainTransaction, err := DeserializedTransactionFromSerializedPartiallySigned(partiallySignedTransactionBytes, ecda)
	if err != nil {
		return nil, err
	}
	return serialization.SerializeDomainTransaction(deserializedDomainTransaction)
}

//Extracts a serialized domain transaction from deserialized partially signed transaction
func SerializedTransactionFromDeserializedPartiallySigned(partiallySignedTransaction *serialization.PartiallySignedTransaction, ecda bool) ([]byte, error) {
	deserializedDomainTransaction, err := DeserializedTransactionFromDeserializedPartiallySigned(partiallySignedTransaction, ecda)
	if err != nil {
		return nil, err
	}
	return serialization.SerializeDomainTransaction(deserializedDomainTransaction)
}

//Extracts a deserialized domain transaction from serialized partially signed transaction
func DeserializedTransactionFromSerializedPartiallySigned(partiallySignedTransactionBytes []byte, ecdsa bool) (*externalapi.DomainTransaction, error) {
	partiallySignedTransaction, err := serialization.DeserializePartiallySignedTransaction(partiallySignedTransactionBytes)
	if err != nil {
		return nil, err
	}
	return DeserializedTransactionFromDeserializedPartiallySigned(partiallySignedTransaction, ecdsa)
}

//Extracts a deserialized domain transaction from deserialized partially signed transaction
func DeserializedTransactionFromDeserializedPartiallySigned(partiallySignedTransaction *serialization.PartiallySignedTransaction, ecdsa bool) (
	*externalapi.DomainTransaction, error) {

	for i, input := range partiallySignedTransaction.PartiallySignedInputs {
		isMultisig := len(input.PubKeySignaturePairs) > 1
		scriptBuilder := txscript.NewScriptBuilder()
		if isMultisig {
			signatureCount := 0
			for _, pair := range input.PubKeySignaturePairs {
				if pair.Signature != nil {
					scriptBuilder.AddData(pair.Signature)
					signatureCount++
				}
			}
			if uint32(signatureCount) < input.MinimumSignatures {
				return nil, errors.Errorf("missing %d signatures", input.MinimumSignatures-uint32(signatureCount))
			}

			redeemScript, err := partiallySignedInputMultisigRedeemScript(input, ecdsa)
			if err != nil {
				return nil, err
			}

			scriptBuilder.AddData(redeemScript)
			sigScript, err := scriptBuilder.Script()
			if err != nil {
				return nil, err
			}

			partiallySignedTransaction.Tx.Inputs[i].SignatureScript = sigScript
		} else {
			if len(input.PubKeySignaturePairs) > 1 {
				return nil, errors.Errorf("Cannot sign on P2PK when len(input.PubKeySignaturePairs) > 1")
			}

			if input.PubKeySignaturePairs[0].Signature == nil {
				return nil, errors.Errorf("missing signature")
			}

			//This function and the related functions are, I think, unusable for sweep, because I do not know what to define
			//as the PubKeySignaturePair equivelent for schnorr. Currently just using partiallySignedTransaction.Tx as a work around.

			sigScript, err := txscript.NewScriptBuilder(). // this is the cause of the error
									AddData(input.PubKeySignaturePairs[0].Signature).
									Script()
			if err != nil {
				return nil, err
			}
			partiallySignedTransaction.Tx.Inputs[i].SignatureScript = sigScript
		}
	}
	return partiallySignedTransaction.Tx, nil
}

func partiallySignedInputMultisigRedeemScript(input *serialization.PartiallySignedInput, ecdsa bool) ([]byte, error) {
	extendedPublicKeys := make([]string, len(input.PubKeySignaturePairs))
	for i, pair := range input.PubKeySignaturePairs {
		extendedPublicKeys[i] = pair.ExtendedPublicKey
	}

	return multiSigRedeemScript(extendedPublicKeys, input.MinimumSignatures, "m", ecdsa)
}
