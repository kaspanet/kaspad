package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/apiserver/config"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/jsonrpc"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

func blockLoop(doneChan chan struct{}) error {
	client, err := jsonrpc.GetClient()
	if err != nil {
		return err
	}

	err = fetchInitialData(client)
	if err != nil {
		return err
	}

loop:
	for {
		select {
		case blockAdded := <-client.OnBlockAdded:
			handleBlockAddedMsg(client, blockAdded)
		case chainChanged := <-client.OnChainChanged:
			handleChainChangedMsg(chainChanged)
		case <-doneChan:
			log.Infof("blockLoop stopped")
			break loop
		}
	}
	return nil
}

func fetchInitialData(client *jsonrpc.Client) error {
	db, err := database.DB()
	if err != nil {
		return err
	}
	dbTx := db.Begin()

	mostRecentBlockHash := findMostRecentBlockHash(dbTx)
	err = SyncBlocks(client, dbTx, mostRecentBlockHash)
	if err != nil {
		return err
	}
	err = SyncSelectedParentChain(client, dbTx, mostRecentBlockHash)
	if err != nil {
		return err
	}

	dbTx.Commit()
	return nil
}

func findMostRecentBlockHash(dbTx *gorm.DB) *string {
	var block models.Block
	dbTx.Order("blue_score DESC").First(&block)

	if block.ID == 0 {
		return nil
	}
	return &block.BlockHash
}

func SyncBlocks(client *jsonrpc.Client, dbTx *gorm.DB, startHash *string) error {
	var blocks []string
	var rawBlocks []btcjson.GetBlockVerboseResult
	for {
		BlocksResult, err := client.GetBlocks(true, false, startHash)
		if err != nil {
			return err
		}
		if len(BlocksResult.Hashes) == 0 {
			break
		}

		RawBlocksResult, err := client.GetBlocks(true, true, startHash)
		if err != nil {
			return err
		}

		startHash = &BlocksResult.Hashes[len(BlocksResult.Hashes)-1]
		blocks = append(blocks, BlocksResult.Blocks...)
		rawBlocks = append(rawBlocks, RawBlocksResult.RawBlocks...)
	}

	return addBlocks(client, dbTx, blocks, rawBlocks)
}

func SyncSelectedParentChain(client *jsonrpc.Client, dbTx *gorm.DB, startHash *string) error {
	for {
		chainFromBlockResult, err := client.GetChainFromBlock(false, startHash)
		if err != nil {
			return err
		}
		if len(chainFromBlockResult.AddedChainBlocks) == 0 {
			break
		}

		startHash = &chainFromBlockResult.AddedChainBlocks[len(chainFromBlockResult.AddedChainBlocks)].Hash
		err = updateSelectedParentChain(dbTx,
			chainFromBlockResult.RemovedChainBlockHashes, chainFromBlockResult.AddedChainBlocks)
		if err != nil {
			return err
		}
	}
	return nil
}

func fetchBlock(client *jsonrpc.Client, blockHash *daghash.Hash) (
	block string, rawBlock *btcjson.GetBlockVerboseResult, err error) {
	msgBlock, err := client.GetBlock(blockHash, nil)
	if err != nil {
		return "", nil, err
	}
	writer := bytes.NewBuffer(make([]byte, 0, msgBlock.SerializeSize()))
	err = msgBlock.Serialize(writer)
	if err != nil {
		return "", nil, err
	}
	block = hex.EncodeToString(writer.Bytes())

	rawBlock, err = client.GetBlockVerboseTx(blockHash, nil)
	if err != nil {
		return "", nil, err
	}
	return block, rawBlock, nil
}

func addBlocks(client *jsonrpc.Client, dbTx *gorm.DB, blocks []string, rawBlocks []btcjson.GetBlockVerboseResult) error {
	for i, rawBlock := range rawBlocks {
		block := blocks[i]
		err := addBlock(client, dbTx, block, rawBlock)
		if err != nil {
			return err
		}
	}
	return nil
}

