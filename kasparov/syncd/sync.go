package main

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/kasparov/database"
	"github.com/kaspanet/kaspad/kasparov/dbmodels"
	"github.com/kaspanet/kaspad/kasparov/jsonrpc"
	"github.com/kaspanet/kaspad/kasparov/syncd/config"
	"github.com/kaspanet/kaspad/kasparov/syncd/mqtt"
	"strconv"
	"strings"
	"time"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/btcjson"
	"github.com/kaspanet/kaspad/httpserverutils"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// pendingChainChangedMsgs holds chainChangedMsgs in order of arrival
var pendingChainChangedMsgs []*jsonrpc.ChainChangedMsg

// startSync keeps the node and the database in sync. On start, it downloads
// all data that's missing from the dabase, and once it's done it keeps
// sync with the node via notifications.
func startSync(doneChan chan struct{}) error {
	client, err := jsonrpc.GetClient()
	if err != nil {
		return err
	}

	// Mass download missing data
	err = fetchInitialData(client)
	if err != nil {
		return err
	}

	// Keep the node and the database in sync
	return sync(client, doneChan)
}

// fetchInitialData downloads all data that's currently missing from
// the database.
func fetchInitialData(client *jsonrpc.Client) error {
	log.Infof("Syncing past blocks")
	err := syncBlocks(client)
	if err != nil {
		return err
	}
	log.Infof("Syncing past selected parent chain")
	err = syncSelectedParentChain(client)
	if err != nil {
		return err
	}
	log.Infof("Finished syncing past data")
	return nil
}

// sync keeps the database in sync with the node via notifications
func sync(client *jsonrpc.Client, doneChan chan struct{}) error {
	// Handle client notifications until we're told to stop
	for {
		select {
		case blockAdded := <-client.OnBlockAdded:
			err := handleBlockAddedMsg(client, blockAdded)
			if err != nil {
				return err
			}
		case chainChanged := <-client.OnChainChanged:
			enqueueChainChangedMsg(chainChanged)
			err := processChainChangedMsgs()
			if err != nil {
				return err
			}
		case <-doneChan:
			log.Infof("startSync stopped")
			return nil
		}
	}
}

func stringPointerToString(str *string) string {
	if str == nil {
		return "<nil>"
	}
	return *str
}

// syncBlocks attempts to download all DAG blocks starting with
// the bluest block, and then inserts them into the database.
func syncBlocks(client *jsonrpc.Client) error {
	// Start syncing from the bluest block hash. We use blue score to
	// simulate the "last" block we have because blue-block order is
	// the order that the node uses in the various JSONRPC calls.
	startHash, err := findHashOfBluestBlock(false)
	if err != nil {
		return err
	}

	var rawBlocks []string
	var verboseBlocks []btcjson.GetBlockVerboseResult
	for {
		log.Debugf("Calling getBlocks with start hash %v", stringPointerToString(startHash))
		blocksResult, err := client.GetBlocks(true, true, startHash)
		if err != nil {
			return err
		}
		if len(blocksResult.Hashes) == 0 {
			break
		}

		startHash = &blocksResult.Hashes[len(blocksResult.Hashes)-1]
		rawBlocks = append(rawBlocks, blocksResult.RawBlocks...)
		verboseBlocks = append(verboseBlocks, blocksResult.VerboseBlocks...)
	}

	return addBlocks(client, rawBlocks, verboseBlocks)
}

// syncSelectedParentChain attempts to download the selected parent
// chain starting with the bluest chain-block, and then updates the
// database accordingly.
func syncSelectedParentChain(client *jsonrpc.Client) error {
	// Start syncing from the selected tip hash
	startHash, err := findHashOfBluestBlock(true)
	if err != nil {
		return err
	}

	for {
		log.Debugf("Calling getChainFromBlock with start hash %s", stringPointerToString(startHash))
		chainFromBlockResult, err := client.GetChainFromBlock(false, startHash)
		if err != nil {
			return err
		}
		if len(chainFromBlockResult.AddedChainBlocks) == 0 {
			break
		}

		startHash = &chainFromBlockResult.AddedChainBlocks[len(chainFromBlockResult.AddedChainBlocks)-1].Hash
		err = updateSelectedParentChain(chainFromBlockResult.RemovedChainBlockHashes,
			chainFromBlockResult.AddedChainBlocks)
		if err != nil {
			return err
		}
	}
	return nil
}

// findHashOfBluestBlock finds the block with the highest
// blue score in the database. If the database is empty,
// return nil.
func findHashOfBluestBlock(mustBeChainBlock bool) (*string, error) {
	db, err := database.DB()
	if err != nil {
		return nil, err
	}

	var blockHashes []string
	dbQuery := db.Model(&dbmodels.Block{}).
		Order("blue_score DESC").
		Limit(1)
	if mustBeChainBlock {
		dbQuery = dbQuery.Where(&dbmodels.Block{IsChainBlock: true})
	}
	dbResult := dbQuery.Pluck("block_hash", &blockHashes)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("failed to find hash of bluest block: ", dbErrors)
	}
	if len(blockHashes) == 0 {
		return nil, nil
	}
	return &blockHashes[0], nil
}

