package appmessage

import (
	"encoding/hex"
	"github.com/pkg/errors"
	"math/big"

	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

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
		Parents:              domainBlockHeader.Parents(),
		HashMerkleRoot:       domainBlockHeader.HashMerkleRoot(),
		AcceptedIDMerkleRoot: domainBlockHeader.AcceptedIDMerkleRoot(),
		UTXOCommitment:       domainBlockHeader.UTXOCommitment(),
		Timestamp:            mstime.UnixMilliseconds(domainBlockHeader.TimeInMilliseconds()),
		Bits:                 domainBlockHeader.Bits(),
		Nonce:                domainBlockHeader.Nonce(),
		BlueScore:            domainBlockHeader.BlueScore(),
		DAAScore:             domainBlockHeader.DAAScore(),
		BlueWork:             domainBlockHeader.BlueWork(),
		PruningPoint:         domainBlockHeader.PruningPoint(),
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
		blockHeader.Parents,
		blockHeader.HashMerkleRoot,
		blockHeader.AcceptedIDMerkleRoot,
		blockHeader.UTXOCommitment,
		blockHeader.Timestamp.UnixMilliseconds(),
		blockHeader.Bits,
		blockHeader.Nonce,
		blockHeader.DAAScore,
		blockHeader.BlueScore,
		blockHeader.BlueWork,
		blockHeader.PruningPoint,
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
		SigOpCount:       domainTransactionInput.SigOpCount,
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
		SigOpCount:       txIn.SigOpCount,
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
		previousOutpoint, err := RPCOutpointToDomainOutpoint(input.PreviousOutpoint)
		if err != nil {
			return nil, err
		}
		signatureScript, err := hex.DecodeString(input.SignatureScript)
		if err != nil {
			return nil, err
		}
		inputs[i] = &externalapi.DomainTransactionInput{
			PreviousOutpoint: *previousOutpoint,
			SignatureScript:  signatureScript,
			Sequence:         input.Sequence,
			SigOpCount:       input.SigOpCount,
		}
	}
	outputs := make([]*externalapi.DomainTransactionOutput, len(rpcTransaction.Outputs))
	for i, output := range rpcTransaction.Outputs {
		scriptPublicKey, err := hex.DecodeString(output.ScriptPublicKey.Script)
		if err != nil {
			return nil, err
		}
		outputs[i] = &externalapi.DomainTransactionOutput{
			Value:           output.Amount,
			ScriptPublicKey: &externalapi.ScriptPublicKey{Script: scriptPublicKey, Version: output.ScriptPublicKey.Version},
		}
	}

	subnetworkID, err := subnetworks.FromString(rpcTransaction.SubnetworkID)
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
		Payload:      payload,
	}, nil
}

// RPCOutpointToDomainOutpoint converts RPCOutpoint to  DomainOutpoint
func RPCOutpointToDomainOutpoint(outpoint *RPCOutpoint) (*externalapi.DomainOutpoint, error) {
	transactionID, err := transactionid.FromString(outpoint.TransactionID)
	if err != nil {
		return nil, err
	}
	return &externalapi.DomainOutpoint{
		TransactionID: *transactionID,
		Index:         outpoint.Index,
	}, nil
}

// RPCUTXOEntryToUTXOEntry converts RPCUTXOEntry to UTXOEntry
func RPCUTXOEntryToUTXOEntry(entry *RPCUTXOEntry) (externalapi.UTXOEntry, error) {
	script, err := hex.DecodeString(entry.ScriptPublicKey.Script)
	if err != nil {
		return nil, err
	}

	return utxo.NewUTXOEntry(
		entry.Amount,
		&externalapi.ScriptPublicKey{
			Script:  script,
			Version: entry.ScriptPublicKey.Version,
		},
		entry.IsCoinbase,
		entry.BlockDAAScore,
	), nil
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
			SigOpCount:       input.SigOpCount,
		}
	}
	outputs := make([]*RPCTransactionOutput, len(transaction.Outputs))
	for i, output := range transaction.Outputs {
		scriptPublicKey := hex.EncodeToString(output.ScriptPublicKey.Script)
		outputs[i] = &RPCTransactionOutput{
			Amount:          output.Value,
			ScriptPublicKey: &RPCScriptPublicKey{Script: scriptPublicKey, Version: output.ScriptPublicKey.Version},
		}
	}
	subnetworkID := transaction.SubnetworkID.String()
	payload := hex.EncodeToString(transaction.Payload)
	return &RPCTransaction{
		Version:      transaction.Version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     transaction.LockTime,
		SubnetworkID: subnetworkID,
		Gas:          transaction.LockTime,
		Payload:      payload,
	}
}