func addBlock(client *jsonrpc.Client, dbTx *gorm.DB, block string, rawBlock btcjson.GetBlockVerboseResult) error {
	shouldCommit := false
	if dbTx == nil {
		db, err := database.DB()
		if err != nil {
			return err
		}
		dbTx = db.Begin()
		shouldCommit = true
	}

	dbBlock, err := insertBlock(dbTx, rawBlock)
	if err != nil {
		return err
	}
	err = insertBlockParents(dbTx, rawBlock, dbBlock)
	if err != nil {
		return err
	}
	err = insertBlockData(dbTx, block, dbBlock)
	if err != nil {
		return err
	}

	for i, transaction := range rawBlock.RawTx {
		dbSubnetwork, err := insertSubnetwork(dbTx, &transaction, client)
		if err != nil {
			return err
		}
		dbTransaction, err := insertTransaction(dbTx, &transaction, dbSubnetwork)
		if err != nil {
			return err
		}
		insertTransactionBlock(dbTx, dbBlock, dbTransaction, uint32(i))
		err = insertTransactionInputs(dbTx, &transaction, dbTransaction)
		if err != nil {
			return err
		}
		err = insertTransactionOutputs(dbTx, &transaction, dbTransaction)
		if err != nil {
			return err
		}
	}

	if shouldCommit {
		dbTx.Commit()
	}

	return nil
}

func insertBlock(dbTx *gorm.DB, rawBlock btcjson.GetBlockVerboseResult) (*models.Block, error) {
	var dbBlock *models.Block
	dbTx.Where(&models.Block{BlockHash: rawBlock.Hash}).First(&dbBlock)
	if dbBlock.ID == 0 {
		bits, err := strconv.ParseUint(rawBlock.Bits, 16, 32)
		if err != nil {
			return nil, err
		}
		dbBlock = &models.Block{
			BlockHash:            rawBlock.Hash,
			Version:              rawBlock.Version,
			HashMerkleRoot:       rawBlock.HashMerkleRoot,
			AcceptedIDMerkleRoot: rawBlock.AcceptedIDMerkleRoot,
			UTXOCommitment:       rawBlock.UTXOCommitment,
			Timestamp:            time.Unix(rawBlock.Time, 0),
			Bits:                 uint32(bits),
			Nonce:                rawBlock.Nonce,
			BlueScore:            rawBlock.BlueScore,
			IsChainBlock:         false, // This must be false for updateSelectedParentChain to work properly
		}
		dbTx.Create(dbBlock)
	}
	return dbBlock, nil
}

func insertBlockParents(dbTx *gorm.DB, rawBlock btcjson.GetBlockVerboseResult, dbBlock *models.Block) error {
	for _, parentHash := range rawBlock.ParentHashes {
		var dbParent models.Block
		dbTx.Where(&models.Block{BlockHash: parentHash}).First(&dbParent)
		if dbParent.ID == 0 {
			return fmt.Errorf("missing parent for hash: %s", parentHash)
		}

		var dbParentBlock models.ParentBlock
		dbTx.Where(&models.ParentBlock{BlockID: dbBlock.ID, ParentBlockID: dbParent.ID}).First(&dbParentBlock)
		if dbParentBlock.BlockID == 0 {
			dbParentBlock = models.ParentBlock{
				BlockID:       dbBlock.ID,
				ParentBlockID: dbParent.ID,
			}
			dbTx.Create(&dbParentBlock)
		}
	}
	return nil
}

func insertBlockData(dbTx *gorm.DB, block string, dbBlock *models.Block) error {
	var dbRawBlock models.RawBlock
	dbTx.Where(&models.RawBlock{BlockID: dbBlock.ID}).First(&dbRawBlock)
	if dbRawBlock.BlockID == 0 {
		blockData, err := hex.DecodeString(block)
		if err != nil {
			return err
		}
		dbRawBlock = models.RawBlock{
			BlockID:   dbBlock.ID,
			BlockData: blockData,
		}
		dbTx.Create(&dbRawBlock)
	}
	return nil
}

func insertSubnetwork(dbTx *gorm.DB, transaction *btcjson.TxRawResult, client *jsonrpc.Client) (*models.Subnetwork, error) {
	var dbSubnetwork *models.Subnetwork
	dbTx.Where(&models.Subnetwork{SubnetworkID: transaction.Subnetwork}).First(&dbSubnetwork)
	if dbSubnetwork.ID == 0 {
		subnetwork, err := client.GetSubnetwork(transaction.Subnetwork)
		if err != nil {
			return nil, err
		}
		dbSubnetwork = &models.Subnetwork{
			SubnetworkID: transaction.Subnetwork,
			GasLimit:     subnetwork.GasLimit,
		}
		dbTx.Create(&dbSubnetwork)
	}
	return dbSubnetwork, nil
}

