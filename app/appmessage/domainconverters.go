package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
)

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

func DomainBlockHeaderToBlockHeader(domainBlockHeader *externalapi.DomainBlockHeader) *BlockHeader {
	return &BlockHeader{
		Version:              domainBlockHeader.Version,
		ParentHashes:         domainBlockHeader.ParentHashes,
		HashMerkleRoot:       &domainBlockHeader.HashMerkleRoot,
		AcceptedIDMerkleRoot: &domainBlockHeader.AcceptedIDMerkleRoot,
		UTXOCommitment:       &domainBlockHeader.UTXOCommitment,
		Timestamp:            mstime.UnixMilliseconds(domainBlockHeader.TimeInMilliseconds),
		Bits:                 domainBlockHeader.Bits,
		Nonce:                domainBlockHeader.Nonce,
	}
}

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

func BlockHeaderToDomainBlockHeader(blockHeader *BlockHeader) *externalapi.DomainBlockHeader {
	return &externalapi.DomainBlockHeader{
		Version:              blockHeader.Version,
		ParentHashes:         blockHeader.ParentHashes,
		HashMerkleRoot:       *blockHeader.HashMerkleRoot,
		AcceptedIDMerkleRoot: *blockHeader.AcceptedIDMerkleRoot,
		UTXOCommitment:       *blockHeader.UTXOCommitment,
		TimeInMilliseconds:   blockHeader.Timestamp.UnixMilliseconds(),
		Bits:                 blockHeader.Bits,
		Nonce:                blockHeader.Nonce,
	}
}

// DomainTransactionToMsgTx converts a DomainTransaction into an appmessage.MsgTx
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
		PayloadHash:  &domainTransaction.PayloadHash,
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

func MsgTxToDomainTransaction(msgTx *MsgTx) *externalapi.DomainTransaction {
	transactionInputs := make([]*externalapi.DomainTransactionInput, 0, len(msgTx.TxIn))
	for _, txIn := range msgTx.TxIn {
		transactionInputs = append(transactionInputs, txInToDomainTransactionInput(txIn))
	}

	transactionOutputs := make([]*externalapi.DomainTransactionOutput, 0, len(msgTx.TxOut))
	for _, txOut := range msgTx.TxOut {
		transactionOutputs = append(transactionOutputs, txOutToDomainTransactionOutput(txOut))
	}
	return &externalapi.DomainTransaction{
		Version:      msgTx.Version,
		Inputs:       transactionInputs,
		Outputs:      transactionOutputs,
		LockTime:     msgTx.LockTime,
		SubnetworkID: msgTx.SubnetworkID,
		Gas:          msgTx.Gas,
		PayloadHash:  *msgTx.PayloadHash,
		Payload:      msgTx.Payload,
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