// fetchBlock downloads the serialized block and raw block data of
// the block with hash blockHash.
func fetchBlock(client *jsonrpc.Client, blockHash *daghash.Hash) (
	*rawAndVerboseBlock, error) {
	log.Debugf("Getting block %s from the RPC server", blockHash)
	msgBlock, err := client.GetBlock(blockHash, nil)
	if err != nil {
		return nil, err
	}
	writer := bytes.NewBuffer(make([]byte, 0, msgBlock.SerializeSize()))
	err = msgBlock.Serialize(writer)
	if err != nil {
		return nil, err
	}
	rawBlock := hex.EncodeToString(writer.Bytes())

	verboseBlock, err := client.GetBlockVerboseTx(blockHash, nil)
	if err != nil {
		return nil, err
	}
	return &rawAndVerboseBlock{
		rawBlock:     rawBlock,
		verboseBlock: verboseBlock,
	}, nil
}

// addBlocks inserts data in the given rawBlocks and verboseBlocks pairwise
// into the database. See addBlock for further details.
func addBlocks(client *jsonrpc.Client, rawBlocks []string, verboseBlocks []btcjson.GetBlockVerboseResult) error {
	for i, rawBlock := range rawBlocks {
		err := addBlockAndMissingAncestors(client, &rawAndVerboseBlock{
			rawBlock:     rawBlock,
			verboseBlock: &verboseBlocks[i],
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func doesBlockExist(dbTx *gorm.DB, blockHash string) (bool, error) {
	var dbBlock dbmodels.Block
	dbResult := dbTx.
		Where(&dbmodels.Block{BlockHash: blockHash}).
		First(&dbBlock)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return false, httpserverutils.NewErrorFromDBErrors("failed to find block: ", dbErrors)
	}
	return !httpserverutils.IsDBRecordNotFoundError(dbErrors), nil
}

// addBlocks inserts all the data that could be gleaned out of the verbose
// block and raw block data into the database. This includes transactions,
// subnetworks, and addresses.
// Note that if this function may take a nil dbTx, in which case it would start
// a database transaction by itself and commit it before returning.
func addBlock(client *jsonrpc.Client, rawBlock string, verboseBlock btcjson.GetBlockVerboseResult) error {
	db, err := database.DB()
	if err != nil {
		return err
	}
	dbTx := db.Begin()
	defer dbTx.RollbackUnlessCommitted()

	// Skip this block if it already exists.
	blockExists, err := doesBlockExist(dbTx, verboseBlock.Hash)
	if err != nil {
		return err
	}
	if blockExists {
		dbTx.Commit()
		return nil
	}

	dbBlock, err := insertBlock(dbTx, verboseBlock)
	if err != nil {
		return err
	}
	err = insertBlockParents(dbTx, verboseBlock, dbBlock)
	if err != nil {
		return err
	}
	err = insertRawBlockData(dbTx, rawBlock, dbBlock)
	if err != nil {
		return err
	}

	blockMass := uint64(0)
	for i, transaction := range verboseBlock.RawTx {
		dbSubnetwork, err := insertSubnetwork(dbTx, &transaction, client)
		if err != nil {
			return err
		}
		dbTransaction, err := insertTransaction(dbTx, &transaction, dbSubnetwork)
		if err != nil {
			return err
		}
		blockMass += dbTransaction.Mass
		err = insertTransactionBlock(dbTx, dbBlock, dbTransaction, uint32(i))
		if err != nil {
			return err
		}
		err = insertTransactionInputs(dbTx, &transaction, dbTransaction)
		if err != nil {
			return err
		}
		err = insertTransactionOutputs(dbTx, &transaction, dbTransaction)
		if err != nil {
			return err
		}
	}

	dbBlock.Mass = blockMass
	dbResult := dbTx.Save(dbBlock)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to update block: ", dbErrors)
	}

	err = mqtt.PublishTransactionsNotifications(verboseBlock.RawTx)
	if err != nil {
		return err
	}

	dbTx.Commit()
	return nil
}

func insertBlock(dbTx *gorm.DB, verboseBlock btcjson.GetBlockVerboseResult) (*dbmodels.Block, error) {
	bits, err := strconv.ParseUint(verboseBlock.Bits, 16, 32)
	if err != nil {
		return nil, err
	}
	dbBlock := dbmodels.Block{
		BlockHash:            verboseBlock.Hash,
		Version:              verboseBlock.Version,
		HashMerkleRoot:       verboseBlock.HashMerkleRoot,
		AcceptedIDMerkleRoot: verboseBlock.AcceptedIDMerkleRoot,
		UTXOCommitment:       verboseBlock.UTXOCommitment,
		Timestamp:            time.Unix(verboseBlock.Time, 0),
		Bits:                 uint32(bits),
		Nonce:                verboseBlock.Nonce,
		BlueScore:            verboseBlock.BlueScore,
		IsChainBlock:         false, // This must be false for updateSelectedParentChain to work properly
	}

	// Set genesis block as the initial chain block
	if len(verboseBlock.ParentHashes) == 0 {
		dbBlock.IsChainBlock = true
	}
	dbResult := dbTx.Create(&dbBlock)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("failed to insert block: ", dbErrors)
	}
	return &dbBlock, nil
}

func insertBlockParents(dbTx *gorm.DB, verboseBlock btcjson.GetBlockVerboseResult, dbBlock *dbmodels.Block) error {
	// Exit early if this is the genesis block
	if len(verboseBlock.ParentHashes) == 0 {
		return nil
	}

	hashesIn := make([]string, len(verboseBlock.ParentHashes))
	for i, parentHash := range verboseBlock.ParentHashes {
		hashesIn[i] = parentHash
	}
	var dbParents []dbmodels.Block
	dbResult := dbTx.
		Where("block_hash in (?)", hashesIn).
		Find(&dbParents)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to find blocks: ", dbErrors)
	}
	if len(dbParents) != len(verboseBlock.ParentHashes) {
		missingParents := make([]string, 0, len(verboseBlock.ParentHashes)-len(dbParents))
	outerLoop:
		for _, parentHash := range verboseBlock.ParentHashes {
			for _, dbParent := range dbParents {
				if dbParent.BlockHash == parentHash {
					continue outerLoop
				}
			}
			missingParents = append(missingParents, parentHash)
		}
		return errors.Errorf("some parents are missing for block %s: %s", verboseBlock.Hash, strings.Join(missingParents, ", "))
	}

	for _, dbParent := range dbParents {
		dbParentBlock := dbmodels.ParentBlock{
			BlockID:       dbBlock.ID,
			ParentBlockID: dbParent.ID,
		}
		dbResult := dbTx.Create(&dbParentBlock)
		dbErrors := dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return httpserverutils.NewErrorFromDBErrors("failed to insert parentBlock: ", dbErrors)
		}
	}
	return nil
}

