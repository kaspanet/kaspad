package appmessage

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/util/mstime"
)

// DomainBlockToMsgBlock converts an externalapi.DomainBlock to MsgBlock
func DomainBlockToMsgBlock(domainBlock *externalapi.DomainBlock) *MsgBlock {
	msgTxs := make([]*MsgTx, 0, len(domainBlock.Transactions))
	for _, domainTransaction := range domainBlock.Transactions {
		msgTxs = append(msgTxs, DomainTransactionToMsgTx(domainTransaction))
	}
	return &MsgBlock{
		Header:       *DomainBlockHeaderToBlockHeader(domainBlock.Header),
		Transactions: msgTxs,
	}
}

// DomainBlockHeaderToBlockHeader converts an externalapi.BlockHeader to MsgBlockHeader
func DomainBlockHeaderToBlockHeader(domainBlockHeader externalapi.BlockHeader) *MsgBlockHeader {
	return &MsgBlockHeader{
		Version:              domainBlockHeader.Version(),
		ParentHashes:         domainBlockHeader.ParentHashes(),
		HashMerkleRoot:       domainBlockHeader.HashMerkleRoot(),
		AcceptedIDMerkleRoot: domainBlockHeader.AcceptedIDMerkleRoot(),
		UTXOCommitment:       domainBlockHeader.UTXOCommitment(),
		Timestamp:            mstime.UnixMilliseconds(domainBlockHeader.TimeInMilliseconds()),
		Bits:                 domainBlockHeader.Bits(),
		Nonce:                domainBlockHeader.Nonce(),
	}
}

// MsgBlockToDomainBlock converts a MsgBlock to externalapi.DomainBlock
func MsgBlockToDomainBlock(msgBlock *MsgBlock) *externalapi.DomainBlock {
	transactions := make([]*externalapi.DomainTransaction, 0, len(msgBlock.Transactions))
	for _, msgTx := range msgBlock.Transactions {
		transactions = append(transactions, MsgTxToDomainTransaction(msgTx))
	}

	return &externalapi.DomainBlock{
		Header:       BlockHeaderToDomainBlockHeader(&msgBlock.Header),
		Transactions: transactions,
	}
}

// BlockHeaderToDomainBlockHeader converts a MsgBlockHeader to externalapi.BlockHeader
func BlockHeaderToDomainBlockHeader(blockHeader *MsgBlockHeader) externalapi.BlockHeader {
	return blockheader.NewImmutableBlockHeader(
		blockHeader.Version,
		blockHeader.ParentHashes,
		blockHeader.HashMerkleRoot,
		blockHeader.AcceptedIDMerkleRoot,
		blockHeader.UTXOCommitment,
		blockHeader.Timestamp.UnixMilliseconds(),
		blockHeader.Bits,
		blockHeader.Nonce,
	)
}

// DomainTransactionToMsgTx converts an externalapi.DomainTransaction into an MsgTx
func DomainTransactionToMsgTx(domainTransaction *externalapi.DomainTransaction) *MsgTx {
	txIns := make([]*TxIn, 0, len(domainTransaction.Inputs))
	for _, input := range domainTransaction.Inputs {
		txIns = append(txIns, domainTransactionInputToTxIn(input))
	}

	txOuts := make([]*TxOut, 0, len(domainTransaction.Outputs))
	for _, output := range domainTransaction.Outputs {
		txOuts = append(txOuts, domainTransactionOutputToTxOut(output))
	}

	return &MsgTx{
		Version:      domainTransaction.Version,
		TxIn:         txIns,
		TxOut:        txOuts,
		LockTime:     domainTransaction.LockTime,
		SubnetworkID: domainTransaction.SubnetworkID,
		Gas:          domainTransaction.Gas,
		PayloadHash:  domainTransaction.PayloadHash,
		Payload:      domainTransaction.Payload,
	}
}

func domainTransactionOutputToTxOut(domainTransactionOutput *externalapi.DomainTransactionOutput) *TxOut {
	return &TxOut{
		Value:        domainTransactionOutput.Value,
		ScriptPubKey: domainTransactionOutput.ScriptPublicKey,
	}
}

func domainTransactionInputToTxIn(domainTransactionInput *externalapi.DomainTransactionInput) *TxIn {
	return &TxIn{
		PreviousOutpoint: *domainOutpointToOutpoint(domainTransactionInput.PreviousOutpoint),
		SignatureScript:  domainTransactionInput.SignatureScript,
		Sequence:         domainTransactionInput.Sequence,
	}
}

func domainOutpointToOutpoint(domainOutpoint externalapi.DomainOutpoint) *Outpoint {
	return NewOutpoint(
		&domainOutpoint.TransactionID,
		domainOutpoint.Index)
}

// MsgTxToDomainTransaction converts an MsgTx into externalapi.DomainTransaction
func MsgTxToDomainTransaction(msgTx *MsgTx) *externalapi.DomainTransaction {
	transactionInputs := make([]*externalapi.DomainTransactionInput, 0, len(msgTx.TxIn))
	for _, txIn := range msgTx.TxIn {
		transactionInputs = append(transactionInputs, txInToDomainTransactionInput(txIn))
	}

	transactionOutputs := make([]*externalapi.DomainTransactionOutput, 0, len(msgTx.TxOut))
	for _, txOut := range msgTx.TxOut {
		transactionOutputs = append(transactionOutputs, txOutToDomainTransactionOutput(txOut))
	}

	payload := make([]byte, 0)
	if msgTx.Payload != nil {
		payload = msgTx.Payload
	}

	return &externalapi.DomainTransaction{
		Version:      msgTx.Version,
		Inputs:       transactionInputs,
		Outputs:      transactionOutputs,
		LockTime:     msgTx.LockTime,
		SubnetworkID: msgTx.SubnetworkID,
		Gas:          msgTx.Gas,
		PayloadHash:  msgTx.PayloadHash,
		Payload:      payload,
	}
}