// OutpointAndUTXOEntryPairsToDomainOutpointAndUTXOEntryPairs converts
// OutpointAndUTXOEntryPairs to domain OutpointAndUTXOEntryPairs
func OutpointAndUTXOEntryPairsToDomainOutpointAndUTXOEntryPairs(
	outpointAndUTXOEntryPairs []*OutpointAndUTXOEntryPair) []*externalapi.OutpointAndUTXOEntryPair {

	domainOutpointAndUTXOEntryPairs := make([]*externalapi.OutpointAndUTXOEntryPair, len(outpointAndUTXOEntryPairs))
	for i, outpointAndUTXOEntryPair := range outpointAndUTXOEntryPairs {
		domainOutpointAndUTXOEntryPairs[i] = outpointAndUTXOEntryPairToDomainOutpointAndUTXOEntryPair(outpointAndUTXOEntryPair)
	}
	return domainOutpointAndUTXOEntryPairs
}

func outpointAndUTXOEntryPairToDomainOutpointAndUTXOEntryPair(
	outpointAndUTXOEntryPair *OutpointAndUTXOEntryPair) *externalapi.OutpointAndUTXOEntryPair {
	return &externalapi.OutpointAndUTXOEntryPair{
		Outpoint: &externalapi.DomainOutpoint{
			TransactionID: outpointAndUTXOEntryPair.Outpoint.TxID,
			Index:         outpointAndUTXOEntryPair.Outpoint.Index,
		},
		UTXOEntry: utxo.NewUTXOEntry(
			outpointAndUTXOEntryPair.UTXOEntry.Amount,
			outpointAndUTXOEntryPair.UTXOEntry.ScriptPublicKey,
			outpointAndUTXOEntryPair.UTXOEntry.IsCoinbase,
			outpointAndUTXOEntryPair.UTXOEntry.BlockDAAScore,
		),
	}
}

// DomainOutpointAndUTXOEntryPairsToOutpointAndUTXOEntryPairs converts
// domain OutpointAndUTXOEntryPairs to OutpointAndUTXOEntryPairs
func DomainOutpointAndUTXOEntryPairsToOutpointAndUTXOEntryPairs(
	outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) []*OutpointAndUTXOEntryPair {

	domainOutpointAndUTXOEntryPairs := make([]*OutpointAndUTXOEntryPair, len(outpointAndUTXOEntryPairs))
	for i, outpointAndUTXOEntryPair := range outpointAndUTXOEntryPairs {
		domainOutpointAndUTXOEntryPairs[i] = &OutpointAndUTXOEntryPair{
			Outpoint: &Outpoint{
				TxID:  outpointAndUTXOEntryPair.Outpoint.TransactionID,
				Index: outpointAndUTXOEntryPair.Outpoint.Index,
			},
			UTXOEntry: &UTXOEntry{
				Amount:          outpointAndUTXOEntryPair.UTXOEntry.Amount(),
				ScriptPublicKey: outpointAndUTXOEntryPair.UTXOEntry.ScriptPublicKey(),
				IsCoinbase:      outpointAndUTXOEntryPair.UTXOEntry.IsCoinbase(),
				BlockDAAScore:   outpointAndUTXOEntryPair.UTXOEntry.BlockDAAScore(),
			},
		}
	}
	return domainOutpointAndUTXOEntryPairs
}

// DomainBlockToRPCBlock converts DomainBlocks to RPCBlocks
func DomainBlockToRPCBlock(block *externalapi.DomainBlock) *RPCBlock {
	parents := make([]*RPCBlockLevelParents, len(block.Header.Parents()))
	for i, blockLevelParents := range block.Header.Parents() {
		parents[i] = &RPCBlockLevelParents{
			ParentHashes: hashes.ToStrings(blockLevelParents),
		}
	}
	header := &RPCBlockHeader{
		Version:              uint32(block.Header.Version()),
		Parents:              parents,
		HashMerkleRoot:       block.Header.HashMerkleRoot().String(),
		AcceptedIDMerkleRoot: block.Header.AcceptedIDMerkleRoot().String(),
		UTXOCommitment:       block.Header.UTXOCommitment().String(),
		Timestamp:            block.Header.TimeInMilliseconds(),
		Bits:                 block.Header.Bits(),
		Nonce:                block.Header.Nonce(),
		DAAScore:             block.Header.DAAScore(),
		BlueScore:            block.Header.BlueScore(),
		BlueWork:             block.Header.BlueWork().Text(16),
		PruningPoint:         block.Header.PruningPoint().String(),
	}
	transactions := make([]*RPCTransaction, len(block.Transactions))
	for i, transaction := range block.Transactions {
		transactions[i] = DomainTransactionToRPCTransaction(transaction)
	}
	return &RPCBlock{
		Header:       header,
		Transactions: transactions,
	}
}