func insertRawBlockData(dbTx *gorm.DB, rawBlock string, dbBlock *dbmodels.Block) error {
	blockData, err := hex.DecodeString(rawBlock)
	if err != nil {
		return err
	}
	dbRawBlock := dbmodels.RawBlock{
		BlockID:   dbBlock.ID,
		BlockData: blockData,
	}
	dbResult := dbTx.Create(&dbRawBlock)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to insert rawBlock: ", dbErrors)
	}
	return nil
}

func insertSubnetwork(dbTx *gorm.DB, transaction *btcjson.TxRawResult, client *jsonrpc.Client) (*dbmodels.Subnetwork, error) {
	var dbSubnetwork dbmodels.Subnetwork
	dbResult := dbTx.
		Where(&dbmodels.Subnetwork{SubnetworkID: transaction.Subnetwork}).
		First(&dbSubnetwork)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("failed to find subnetwork: ", dbErrors)
	}
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		subnetwork, err := client.GetSubnetwork(transaction.Subnetwork)
		if err != nil {
			return nil, err
		}
		dbSubnetwork = dbmodels.Subnetwork{
			SubnetworkID: transaction.Subnetwork,
			GasLimit:     subnetwork.GasLimit,
		}
		dbResult := dbTx.Create(&dbSubnetwork)
		dbErrors := dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return nil, httpserverutils.NewErrorFromDBErrors("failed to insert subnetwork: ", dbErrors)
		}
	}
	return &dbSubnetwork, nil
}

func insertTransaction(dbTx *gorm.DB, transaction *btcjson.TxRawResult, dbSubnetwork *dbmodels.Subnetwork) (*dbmodels.Transaction, error) {
	var dbTransaction dbmodels.Transaction
	dbResult := dbTx.
		Where(&dbmodels.Transaction{TransactionID: transaction.TxID}).
		First(&dbTransaction)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("failed to find transaction: ", dbErrors)
	}
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		mass, err := calcTxMass(dbTx, transaction)
		if err != nil {
			return nil, err
		}
		payload, err := hex.DecodeString(transaction.Payload)
		if err != nil {
			return nil, err
		}
		dbTransaction = dbmodels.Transaction{
			TransactionHash: transaction.Hash,
			TransactionID:   transaction.TxID,
			LockTime:        transaction.LockTime,
			SubnetworkID:    dbSubnetwork.ID,
			Gas:             transaction.Gas,
			PayloadHash:     transaction.PayloadHash,
			Payload:         payload,
			Mass:            mass,
		}
		dbResult := dbTx.Create(&dbTransaction)
		dbErrors := dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return nil, httpserverutils.NewErrorFromDBErrors("failed to insert transaction: ", dbErrors)
		}
	}
	return &dbTransaction, nil
}