func txOutToDomainTransactionOutput(txOut *TxOut) *externalapi.DomainTransactionOutput {
	return &externalapi.DomainTransactionOutput{
		Value:           txOut.Value,
		ScriptPublicKey: txOut.ScriptPubKey,
	}
}

func txInToDomainTransactionInput(txIn *TxIn) *externalapi.DomainTransactionInput {
	return &externalapi.DomainTransactionInput{
		PreviousOutpoint: *outpointToDomainOutpoint(&txIn.PreviousOutpoint), //TODO
		SignatureScript:  txIn.SignatureScript,
		Sequence:         txIn.Sequence,
	}
}

func outpointToDomainOutpoint(outpoint *Outpoint) *externalapi.DomainOutpoint {
	return &externalapi.DomainOutpoint{
		TransactionID: outpoint.TxID,
		Index:         outpoint.Index,
	}
}

// RPCTransactionToDomainTransaction converts RPCTransactions to DomainTransactions
func RPCTransactionToDomainTransaction(rpcTransaction *RPCTransaction) (*externalapi.DomainTransaction, error) {
	inputs := make([]*externalapi.DomainTransactionInput, len(rpcTransaction.Inputs))
	for i, input := range rpcTransaction.Inputs {
		transactionIDBytes, err := hex.DecodeString(input.PreviousOutpoint.TransactionID)
		if err != nil {
			return nil, err
		}
		transactionID, err := transactionid.FromBytes(transactionIDBytes)
		if err != nil {
			return nil, err
		}
		previousOutpoint := &externalapi.DomainOutpoint{
			TransactionID: *transactionID,
			Index:         input.PreviousOutpoint.Index,
		}
		signatureScript, err := hex.DecodeString(input.SignatureScript)
		if err != nil {
			return nil, err
		}
		inputs[i] = &externalapi.DomainTransactionInput{
			PreviousOutpoint: *previousOutpoint,
			SignatureScript:  signatureScript,
			Sequence:         input.Sequence,
		}
	}
	outputs := make([]*externalapi.DomainTransactionOutput, len(rpcTransaction.Outputs))
	for i, output := range rpcTransaction.Outputs {
		scriptPublicKey, err := hex.DecodeString(output.ScriptPubKey.Script)
		if err != nil {
			return nil, err
		}
		outputs[i] = &externalapi.DomainTransactionOutput{
			Value:           output.Amount,
			ScriptPublicKey: &externalapi.ScriptPublicKey{Script: scriptPublicKey, Version: output.ScriptPubKey.Version},
		}
	}

	subnetworkIDBytes, err := hex.DecodeString(rpcTransaction.SubnetworkID)
	if err != nil {
		return nil, err
	}
	subnetworkID, err := subnetworks.FromBytes(subnetworkIDBytes)
	if err != nil {
		return nil, err
	}
	payloadHashBytes, err := hex.DecodeString(rpcTransaction.PayloadHash)
	if err != nil {
		return nil, err
	}
	payloadHash, err := externalapi.NewDomainHashFromByteSlice(payloadHashBytes)
	if err != nil {
		return nil, err
	}
	payload, err := hex.DecodeString(rpcTransaction.Payload)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainTransaction{
		Version:      rpcTransaction.Version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     rpcTransaction.LockTime,
		SubnetworkID: *subnetworkID,
		Gas:          rpcTransaction.LockTime,
		PayloadHash:  *payloadHash,
		Payload:      payload,
	}, nil
}

// DomainTransactionToRPCTransaction converts DomainTransactions to RPCTransactions
func DomainTransactionToRPCTransaction(transaction *externalapi.DomainTransaction) *RPCTransaction {
	inputs := make([]*RPCTransactionInput, len(transaction.Inputs))
	for i, input := range transaction.Inputs {
		transactionID := input.PreviousOutpoint.TransactionID.String()
		previousOutpoint := &RPCOutpoint{
			TransactionID: transactionID,
			Index:         input.PreviousOutpoint.Index,
		}
		signatureScript := hex.EncodeToString(input.SignatureScript)
		inputs[i] = &RPCTransactionInput{
			PreviousOutpoint: previousOutpoint,
			SignatureScript:  signatureScript,
			Sequence:         input.Sequence,
		}
	}
	outputs := make([]*RPCTransactionOutput, len(transaction.Outputs))
	for i, output := range transaction.Outputs {
		scriptPublicKey := hex.EncodeToString(output.ScriptPublicKey.Script)
		outputs[i] = &RPCTransactionOutput{
			Amount:       output.Value,
			ScriptPubKey: &RPCScriptPublicKey{Script: scriptPublicKey, Version: output.ScriptPublicKey.Version},
		}
	}
	subnetworkID := hex.EncodeToString(transaction.SubnetworkID[:])
	payloadHash := transaction.PayloadHash.String()
	payload := hex.EncodeToString(transaction.Payload)
	return &RPCTransaction{
		Version:      transaction.Version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     transaction.LockTime,
		SubnetworkID: subnetworkID,
		Gas:          transaction.LockTime,
		PayloadHash:  payloadHash,
		Payload:      payload,
	}
}