// RPCBlockToDomainBlock converts `block` into a DomainBlock
func RPCBlockToDomainBlock(block *RPCBlock) (*externalapi.DomainBlock, error) {
	parents := make([]externalapi.BlockLevelParents, len(block.Header.Parents))
	for i, blockLevelParents := range block.Header.Parents {
		parents[i] = make(externalapi.BlockLevelParents, len(blockLevelParents.ParentHashes))
		for j, parentHash := range blockLevelParents.ParentHashes {
			var err error
			parents[i][j], err = externalapi.NewDomainHashFromString(parentHash)
			if err != nil {
				return nil, err
			}
		}
	}
	hashMerkleRoot, err := externalapi.NewDomainHashFromString(block.Header.HashMerkleRoot)
	if err != nil {
		return nil, err
	}
	acceptedIDMerkleRoot, err := externalapi.NewDomainHashFromString(block.Header.AcceptedIDMerkleRoot)
	if err != nil {
		return nil, err
	}
	utxoCommitment, err := externalapi.NewDomainHashFromString(block.Header.UTXOCommitment)
	if err != nil {
		return nil, err
	}
	blueWork, success := new(big.Int).SetString(block.Header.BlueWork, 16)
	if !success {
		return nil, errors.Errorf("failed to parse blue work: %s", block.Header.BlueWork)
	}
	pruningPoint, err := externalapi.NewDomainHashFromString(block.Header.PruningPoint)
	if err != nil {
		return nil, err
	}
	header := blockheader.NewImmutableBlockHeader(
		uint16(block.Header.Version),
		parents,
		hashMerkleRoot,
		acceptedIDMerkleRoot,
		utxoCommitment,
		block.Header.Timestamp,
		block.Header.Bits,
		block.Header.Nonce,
		block.Header.DAAScore,
		block.Header.BlueScore,
		blueWork,
		pruningPoint)
	transactions := make([]*externalapi.DomainTransaction, len(block.Transactions))
	for i, transaction := range block.Transactions {
		domainTransaction, err := RPCTransactionToDomainTransaction(transaction)
		if err != nil {
			return nil, err
		}
		transactions[i] = domainTransaction
	}
	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactions,
	}, nil
}

// BlockWithTrustedDataToDomainBlockWithTrustedData converts *MsgBlockWithTrustedData to *externalapi.BlockWithTrustedData
func BlockWithTrustedDataToDomainBlockWithTrustedData(block *MsgBlockWithTrustedData) *externalapi.BlockWithTrustedData {
	daaWindow := make([]*externalapi.TrustedDataDataDAABlock, len(block.DAAWindow))
	for i, daaBlock := range block.DAAWindow {
		daaWindow[i] = &externalapi.TrustedDataDataDAABlock{
			Header:       BlockHeaderToDomainBlockHeader(daaBlock.Header),
			GHOSTDAGData: ghostdagDataToDomainGHOSTDAGData(daaBlock.GHOSTDAGData),
		}
	}

	ghostdagData := make([]*externalapi.BlockGHOSTDAGDataHashPair, len(block.GHOSTDAGData))
	for i, datum := range block.GHOSTDAGData {
		ghostdagData[i] = &externalapi.BlockGHOSTDAGDataHashPair{
			Hash:         datum.Hash,
			GHOSTDAGData: ghostdagDataToDomainGHOSTDAGData(datum.GHOSTDAGData),
		}
	}

	return &externalapi.BlockWithTrustedData{
		Block:        MsgBlockToDomainBlock(block.Block),
		DAAScore:     block.DAAScore,
		DAAWindow:    daaWindow,
		GHOSTDAGData: ghostdagData,
	}
}