func calcTxMass(dbTx *gorm.DB, transaction *btcjson.TxRawResult) (uint64, error) {
	msgTx, err := convertTxRawResultToMsgTx(transaction)
	if err != nil {
		return 0, err
	}
	prevTxIDs := make([]string, len(transaction.Vin))
	for i, txIn := range transaction.Vin {
		prevTxIDs[i] = txIn.TxID
	}
	var prevDBTransactionsOutputs []dbmodels.TransactionOutput
	dbResult := dbTx.
		Joins("LEFT JOIN `transactions` ON `transactions`.`id` = `transaction_outputs`.`transaction_id`").
		Where("transactions.transaction_id in (?)", prevTxIDs).
		Preload("Transaction").
		Find(&prevDBTransactionsOutputs)
	dbErrors := dbResult.GetErrors()
	if len(dbErrors) > 0 {
		return 0, httpserverutils.NewErrorFromDBErrors("error fetching previous transactions: ", dbErrors)
	}
	prevScriptPubKeysMap := make(map[string]map[uint32][]byte)
	for _, prevDBTransactionsOutput := range prevDBTransactionsOutputs {
		txID := prevDBTransactionsOutput.Transaction.TransactionID
		if prevScriptPubKeysMap[txID] == nil {
			prevScriptPubKeysMap[txID] = make(map[uint32][]byte)
		}
		prevScriptPubKeysMap[txID][prevDBTransactionsOutput.Index] = prevDBTransactionsOutput.ScriptPubKey
	}
	orderedPrevScriptPubKeys := make([][]byte, len(transaction.Vin))
	for i, txIn := range transaction.Vin {
		orderedPrevScriptPubKeys[i] = prevScriptPubKeysMap[txIn.TxID][uint32(i)]
	}
	return blockdag.CalcTxMass(util.NewTx(msgTx), orderedPrevScriptPubKeys), nil
}

func convertTxRawResultToMsgTx(tx *btcjson.TxRawResult) (*wire.MsgTx, error) {
	txIns := make([]*wire.TxIn, len(tx.Vin))
	for i, txIn := range tx.Vin {
		prevTxID, err := daghash.NewTxIDFromStr(txIn.TxID)
		if err != nil {
			return nil, err
		}
		signatureScript, err := hex.DecodeString(txIn.ScriptSig.Hex)
		if err != nil {
			return nil, err
		}
		txIns[i] = &wire.TxIn{
			PreviousOutpoint: wire.Outpoint{
				TxID:  *prevTxID,
				Index: txIn.Vout,
			},
			SignatureScript: signatureScript,
			Sequence:        txIn.Sequence,
		}
	}
	txOuts := make([]*wire.TxOut, len(tx.Vout))
	for i, txOut := range tx.Vout {
		scriptPubKey, err := hex.DecodeString(txOut.ScriptPubKey.Hex)
		if err != nil {
			return nil, err
		}
		txOuts[i] = &wire.TxOut{
			Value:        txOut.Value,
			ScriptPubKey: scriptPubKey,
		}
	}
	subnetworkID, err := subnetworkid.NewFromStr(tx.Subnetwork)
	if err != nil {
		return nil, err
	}
	if subnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) {
		return wire.NewNativeMsgTx(tx.Version, txIns, txOuts), nil
	}
	payload, err := hex.DecodeString(tx.Payload)
	if err != nil {
		return nil, err
	}
	return wire.NewSubnetworkMsgTx(tx.Version, txIns, txOuts, subnetworkID, tx.Gas, payload), nil
}

func insertTransactionBlock(dbTx *gorm.DB, dbBlock *dbmodels.Block, dbTransaction *dbmodels.Transaction, index uint32) error {
	var dbTransactionBlock dbmodels.TransactionBlock
	dbResult := dbTx.
		Where(&dbmodels.TransactionBlock{TransactionID: dbTransaction.ID, BlockID: dbBlock.ID}).
		First(&dbTransactionBlock)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to find transactionBlock: ", dbErrors)
	}
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		dbTransactionBlock = dbmodels.TransactionBlock{
			TransactionID: dbTransaction.ID,
			BlockID:       dbBlock.ID,
			Index:         index,
		}
		dbResult := dbTx.Create(&dbTransactionBlock)
		dbErrors := dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return httpserverutils.NewErrorFromDBErrors("failed to insert transactionBlock: ", dbErrors)
		}
	}
	return nil
}