func insertTransaction(dbTx *gorm.DB, transaction *btcjson.TxRawResult, dbSubnetwork *models.Subnetwork) (*models.Transaction, error) {
	var dbTransaction *models.Transaction
	dbTx.Where(&models.Transaction{TransactionID: transaction.TxID}).First(&dbTransaction)
	if dbTransaction.ID == 0 {
		payload, err := hex.DecodeString(transaction.Payload)
		if err != nil {
			return nil, err
		}
		dbTransaction = &models.Transaction{
			TransactionHash: transaction.Hash,
			TransactionID:   transaction.TxID,
			LockTime:        transaction.LockTime,
			SubnetworkID:    dbSubnetwork.ID,
			Gas:             transaction.Gas,
			PayloadHash:     transaction.PayloadHash,
			Payload:         payload,
		}
		dbTx.Create(&dbTransaction)
	}
	return dbTransaction, nil
}

func insertTransactionBlock(dbTx *gorm.DB, dbBlock *models.Block, dbTransaction *models.Transaction, index uint32) {
	var dbTransactionBlock models.TransactionBlock
	dbTx.Where(&models.TransactionBlock{TransactionID: dbTransaction.ID, BlockID: dbBlock.ID}).First(&dbTransactionBlock)
	if dbTransactionBlock.TransactionID == 0 {
		dbTransactionBlock = models.TransactionBlock{
			TransactionID: dbTransaction.ID,
			BlockID:       dbBlock.ID,
			Index:         index,
		}
		dbTx.Create(&dbTransactionBlock)
	}
}