func ghostdagDataToDomainGHOSTDAGData(data *BlockGHOSTDAGData) *externalapi.BlockGHOSTDAGData {
	bluesAnticoneSizes := make(map[externalapi.DomainHash]externalapi.KType, len(data.BluesAnticoneSizes))
	for _, pair := range data.BluesAnticoneSizes {
		bluesAnticoneSizes[*pair.BlueHash] = pair.AnticoneSize
	}
	return externalapi.NewBlockGHOSTDAGData(
		data.BlueScore,
		data.BlueWork,
		data.SelectedParent,
		data.MergeSetBlues,
		data.MergeSetReds,
		bluesAnticoneSizes,
	)
}

func domainGHOSTDAGDataGHOSTDAGData(data *externalapi.BlockGHOSTDAGData) *BlockGHOSTDAGData {
	bluesAnticoneSizes := make([]*BluesAnticoneSizes, 0, len(data.BluesAnticoneSizes()))
	for blueHash, anticoneSize := range data.BluesAnticoneSizes() {
		blueHashCopy := blueHash
		bluesAnticoneSizes = append(bluesAnticoneSizes, &BluesAnticoneSizes{
			BlueHash:     &blueHashCopy,
			AnticoneSize: anticoneSize,
		})
	}

	return &BlockGHOSTDAGData{
		BlueScore:          data.BlueScore(),
		BlueWork:           data.BlueWork(),
		SelectedParent:     data.SelectedParent(),
		MergeSetBlues:      data.MergeSetBlues(),
		MergeSetReds:       data.MergeSetReds(),
		BluesAnticoneSizes: bluesAnticoneSizes,
	}
}

// DomainBlockWithTrustedDataToBlockWithTrustedData converts *externalapi.BlockWithTrustedData to *MsgBlockWithTrustedData
func DomainBlockWithTrustedDataToBlockWithTrustedData(block *externalapi.BlockWithTrustedData) *MsgBlockWithTrustedData {
	daaWindow := make([]*TrustedDataDataDAABlock, len(block.DAAWindow))
	for i, daaBlock := range block.DAAWindow {
		daaWindow[i] = &TrustedDataDataDAABlock{
			Header:       DomainBlockHeaderToBlockHeader(daaBlock.Header),
			GHOSTDAGData: domainGHOSTDAGDataGHOSTDAGData(daaBlock.GHOSTDAGData),
		}
	}

	ghostdagData := make([]*BlockGHOSTDAGDataHashPair, len(block.GHOSTDAGData))
	for i, datum := range block.GHOSTDAGData {
		ghostdagData[i] = &BlockGHOSTDAGDataHashPair{
			Hash:         datum.Hash,
			GHOSTDAGData: domainGHOSTDAGDataGHOSTDAGData(datum.GHOSTDAGData),
		}
	}

	return &MsgBlockWithTrustedData{
		Block:        DomainBlockToMsgBlock(block.Block),
		DAAScore:     block.DAAScore,
		DAAWindow:    daaWindow,
		GHOSTDAGData: ghostdagData,
	}
}

// MsgPruningPointProofToDomainPruningPointProof converts *MsgPruningPointProof to *externalapi.PruningPointProof
func MsgPruningPointProofToDomainPruningPointProof(pruningPointProofMessage *MsgPruningPointProof) *externalapi.PruningPointProof {
	headers := make([][]externalapi.BlockHeader, len(pruningPointProofMessage.Headers))
	for blockLevel, blockLevelParents := range pruningPointProofMessage.Headers {
		headers[blockLevel] = make([]externalapi.BlockHeader, len(blockLevelParents))
		for i, header := range blockLevelParents {
			headers[blockLevel][i] = BlockHeaderToDomainBlockHeader(header)
		}
	}
	return &externalapi.PruningPointProof{
		Headers: headers,
	}
}

// DomainPruningPointProofToMsgPruningPointProof converts *externalapi.PruningPointProof to *MsgPruningPointProof
func DomainPruningPointProofToMsgPruningPointProof(pruningPointProof *externalapi.PruningPointProof) *MsgPruningPointProof {
	headers := make([][]*MsgBlockHeader, len(pruningPointProof.Headers))
	for blockLevel, blockLevelParents := range pruningPointProof.Headers {
		headers[blockLevel] = make([]*MsgBlockHeader, len(blockLevelParents))
		for i, header := range blockLevelParents {
			headers[blockLevel][i] = DomainBlockHeaderToBlockHeader(header)
		}
	}
	return &MsgPruningPointProof{
		Headers: headers,
	}
}