func insertTransactionInputs(dbTx *gorm.DB, transaction *btcjson.TxRawResult, dbTransaction *dbmodels.Transaction) error {
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

func isTransactionCoinbase(transaction *btcjson.TxRawResult) (bool, error) {
	subnetwork, err := subnetworkid.NewFromStr(transaction.Subnetwork)
	if err != nil {
		return false, err
	}
	return subnetwork.IsEqual(subnetworkid.SubnetworkIDCoinbase), nil
}

func insertTransactionInput(dbTx *gorm.DB, dbTransaction *dbmodels.Transaction, input *btcjson.Vin) error {
	var dbPreviousTransactionOutput dbmodels.TransactionOutput
	dbResult := dbTx.
		Joins("LEFT JOIN `transactions` ON `transactions`.`id` = `transaction_outputs`.`transaction_id`").
		Where("`transactions`.`transaction_id` = ? AND `transaction_outputs`.`index` = ?", input.TxID, input.Vout).
		First(&dbPreviousTransactionOutput)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to find previous transactionOutput: ", dbErrors)
	}
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		return errors.Errorf("missing output transaction output for txID: %s and index: %d", input.TxID, input.Vout)
	}

	var dbTransactionInputCount int
	dbResult = dbTx.
		Model(&dbmodels.TransactionInput{}).
		Where(&dbmodels.TransactionInput{TransactionID: dbTransaction.ID, PreviousTransactionOutputID: dbPreviousTransactionOutput.ID}).
		Count(&dbTransactionInputCount)
	dbErrors = dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to find transactionInput: ", dbErrors)
	}
	if dbTransactionInputCount == 0 {
		scriptSig, err := hex.DecodeString(input.ScriptSig.Hex)
		if err != nil {
			return nil
		}
		dbTransactionInput := dbmodels.TransactionInput{
			TransactionID:               dbTransaction.ID,
			PreviousTransactionOutputID: dbPreviousTransactionOutput.ID,
			Index:                       input.Vout,
			SignatureScript:             scriptSig,
			Sequence:                    input.Sequence,
		}
		dbResult := dbTx.Create(&dbTransactionInput)
		dbErrors := dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return httpserverutils.NewErrorFromDBErrors("failed to insert transactionInput: ", dbErrors)
		}
	}

	return nil
}

func insertTransactionOutputs(dbTx *gorm.DB, transaction *btcjson.TxRawResult, dbTransaction *dbmodels.Transaction) error {
	for _, output := range transaction.Vout {
		scriptPubKey, err := hex.DecodeString(output.ScriptPubKey.Hex)
		if err != nil {
			return err
		}
		dbAddress, err := insertAddress(dbTx, scriptPubKey)
		if err != nil {
			return err
		}
		err = insertTransactionOutput(dbTx, dbTransaction, &output, scriptPubKey, dbAddress)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertAddress(dbTx *gorm.DB, scriptPubKey []byte) (*dbmodels.Address, error) {
	_, addr, err := txscript.ExtractScriptPubKeyAddress(scriptPubKey, config.ActiveConfig().NetParams())
	if err != nil {
		return nil, err
	}
	hexAddress := addr.EncodeAddress()

	var dbAddress dbmodels.Address
	dbResult := dbTx.
		Where(&dbmodels.Address{Address: hexAddress}).
		First(&dbAddress)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("failed to find address: ", dbErrors)
	}
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		dbAddress = dbmodels.Address{
			Address: hexAddress,
		}
		dbResult := dbTx.Create(&dbAddress)
		dbErrors := dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return nil, httpserverutils.NewErrorFromDBErrors("failed to insert address: ", dbErrors)
		}
	}
	return &dbAddress, nil
}

func insertTransactionOutput(dbTx *gorm.DB, dbTransaction *dbmodels.Transaction,
	output *btcjson.Vout, scriptPubKey []byte, dbAddress *dbmodels.Address) error {
	var dbTransactionOutputCount int
	dbResult := dbTx.
		Model(&dbmodels.TransactionOutput{}).
		Where(&dbmodels.TransactionOutput{TransactionID: dbTransaction.ID, Index: output.N}).
		Count(&dbTransactionOutputCount)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to find transactionOutput: ", dbErrors)
	}
	if dbTransactionOutputCount == 0 {
		dbTransactionOutput := dbmodels.TransactionOutput{
			TransactionID: dbTransaction.ID,
			Index:         output.N,
			Value:         output.Value,
			IsSpent:       false, // This must be false for updateSelectedParentChain to work properly
			ScriptPubKey:  scriptPubKey,
			AddressID:     dbAddress.ID,
		}
		dbResult := dbTx.Create(&dbTransactionOutput)
		dbErrors := dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return httpserverutils.NewErrorFromDBErrors("failed to insert transactionOutput: ", dbErrors)
		}
	}
	return nil
}

// updateSelectedParentChain updates the database to reflect the current selected
// parent chain. First it "unaccepts" all removedChainHashes and then it "accepts"
// all addChainBlocks.
// Note that if this function may take a nil dbTx, in which case it would start
// a database transaction by itself and commit it before returning.
func updateSelectedParentChain(removedChainHashes []string, addedChainBlocks []btcjson.ChainBlock) error {
	db, err := database.DB()
	if err != nil {
		return err
	}
	dbTx := db.Begin()
	defer dbTx.RollbackUnlessCommitted()

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

	dbTx.Commit()
	return nil
}