func insertTransactionInputs(dbTx *gorm.DB, transaction *btcjson.TxRawResult, dbTransaction *models.Transaction) error {
	isCoinbase, err := isTransactionCoinbase(transaction)
	if err != nil {
		return err
	}

	if !isCoinbase {
		for _, input := range transaction.Vin {
			err := insertTransactionInput(dbTx, dbTransaction, &input)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func insertTransactionInput(dbTx *gorm.DB, dbTransaction *models.Transaction, input *btcjson.Vin) error {
	var dbOutputTransaction models.Transaction
	dbTx.Where(&models.Transaction{TransactionID: input.TxID}).First(&dbOutputTransaction)
	if dbOutputTransaction.ID == 0 {
		return fmt.Errorf("missing output transaction for txID: %s", input.TxID)
	}

	var dbOutputTransactionOutput models.TransactionOutput
	dbTx.Where(&models.TransactionOutput{TransactionID: dbOutputTransaction.ID, Index: input.Vout}).First(&dbOutputTransactionOutput)
	if dbOutputTransactionOutput.ID == 0 {
		return fmt.Errorf("missing output transaction output for txID: %s and index: %d", input.TxID, input.Vout)
	}

	var dbTransactionInput models.TransactionInput
	dbTx.Where(models.TransactionInput{TransactionID: dbTransaction.ID, TransactionOutputID: dbOutputTransactionOutput.ID}).First(&dbTransactionInput)
	if dbTransactionInput.TransactionID == 0 {
		scriptSig, err := hex.DecodeString(input.ScriptSig.Hex)
		if err != nil {
			return nil
		}
		dbTransactionInput = models.TransactionInput{
			TransactionID:       dbTransaction.ID,
			TransactionOutputID: dbOutputTransactionOutput.ID,
			Index:               input.Vout,
			SignatureScript:     scriptSig,
			Sequence:            input.Sequence,
		}
		dbTx.Create(&dbTransactionInput)
	}

	return nil
}

func insertTransactionOutputs(dbTx *gorm.DB, transaction *btcjson.TxRawResult, dbTransaction *models.Transaction) error {
	for _, output := range transaction.Vout {
		scriptPubKey, err := hex.DecodeString(output.ScriptPubKey.Hex)
		if err != nil {
			return err
		}
		dbAddress, err := insertAddress(dbTx, scriptPubKey)
		if err != nil {
			return err
		}
		insertTransactionOutput(dbTx, dbTransaction, &output, scriptPubKey, dbAddress)
	}
	return nil
}

func insertAddress(dbTx *gorm.DB, scriptPubKey []byte) (*models.Address, error) {
	_, addrs, _, err := txscript.ExtractScriptPubKeyAddrs(scriptPubKey, &config.ActiveNetParams)
	if err != nil {
		return nil, err
	}
	address := addrs[0].EncodeAddress()

	var dbAddress *models.Address
	dbTx.Where(&models.Address{Address: address}).First(&dbAddress)
	if dbAddress.ID == 0 {
		dbAddress = &models.Address{
			Address: address,
		}
		dbTx.Create(&dbAddress)
	}
	return dbAddress, nil
}

func insertTransactionOutput(dbTx *gorm.DB, dbTransaction *models.Transaction,
	output *btcjson.Vout, scriptPubKey []byte, dbAddress *models.Address) {
	var dbTransactionOutput models.TransactionOutput
	dbTx.Where(&models.TransactionOutput{TransactionID: dbTransaction.ID, Index: output.N}).First(&dbTransactionOutput)
	if dbTransactionOutput.TransactionID == 0 {
		dbTransactionOutput = models.TransactionOutput{
			TransactionID: dbTransaction.ID,
			Index:         output.N,
			Value:         output.Value,
			IsSpent:       false,
			ScriptPubKey:  scriptPubKey,
			AddressID:     dbAddress.ID,
		}
		dbTx.Create(&dbTransactionOutput)
	}
}

func isTransactionCoinbase(transaction *btcjson.TxRawResult) (bool, error) {
	subnetwork, err := subnetworkid.NewFromStr(transaction.Subnetwork)
	if err != nil {
		return false, err
	}
	return subnetwork.IsEqual(subnetworkid.SubnetworkIDCoinbase), nil
}

func updateSelectedParentChain(dbTx *gorm.DB, removedChainHashes []string, addedChainBlocks []btcjson.ChainBlock) error {
	shouldCommit := false
	if dbTx == nil {
		db, err := database.DB()
		if err != nil {
			return err
		}
		dbTx = db.Begin()
		shouldCommit = true
	}

	for _, removedHash := range removedChainHashes {
		err := updateRemovedChainHashes(dbTx, removedHash)
		if err != nil {
			return err
		}
	}
	for _, addedBlock := range addedChainBlocks {
		err := updateAddedChainBlocks(dbTx, &addedBlock)
		if err != nil {
			return err
		}
	}

	if shouldCommit {
		dbTx.Commit()
	}

	return nil
}

func updateRemovedChainHashes(dbTx *gorm.DB, removedHash string) error {
	var dbBlock models.Block
	dbTx.Where(&models.Block{BlockHash: removedHash}).First(&dbBlock)
	if dbBlock.ID == 0 {
		return fmt.Errorf("missing block for hash: %s", removedHash)
	}
	if dbBlock.IsChainBlock == false {
		return fmt.Errorf("block erroneously marked as not a chain block: %s", removedHash)
	}

	var dbTransactions []models.Transaction
	dbTx.Where(&models.Transaction{AcceptingBlockID: &dbBlock.ID}).Preload("TransactionInputs").Find(&dbTransactions)
	for _, dbTransaction := range dbTransactions {
		for _, dbTransactionInput := range dbTransaction.TransactionInputs {
			var dbTransactionOutput models.TransactionOutput
			dbTx.Where(&models.TransactionOutput{ID: dbTransactionInput.TransactionOutputID}).First(&dbTransactionOutput)
			if dbTransactionOutput.ID == 0 {
				return fmt.Errorf("missing transaction output for transaction: %s index: %d",
					dbTransaction.TransactionID, dbTransactionInput.Index)
			}
			if dbTransactionOutput.IsSpent == false {
				return fmt.Errorf("cannot de-spend an unspent transaction output: %s index: %d",
					dbTransaction.TransactionID, dbTransactionInput.Index)
			}

			dbTransactionOutput.IsSpent = false
			dbTx.Save(&dbTransactionOutput)
		}

		dbTransaction.AcceptingBlockID = nil
		dbTx.Save(&dbTransaction)
	}

	dbBlock.IsChainBlock = false
	dbTx.Save(&dbBlock)

	return nil
}

func updateAddedChainBlocks(dbTx *gorm.DB, addedBlock *btcjson.ChainBlock) error {
	for _, acceptedBlock := range addedBlock.AcceptedBlocks {
		var dbAcceptingBlock models.Block
		dbTx.Where(&models.Block{BlockHash: acceptedBlock.Hash}).First(dbAcceptingBlock)
		if dbAcceptingBlock.ID == 0 {
			return fmt.Errorf("missing block for hash: %s", acceptedBlock.Hash)
		}
		if dbAcceptingBlock.IsChainBlock == true {
			return fmt.Errorf("block erroneously marked as a chain block: %s", acceptedBlock.Hash)
		}

		for _, acceptedTxID := range acceptedBlock.AcceptedTxIDs {
			var dbTransaction models.Transaction
			dbTx.Where(&models.Transaction{TransactionID: acceptedTxID}).First(&dbTransaction)
			if dbTransaction.ID == 0 {
				return fmt.Errorf("missing transaction for txID: %s", acceptedTxID)
			}

			var dbTransactionInputs []models.TransactionInput
			dbTx.Where(&models.TransactionInput{TransactionID: dbTransaction.ID}).Preload("TransactionInputs").Find(&dbTransactionInputs)
			for _, dbTransactionInput := range dbTransactionInputs {
				var dbTransactionOutput models.TransactionOutput
				dbTx.Where(&models.TransactionOutput{ID: dbTransactionInput.TransactionOutputID}).First(&dbTransactionOutput)
				if dbTransactionOutput.ID == 0 {
					return fmt.Errorf("missing transaction output for transaction: %s index: %d",
						dbTransaction.TransactionID, dbTransactionInput.Index)
				}
				if dbTransactionOutput.IsSpent == true {
					return fmt.Errorf("cannot spend an already spent transaction output: %s index: %d",
						dbTransaction.TransactionID, dbTransactionInput.Index)
				}

				dbTransactionOutput.IsSpent = true
				dbTx.Save(&dbTransactionOutput)
			}

			dbTransaction.AcceptingBlockID = &dbAcceptingBlock.ID
			dbTx.Save(&dbTransaction)
		}

		dbAcceptingBlock.IsChainBlock = true
		dbTx.Save(&dbAcceptingBlock)
	}
	return nil
}

func handleBlockAddedMsg(client *jsonrpc.Client, blockAdded *jsonrpc.BlockAddedMsg) {
	hash := blockAdded.Header.BlockHash()
	block, rawBlock, err := fetchBlock(client, hash)
	if err != nil {
		log.Warnf("Could not fetch block %s: %s", hash, err)
		return
	}
	err = addBlock(client, nil, block, *rawBlock)
	if err != nil {
		log.Warnf("Could not insert block %s: %s", hash, err)
		return
	}
	log.Infof("Added block %s", hash)
}

func handleChainChangedMsg(chainChanged *jsonrpc.ChainChangedMsg) {
	removedHashes := make([]string, len(chainChanged.RemovedChainBlockHashes))
	for i, hash := range chainChanged.RemovedChainBlockHashes {
		removedHashes[i] = hash.String()
	}
	addedBlocks := make([]btcjson.ChainBlock, len(chainChanged.AddedChainBlocks))
	for i, addedBlock := range chainChanged.AddedChainBlocks {
		acceptedBlocks := make([]btcjson.AcceptedBlock, len(addedBlock.AcceptedBlocks))
		for j, acceptedBlock := range addedBlock.AcceptedBlocks {
			acceptedTxIDs := make([]string, len(acceptedBlock.AcceptedTxIDs))
			for k, acceptedTxID := range acceptedBlock.AcceptedTxIDs {
				acceptedTxIDs[k] = acceptedTxID.String()
			}
			acceptedBlocks[j] = btcjson.AcceptedBlock{
				Hash:          acceptedBlock.Hash.String(),
				AcceptedTxIDs: acceptedTxIDs,
			}
		}
		addedBlocks[i] = btcjson.ChainBlock{
			Hash:           addedBlock.Hash.String(),
			AcceptedBlocks: acceptedBlocks,
		}
	}

	err := updateSelectedParentChain(nil, removedHashes, addedBlocks)
	if err != nil {
		log.Warnf("Could not update selected parent chain: %s", err)
		return
	}
	log.Infof("Chain changed: removed &d blocks and added %d block",
		len(removedHashes), len(addedBlocks))
}