// updateRemovedChainHashes "unaccepts" the block of the given removedHash.
// That is to say, it marks it as not in the selected parent chain in the
// following ways:
// * All its TransactionInputs.PreviousTransactionOutputs are set IsSpent = false
// * All its Transactions are set AcceptingBlockID = nil
// * The block is set IsChainBlock = false
// This function will return an error if any of the above are in an unexpected state
func updateRemovedChainHashes(dbTx *gorm.DB, removedHash string) error {
	var dbBlock dbmodels.Block
	dbResult := dbTx.
		Where(&dbmodels.Block{BlockHash: removedHash}).
		First(&dbBlock)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to find block: ", dbErrors)
	}
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		return errors.Errorf("missing block for hash: %s", removedHash)
	}
	if !dbBlock.IsChainBlock {
		return errors.Errorf("block erroneously marked as not a chain block: %s", removedHash)
	}

	var dbTransactions []dbmodels.Transaction
	dbResult = dbTx.
		Where(&dbmodels.Transaction{AcceptingBlockID: &dbBlock.ID}).
		Preload("TransactionInputs.PreviousTransactionOutput").
		Find(&dbTransactions)
	dbErrors = dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to find transactions: ", dbErrors)
	}
	for _, dbTransaction := range dbTransactions {
		for _, dbTransactionInput := range dbTransaction.TransactionInputs {
			dbPreviousTransactionOutput := dbTransactionInput.PreviousTransactionOutput
			if !dbPreviousTransactionOutput.IsSpent {
				return errors.Errorf("cannot de-spend an unspent transaction output: %s index: %d",
					dbTransaction.TransactionID, dbTransactionInput.Index)
			}

			dbPreviousTransactionOutput.IsSpent = false
			dbResult = dbTx.Save(&dbPreviousTransactionOutput)
			dbErrors = dbResult.GetErrors()
			if httpserverutils.HasDBError(dbErrors) {
				return httpserverutils.NewErrorFromDBErrors("failed to update transactionOutput: ", dbErrors)
			}
		}

		dbTransaction.AcceptingBlockID = nil
		dbResult := dbTx.Save(&dbTransaction)
		dbErrors := dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return httpserverutils.NewErrorFromDBErrors("failed to update transaction: ", dbErrors)
		}
	}

	dbResult = dbTx.
		Model(&dbmodels.Block{}).
		Where(&dbmodels.Block{AcceptingBlockID: btcjson.Uint64(dbBlock.ID)}).
		Updates(map[string]interface{}{"AcceptingBlockID": nil})

	dbErrors = dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to update blocks: ", dbErrors)
	}

	dbBlock.IsChainBlock = false
	dbResult = dbTx.Save(&dbBlock)
	dbErrors = dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to update block: ", dbErrors)
	}

	return nil
}

// updateAddedChainBlocks "accepts" the given addedBlock. That is to say,
// it marks it as in the selected parent chain in the following ways:
// * All its TransactionInputs.PreviousTransactionOutputs are set IsSpent = true
// * All its Transactions are set AcceptingBlockID = addedBlock
// * The block is set IsChainBlock = true
// This function will return an error if any of the above are in an unexpected state
func updateAddedChainBlocks(dbTx *gorm.DB, addedBlock *btcjson.ChainBlock) error {
	var dbAddedBlock dbmodels.Block
	dbResult := dbTx.
		Where(&dbmodels.Block{BlockHash: addedBlock.Hash}).
		First(&dbAddedBlock)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to find block: ", dbErrors)
	}
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
		return errors.Errorf("missing block for hash: %s", addedBlock.Hash)
	}
	if dbAddedBlock.IsChainBlock {
		return errors.Errorf("block erroneously marked as a chain block: %s", addedBlock.Hash)
	}

	for _, acceptedBlock := range addedBlock.AcceptedBlocks {
		var dbAccepedBlock dbmodels.Block
		dbResult := dbTx.
			Where(&dbmodels.Block{BlockHash: acceptedBlock.Hash}).
			First(&dbAccepedBlock)
		dbErrors := dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return httpserverutils.NewErrorFromDBErrors("failed to find block: ", dbErrors)
		}
		if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
			return errors.Errorf("missing block for hash: %s", acceptedBlock.Hash)
		}
		if dbAccepedBlock.AcceptingBlockID != nil && *dbAccepedBlock.AcceptingBlockID == dbAddedBlock.ID {
			return errors.Errorf("block %s erroneously marked as accepted by %s", acceptedBlock.Hash, addedBlock.Hash)
		}

		transactionIDsIn := make([]string, len(acceptedBlock.AcceptedTxIDs))
		for i, acceptedTxID := range acceptedBlock.AcceptedTxIDs {
			transactionIDsIn[i] = acceptedTxID
		}
		var dbAcceptedTransactions []dbmodels.Transaction
		dbResult = dbTx.
			Where("transaction_id in (?)", transactionIDsIn).
			Preload("TransactionInputs.PreviousTransactionOutput").
			Find(&dbAcceptedTransactions)
		dbErrors = dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return httpserverutils.NewErrorFromDBErrors("failed to find transactions: ", dbErrors)
		}
		if len(dbAcceptedTransactions) != len(acceptedBlock.AcceptedTxIDs) {
			return errors.Errorf("some transaction are missing for block: %s", acceptedBlock.Hash)
		}

		for _, dbAcceptedTransaction := range dbAcceptedTransactions {
			for _, dbTransactionInput := range dbAcceptedTransaction.TransactionInputs {
				dbPreviousTransactionOutput := dbTransactionInput.PreviousTransactionOutput
				if dbPreviousTransactionOutput.IsSpent {
					return errors.Errorf("cannot spend an already spent transaction output: %s index: %d",
						dbAcceptedTransaction.TransactionID, dbTransactionInput.Index)
				}

				dbPreviousTransactionOutput.IsSpent = true
				dbResult = dbTx.Save(&dbPreviousTransactionOutput)
				dbErrors = dbResult.GetErrors()
				if httpserverutils.HasDBError(dbErrors) {
					return httpserverutils.NewErrorFromDBErrors("failed to update transactionOutput: ", dbErrors)
				}
			}

			dbAcceptedTransaction.AcceptingBlockID = &dbAccepedBlock.ID
			dbResult = dbTx.Save(&dbAcceptedTransaction)
			dbErrors = dbResult.GetErrors()
			if httpserverutils.HasDBError(dbErrors) {
				return httpserverutils.NewErrorFromDBErrors("failed to update transaction: ", dbErrors)
			}
		}

		dbAccepedBlock.AcceptingBlockID = btcjson.Uint64(dbAddedBlock.ID)
		dbResult = dbTx.Save(&dbAccepedBlock)
		dbErrors = dbResult.GetErrors()
		if httpserverutils.HasDBError(dbErrors) {
			return httpserverutils.NewErrorFromDBErrors("failed to update block: ", dbErrors)
		}
	}

	dbAddedBlock.IsChainBlock = true
	dbResult = dbTx.Save(&dbAddedBlock)
	dbErrors = dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("failed to update block: ", dbErrors)
	}

	return nil
}

type rawAndVerboseBlock struct {
	rawBlock     string
	verboseBlock *btcjson.GetBlockVerboseResult
}

func (r *rawAndVerboseBlock) String() string {
	return r.verboseBlock.Hash
}

func handleBlockAddedMsg(client *jsonrpc.Client, blockAdded *jsonrpc.BlockAddedMsg) error {
	block, err := fetchBlock(client, blockAdded.Header.BlockHash())
	if err != nil {
		return err
	}
	return addBlockAndMissingAncestors(client, block)
}

func addBlockAndMissingAncestors(client *jsonrpc.Client, block *rawAndVerboseBlock) error {
	blocks, err := fetchBlockAndMissingAncestors(client, block)
	if err != nil {
		return err
	}
	for _, block := range blocks {
		err = addBlock(client, block.rawBlock, *block.verboseBlock)
		if err != nil {
			return err
		}
		log.Infof("Added block %s", block.verboseBlock.Hash)
	}
	return nil
}

func fetchBlockAndMissingAncestors(client *jsonrpc.Client, block *rawAndVerboseBlock) ([]*rawAndVerboseBlock, error) {
	pendingBlocks := []*rawAndVerboseBlock{block}
	blocksToAdd := make([]*rawAndVerboseBlock, 0)
	blocksToAddSet := make(map[string]struct{})
	for len(pendingBlocks) > 0 {
		var currentBlock *rawAndVerboseBlock
		currentBlock, pendingBlocks = pendingBlocks[0], pendingBlocks[1:]
		missingHashes, err := missingParentHashes(currentBlock.verboseBlock.ParentHashes)
		if err != nil {
			return nil, err
		}
		blocksToPrependToPending := make([]*rawAndVerboseBlock, 0, len(missingHashes))
		for _, missingHash := range missingHashes {
			if _, ok := blocksToAddSet[missingHash]; ok {
				continue
			}
			hash, err := daghash.NewHashFromStr(missingHash)
			if err != nil {
				return nil, err
			}
			block, err := fetchBlock(client, hash)
			if err != nil {
				return nil, err
			}
			blocksToPrependToPending = append(blocksToPrependToPending, block)
		}
		if len(blocksToPrependToPending) == 0 {
			blocksToAddSet[currentBlock.verboseBlock.Hash] = struct{}{}
			blocksToAdd = append(blocksToAdd, currentBlock)
			continue
		}
		log.Debugf("Found %s missing parents for block %s and fetched them", blocksToPrependToPending, currentBlock)
		blocksToPrependToPending = append(blocksToPrependToPending, currentBlock)
		pendingBlocks = append(blocksToPrependToPending, pendingBlocks...)
	}
	return blocksToAdd, nil
}

func missingParentHashes(parentHashes []string) ([]string, error) {
	db, err := database.DB()
	if err != nil {
		return nil, err
	}

	// Make sure that all the parent hashes exist in the database
	var dbParentBlocks []dbmodels.Block
	dbResult := db.
		Model(&dbmodels.Block{}).
		Where("block_hash in (?)", parentHashes).
		Find(&dbParentBlocks)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("failed to find parent blocks: ", dbErrors)
	}
	if len(parentHashes) != len(dbParentBlocks) {
		// Some parent hashes are missing. Collect and return them
		var missingHashes []string
	outerLoop:
		for _, hash := range parentHashes {
			for _, dbParentBlock := range dbParentBlocks {
				if dbParentBlock.BlockHash == hash {
					continue outerLoop
				}
			}
			missingHashes = append(missingHashes, hash)
		}
		return missingHashes, nil
	}

	return nil, nil
}

// enqueueChainChangedMsg enqueues onChainChanged messages to be handled later
func enqueueChainChangedMsg(chainChanged *jsonrpc.ChainChangedMsg) {
	pendingChainChangedMsgs = append(pendingChainChangedMsgs, chainChanged)
}

// processChainChangedMsgs processes all pending onChainChanged messages.
// Messages that cannot yet be processed are re-enqueued.
func processChainChangedMsgs() error {
	var unprocessedChainChangedMessages []*jsonrpc.ChainChangedMsg
	for _, chainChanged := range pendingChainChangedMsgs {
		canHandle, err := canHandleChainChangedMsg(chainChanged)
		if err != nil {
			return errors.Wrap(err, "Could not resolve if can handle ChainChangedMsg")
		}
		if !canHandle {
			unprocessedChainChangedMessages = append(unprocessedChainChangedMessages, chainChanged)
			continue
		}

		err = mqtt.PublishUnacceptedTransactionsNotifications(chainChanged.RemovedChainBlockHashes)
		if err != nil {
			panic(errors.Errorf("Error while publishing unaccepted transactions notifications %s", err))
		}

		err = handleChainChangedMsg(chainChanged)
		if err != nil {
			return err
		}
	}
	pendingChainChangedMsgs = unprocessedChainChangedMessages
	return nil
}

func handleChainChangedMsg(chainChanged *jsonrpc.ChainChangedMsg) error {
	// Convert the data in chainChanged to something we can feed into
	// updateSelectedParentChain
	removedHashes, addedBlocks := convertChainChangedMsg(chainChanged)

	err := updateSelectedParentChain(removedHashes, addedBlocks)
	if err != nil {
		return errors.Wrap(err, "Could not update selected parent chain")
	}
	log.Infof("Chain changed: removed %d blocks and added %d block",
		len(removedHashes), len(addedBlocks))

	err = mqtt.PublishAcceptedTransactionsNotifications(chainChanged.AddedChainBlocks)
	if err != nil {
		return errors.Wrap(err, "Error while publishing accepted transactions notifications")
	}
	return mqtt.PublishSelectedTipNotification(addedBlocks[len(addedBlocks)-1].Hash)
}

// canHandleChainChangedMsg checks whether we have all the necessary data
// to successfully handle a ChainChangedMsg.
func canHandleChainChangedMsg(chainChanged *jsonrpc.ChainChangedMsg) (bool, error) {
	db, err := database.DB()
	if err != nil {
		return false, err
	}

	// Collect all referenced block hashes
	hashesIn := make([]string, 0, len(chainChanged.AddedChainBlocks)+len(chainChanged.RemovedChainBlockHashes))
	for _, hash := range chainChanged.RemovedChainBlockHashes {
		hashesIn = append(hashesIn, hash.String())
	}
	for _, block := range chainChanged.AddedChainBlocks {
		hashesIn = append(hashesIn, block.Hash.String())
	}

	// Make sure that all the hashes exist in the database
	var dbBlocks []dbmodels.Block
	dbResult := db.
		Model(&dbmodels.Block{}).
		Where("block_hash in (?)", hashesIn).
		Find(&dbBlocks)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return false, httpserverutils.NewErrorFromDBErrors("failed to find blocks: ", dbErrors)
	}
	if len(hashesIn) != len(dbBlocks) {
		return false, nil
	}

	// Make sure that chain changes are valid for this message
	hashesToIsChainBlocks := make(map[string]bool)
	for _, dbBlock := range dbBlocks {
		hashesToIsChainBlocks[dbBlock.BlockHash] = dbBlock.IsChainBlock
	}
	for _, hash := range chainChanged.RemovedChainBlockHashes {
		isDBBlockChainBlock := hashesToIsChainBlocks[hash.String()]
		if !isDBBlockChainBlock {
			return false, nil
		}
		hashesToIsChainBlocks[hash.String()] = false
	}
	for _, block := range chainChanged.AddedChainBlocks {
		isDBBlockChainBlock := hashesToIsChainBlocks[block.Hash.String()]
		if isDBBlockChainBlock {
			return false, nil
		}
		hashesToIsChainBlocks[block.Hash.String()] = true
	}

	return true, nil
}

func convertChainChangedMsg(chainChanged *jsonrpc.ChainChangedMsg) (
	removedHashes []string, addedBlocks []btcjson.ChainBlock) {

	removedHashes = make([]string, len(chainChanged.RemovedChainBlockHashes))
	for i, hash := range chainChanged.RemovedChainBlockHashes {
		removedHashes[i] = hash.String()
	}

	addedBlocks = make([]btcjson.ChainBlock, len(chainChanged.AddedChainBlocks))
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

	return removedHashes, addedBlocks
}
