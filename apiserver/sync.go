package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/apiserver/config"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/jsonrpc"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/apiserver/utils"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

// startSync keeps the node and the API server in sync. On start, it downloads
// all data that's missing from the API server, and once it's done it keeps
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

	// Keep the node and the API server in sync
	sync(client, doneChan)
	return nil
}

// fetchInitialData downloads all data that's currently missing from
// the database.
func fetchInitialData(client *jsonrpc.Client) error {
	err := syncBlocks(client)
	if err != nil {
		return err
	}
	err = syncSelectedParentChain(client)
	if err != nil {
		return err
	}
	return nil
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

	var blocks []string
	var rawBlocks []btcjson.GetBlockVerboseResult
	for {
		blocksResult, err := client.GetBlocks(true, false, startHash)
		if err != nil {
			return err
		}
		if len(blocksResult.Hashes) == 0 {
			break
		}

		rawBlocksResult, err := client.GetBlocks(true, true, startHash)
		if err != nil {
			return err
		}

		startHash = &blocksResult.Hashes[len(blocksResult.Hashes)-1]
		blocks = append(blocks, blocksResult.Blocks...)
		rawBlocks = append(rawBlocks, rawBlocksResult.RawBlocks...)
	}

	return addBlocks(client, blocks, rawBlocks)
}

// syncSelectedParentChain attempts to download the selected parent
// chain starting with the bluest chain-block, and then updates the
// database accordingly.
func syncSelectedParentChain(client *jsonrpc.Client) error {
	// Start syncing from the bluest chain-block hash. We use blue
	// score to simulate the "last" block we have because blue-block
	// order is the order that the node uses in the various JSONRPC
	// calls.
	startHash, err := findHashOfBluestBlock(true)
	if err != nil {
		return err
	}

	for {
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
	dbTx, err := database.DB()
	if err != nil {
		return nil, err
	}

	var block models.Block
	dbQuery := dbTx.Order("blue_score DESC")
	if mustBeChainBlock {
		dbQuery.Where(&models.Block{IsChainBlock: true})
	}
	dbResult := dbQuery.First(&block)
	if utils.HasDBError(dbResult) {
		return nil, utils.NewErrorFromDBErrors("failed to find hash of bluest block: ", dbResult.GetErrors())
	}
	if utils.HasDBRecordNotFoundError(dbResult) {
		return nil, nil
	}
	return &block.BlockHash, nil
}

// fetchBlock downloads the serialized block and raw block data of
// the block with hash blockHash.
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

// addBlocks inserts data in the given blocks and rawBlocks pairwise
// into the database. See addBlock for further details.
func addBlocks(client *jsonrpc.Client, blocks []string, rawBlocks []btcjson.GetBlockVerboseResult) error {
	for i, rawBlock := range rawBlocks {
		block := blocks[i]
		err := addBlock(client, block, rawBlock)
		if err != nil {
			return err
		}
	}
	return nil
}

func doesBlockExist(dbTx *gorm.DB, blockHash string) (bool, error) {
	var dbBlock models.Block
	dbResult := dbTx.
		Where(&models.Block{BlockHash: blockHash}).
		First(&dbBlock)
	if utils.HasDBError(dbResult) {
		return false, utils.NewErrorFromDBErrors("failed to find block: ", dbResult.GetErrors())
	}
	return !utils.HasDBRecordNotFoundError(dbResult), nil
}

// addBlocks inserts all the data that could be gleaned out of the serialized
// block and raw block data into the database. This includes transactions,
// subnetworks, and addresses.
// Note that if this function may take a nil dbTx, in which case it would start
// a database transaction by itself and commit it before returning.
func addBlock(client *jsonrpc.Client, block string, rawBlock btcjson.GetBlockVerboseResult) error {
	db, err := database.DB()
	if err != nil {
		return err
	}
	dbTx := db.Begin()

	// Skip this block if it already exists.
	blockExists, err := doesBlockExist(dbTx, rawBlock.Hash)
	if err != nil {
		return err
	}
	if blockExists {
		dbTx.Commit()
		return nil
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

	dbTx.Commit()
	return nil
}

func insertBlock(dbTx *gorm.DB, rawBlock btcjson.GetBlockVerboseResult) (*models.Block, error) {
	bits, err := strconv.ParseUint(rawBlock.Bits, 16, 32)
	if err != nil {
		return nil, err
	}
	dbBlock := models.Block{
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
		Mass:                 rawBlock.Mass,
	}
	dbResult := dbTx.Create(&dbBlock)
	if utils.HasDBError(dbResult) {
		return nil, utils.NewErrorFromDBErrors("failed to insert block: ", dbResult.GetErrors())
	}
	return &dbBlock, nil
}

func insertBlockParents(dbTx *gorm.DB, rawBlock btcjson.GetBlockVerboseResult, dbBlock *models.Block) error {
	// Exit early if this is the genesis block
	if len(rawBlock.ParentHashes) == 0 {
		return nil
	}

	dbWhereBlockIDsIn := make([]*models.Block, len(rawBlock.ParentHashes))
	for i, parentHash := range rawBlock.ParentHashes {
		dbWhereBlockIDsIn[i] = &models.Block{BlockHash: parentHash}
	}
	var dbParents []models.Block
	dbResult := dbTx.
		Where(dbWhereBlockIDsIn).
		First(&dbParents)
	if utils.HasDBError(dbResult) {
		return utils.NewErrorFromDBErrors("failed to find blocks: ", dbResult.GetErrors())
	}
	if len(dbParents) != len(rawBlock.ParentHashes) {
		return fmt.Errorf("some parents are missing for block: %s", rawBlock.Hash)
	}

	for _, dbParent := range dbParents {
		dbParentBlock := models.ParentBlock{
			BlockID:       dbBlock.ID,
			ParentBlockID: dbParent.ID,
		}
		dbResult := dbTx.Create(&dbParentBlock)
		if utils.HasDBError(dbResult) {
			return utils.NewErrorFromDBErrors("failed to insert parentBlock: ", dbResult.GetErrors())
		}
	}
	return nil
}

func insertBlockData(dbTx *gorm.DB, block string, dbBlock *models.Block) error {
	blockData, err := hex.DecodeString(block)
	if err != nil {
		return err
	}
	dbRawBlock := models.RawBlock{
		BlockID:   dbBlock.ID,
		BlockData: blockData,
	}
	dbResult := dbTx.Create(&dbRawBlock)
	if utils.HasDBError(dbResult) {
		return utils.NewErrorFromDBErrors("failed to insert rawBlock: ", dbResult.GetErrors())
	}
	return nil
}

func insertSubnetwork(dbTx *gorm.DB, transaction *btcjson.TxRawResult, client *jsonrpc.Client) (*models.Subnetwork, error) {
	var dbSubnetwork models.Subnetwork
	dbResult := dbTx.
		Where(&models.Subnetwork{SubnetworkID: transaction.Subnetwork}).
		First(&dbSubnetwork)
	if utils.HasDBError(dbResult) {
		return nil, utils.NewErrorFromDBErrors("failed to find subnetwork: ", dbResult.GetErrors())
	}
	if utils.HasDBRecordNotFoundError(dbResult) {
		subnetwork, err := client.GetSubnetwork(transaction.Subnetwork)
		if err != nil {
			return nil, err
		}
		dbSubnetwork = models.Subnetwork{
			SubnetworkID: transaction.Subnetwork,
			GasLimit:     subnetwork.GasLimit,
		}
		dbResult := dbTx.Create(&dbSubnetwork)
		if utils.HasDBError(dbResult) {
			return nil, utils.NewErrorFromDBErrors("failed to insert subnetwork: ", dbResult.GetErrors())
		}
	}
	return &dbSubnetwork, nil
}

func insertTransaction(dbTx *gorm.DB, transaction *btcjson.TxRawResult, dbSubnetwork *models.Subnetwork) (*models.Transaction, error) {
	var dbTransaction models.Transaction
	dbResult := dbTx.
		Where(&models.Transaction{TransactionID: transaction.TxID}).
		First(&dbTransaction)
	if utils.HasDBError(dbResult) {
		return nil, utils.NewErrorFromDBErrors("failed to find transaction: ", dbResult.GetErrors())
	}
	if utils.HasDBRecordNotFoundError(dbResult) {
		payload, err := hex.DecodeString(transaction.Payload)
		if err != nil {
			return nil, err
		}
		dbTransaction = models.Transaction{
			TransactionHash: transaction.Hash,
			TransactionID:   transaction.TxID,
			LockTime:        transaction.LockTime,
			SubnetworkID:    dbSubnetwork.ID,
			Gas:             transaction.Gas,
			Mass:            transaction.Mass,
			PayloadHash:     transaction.PayloadHash,
			Payload:         payload,
		}
		dbResult := dbTx.Create(&dbTransaction)
		if utils.HasDBError(dbResult) {
			return nil, utils.NewErrorFromDBErrors("failed to insert transaction: ", dbResult.GetErrors())
		}
	}
	return &dbTransaction, nil
}

func insertTransactionBlock(dbTx *gorm.DB, dbBlock *models.Block, dbTransaction *models.Transaction, index uint32) error {
	var dbTransactionBlock models.TransactionBlock
	dbResult := dbTx.
		Where(&models.TransactionBlock{TransactionID: dbTransaction.ID, BlockID: dbBlock.ID}).
		First(&dbTransactionBlock)
	if utils.HasDBError(dbResult) {
		return utils.NewErrorFromDBErrors("failed to find transactionBlock: ", dbResult.GetErrors())
	}
	if utils.HasDBRecordNotFoundError(dbResult) {
		dbTransactionBlock = models.TransactionBlock{
			TransactionID: dbTransaction.ID,
			BlockID:       dbBlock.ID,
			Index:         index,
		}
		dbResult := dbTx.Create(&dbTransactionBlock)
		if utils.HasDBError(dbResult) {
			return utils.NewErrorFromDBErrors("failed to insert transactionBlock: ", dbResult.GetErrors())
		}
	}
	return nil
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

func isTransactionCoinbase(transaction *btcjson.TxRawResult) (bool, error) {
	subnetwork, err := subnetworkid.NewFromStr(transaction.Subnetwork)
	if err != nil {
		return false, err
	}
	return subnetwork.IsEqual(subnetworkid.SubnetworkIDCoinbase), nil
}

func insertTransactionInput(dbTx *gorm.DB, dbTransaction *models.Transaction, input *btcjson.Vin) error {
	var dbPreviousTransactionOutput models.TransactionOutput
	dbResult := dbTx.
		Joins("LEFT JOIN `transactions` ON `transactions`.`id` = `transaction_outputs`.`transaction_id`").
		Where("`transactions`.`transactiond_id` = ? AND `transaction_outputs`.`index` = ?", input.TxID, input.Vout).
		First(&dbPreviousTransactionOutput)
	if utils.HasDBError(dbResult) {
		return utils.NewErrorFromDBErrors("failed to find previous transactionOutput: ", dbResult.GetErrors())
	}
	if utils.HasDBRecordNotFoundError(dbResult) {
		return fmt.Errorf("missing output transaction output for txID: %s and index: %d", input.TxID, input.Vout)
	}

	var dbTransactionInput models.TransactionInput
	dbResult = dbTx.
		Where(models.TransactionInput{TransactionID: dbTransaction.ID, PreviousTransactionOutputID: dbPreviousTransactionOutput.ID}).
		First(&dbTransactionInput)
	if utils.HasDBError(dbResult) {
		return utils.NewErrorFromDBErrors("failed to find transactionInput: ", dbResult.GetErrors())
	}
	if utils.HasDBRecordNotFoundError(dbResult) {
		scriptSig, err := hex.DecodeString(input.ScriptSig.Hex)
		if err != nil {
			return nil
		}
		dbTransactionInput = models.TransactionInput{
			TransactionID:               dbTransaction.ID,
			PreviousTransactionOutputID: dbPreviousTransactionOutput.ID,
			Index:                       input.Vout,
			SignatureScript:             scriptSig,
			Sequence:                    input.Sequence,
		}
		dbResult := dbTx.Create(&dbTransactionInput)
		if utils.HasDBError(dbResult) {
			return utils.NewErrorFromDBErrors("failed to insert transactionInput: ", dbResult.GetErrors())
		}
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
		err = insertTransactionOutput(dbTx, dbTransaction, &output, scriptPubKey, dbAddress)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertAddress(dbTx *gorm.DB, scriptPubKey []byte) (*models.Address, error) {
	_, addr, err := txscript.ExtractScriptPubKeyAddress(scriptPubKey, config.ActiveNetParams())
	if err != nil {
		return nil, err
	}
	address := addr.EncodeAddress()

	var dbAddress models.Address
	dbResult := dbTx.
		Where(&models.Address{Address: address}).
		First(&dbAddress)
	if utils.HasDBError(dbResult) {
		return nil, utils.NewErrorFromDBErrors("failed to find address: ", dbResult.GetErrors())
	}
	if utils.HasDBRecordNotFoundError(dbResult) {
		dbAddress = models.Address{
			Address: address,
		}
		dbResult := dbTx.Create(&dbAddress)
		if utils.HasDBError(dbResult) {
			return nil, utils.NewErrorFromDBErrors("failed to insert address: ", dbResult.GetErrors())
		}
	}
	return &dbAddress, nil
}

func insertTransactionOutput(dbTx *gorm.DB, dbTransaction *models.Transaction,
	output *btcjson.Vout, scriptPubKey []byte, dbAddress *models.Address) error {
	var dbTransactionOutput models.TransactionOutput
	dbResult := dbTx.
		Where(&models.TransactionOutput{TransactionID: dbTransaction.ID, Index: output.N}).
		First(&dbTransactionOutput)
	if utils.HasDBError(dbResult) {
		return utils.NewErrorFromDBErrors("failed to find transactionOutput: ", dbResult.GetErrors())
	}
	if utils.HasDBRecordNotFoundError(dbResult) {
		dbTransactionOutput = models.TransactionOutput{
			TransactionID: dbTransaction.ID,
			Index:         output.N,
			Value:         output.Value,
			IsSpent:       false, // This must be false for updateSelectedParentChain to work properly
			ScriptPubKey:  scriptPubKey,
			AddressID:     dbAddress.ID,
		}
		dbResult := dbTx.Create(&dbTransactionOutput)
		if utils.HasDBError(dbResult) {
			return utils.NewErrorFromDBErrors("failed to insert transactionOutput: ", dbResult.GetErrors())
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
	var dbBlock models.Block
	dbResult := dbTx.
		Where(&models.Block{BlockHash: removedHash}).
		First(&dbBlock)
	if utils.HasDBError(dbResult) {
		return utils.NewErrorFromDBErrors("failed to find block: ", dbResult.GetErrors())
	}
	if utils.HasDBRecordNotFoundError(dbResult) {
		return fmt.Errorf("missing block for hash: %s", removedHash)
	}
	if !dbBlock.IsChainBlock {
		return fmt.Errorf("block erroneously marked as not a chain block: %s", removedHash)
	}

	var dbTransactions []models.Transaction
	dbResult = dbTx.
		Where(&models.Transaction{AcceptingBlockID: &dbBlock.ID}).
		Preload("TransactionInputs.PreviousTransactionOutput").
		Find(&dbTransactions)
	if utils.HasDBError(dbResult) {
		return utils.NewErrorFromDBErrors("failed to find transactions: ", dbResult.GetErrors())
	}
	for _, dbTransaction := range dbTransactions {
		for _, dbTransactionInput := range dbTransaction.TransactionInputs {
			dbPreviousTransactionOutput := dbTransactionInput.PreviousTransactionOutput
			if !dbPreviousTransactionOutput.IsSpent {
				return fmt.Errorf("cannot de-spend an unspent transaction output: %s index: %d",
					dbTransaction.TransactionID, dbTransactionInput.Index)
			}

			dbPreviousTransactionOutput.IsSpent = false
			dbResult = dbTx.Save(&dbPreviousTransactionOutput)
			if utils.HasDBError(dbResult) {
				return utils.NewErrorFromDBErrors("failed to update transactionOutput: ", dbResult.GetErrors())
			}
		}

		dbTransaction.AcceptingBlockID = nil
		dbResult := dbTx.Save(&dbTransaction)
		if utils.HasDBError(dbResult) {
			return utils.NewErrorFromDBErrors("failed to update transaction: ", dbResult.GetErrors())
		}
	}

	dbBlock.IsChainBlock = false
	dbResult = dbTx.Save(&dbBlock)
	if utils.HasDBError(dbResult) {
		return utils.NewErrorFromDBErrors("failed to update block: ", dbResult.GetErrors())
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
	for _, acceptedBlock := range addedBlock.AcceptedBlocks {
		var dbAccepedBlock models.Block
		dbResult := dbTx.
			Where(&models.Block{BlockHash: acceptedBlock.Hash}).
			First(&dbAccepedBlock)
		if utils.HasDBError(dbResult) {
			return utils.NewErrorFromDBErrors("failed to find block: ", dbResult.GetErrors())
		}
		if utils.HasDBRecordNotFoundError(dbResult) {
			return fmt.Errorf("missing block for hash: %s", acceptedBlock.Hash)
		}
		if dbAccepedBlock.IsChainBlock {
			return fmt.Errorf("block erroneously marked as a chain block: %s", acceptedBlock.Hash)
		}

		dbWhereTransactionIDsIn := make([]*models.Transaction, len(acceptedBlock.AcceptedTxIDs))
		for i, acceptedTxID := range acceptedBlock.AcceptedTxIDs {
			dbWhereTransactionIDsIn[i] = &models.Transaction{TransactionID: acceptedTxID}
		}
		var dbAcceptedTransactions []models.Transaction
		dbResult = dbTx.
			Where(dbWhereTransactionIDsIn).
			Preload("TransactionInputs.PreviousTransactionOutput").
			First(&dbAcceptedTransactions)
		if utils.HasDBError(dbResult) {
			return utils.NewErrorFromDBErrors("failed to find transactions: ", dbResult.GetErrors())
		}
		if len(dbAcceptedTransactions) != len(acceptedBlock.AcceptedTxIDs) {
			return fmt.Errorf("some transaction are missing for block: %s", acceptedBlock.Hash)
		}

		for _, dbAcceptedTransaction := range dbAcceptedTransactions {
			for _, dbTransactionInput := range dbAcceptedTransaction.TransactionInputs {
				dbPreviousTransactionOutput := dbTransactionInput.PreviousTransactionOutput
				if dbPreviousTransactionOutput.IsSpent {
					return fmt.Errorf("cannot spend an already spent transaction output: %s index: %d",
						dbAcceptedTransaction.TransactionID, dbTransactionInput.Index)
				}

				dbPreviousTransactionOutput.IsSpent = true
				dbResult = dbTx.Save(&dbPreviousTransactionOutput)
				if utils.HasDBError(dbResult) {
					return utils.NewErrorFromDBErrors("failed to update transactionOutput: ", dbResult.GetErrors())
				}
			}

			dbAcceptedTransaction.AcceptingBlockID = &dbAccepedBlock.ID
			dbResult = dbTx.Save(&dbAcceptedTransaction)
			if utils.HasDBError(dbResult) {
				return utils.NewErrorFromDBErrors("failed to update transaction: ", dbResult.GetErrors())
			}
		}

		dbAccepedBlock.IsChainBlock = true
		dbResult = dbTx.Save(&dbAccepedBlock)
		if utils.HasDBError(dbResult) {
			return utils.NewErrorFromDBErrors("failed to update block: ", dbResult.GetErrors())
		}
	}
	return nil
}

// sync keeps the API server in sync with the node via notifications
func sync(client *jsonrpc.Client, doneChan chan struct{}) {
	// ChainChangedMsgs must be processed in order and there may be times
	// when we may not be able to process them (e.g. appropriate
	// BlockAddedMsgs haven't arrived yet). As such, we pop messages from
	// client.OnChainChanged, make sure we're able to handle them, and
	// only then push them into nextChainChangedChan for them to be
	// actually handled.
	blockAddedMsgHandledChan := make(chan struct{})
	nextChainChangedChan := make(chan *jsonrpc.ChainChangedMsg)
	spawn(func() {
		for chainChanged := range client.OnChainChanged {
			for range blockAddedMsgHandledChan {
				canHandle, err := canHandleChainChangedMsg(chainChanged)
				if err != nil {
					panic(err)
				}
				if canHandle {
					break
				}
			}
			nextChainChangedChan <- chainChanged
		}
	})

	// Handle client notifications until we're told to stop
loop:
	for {
		select {
		case blockAdded := <-client.OnBlockAdded:
			handleBlockAddedMsg(client, blockAdded)
			blockAddedMsgHandledChan <- struct{}{}
		case chainChanged := <-nextChainChangedChan:
			handleChainChangedMsg(chainChanged)
		case <-doneChan:
			log.Infof("startSync stopped")
			break loop
		}
	}
}

// handleBlockAddedMsg handles onBlockAdded messages
func handleBlockAddedMsg(client *jsonrpc.Client, blockAdded *jsonrpc.BlockAddedMsg) {
	hash := blockAdded.Header.BlockHash()
	block, rawBlock, err := fetchBlock(client, hash)
	if err != nil {
		log.Warnf("Could not fetch block %s: %s", hash, err)
		return
	}
	err = addBlock(client, block, *rawBlock)
	if err != nil {
		log.Warnf("Could not insert block %s: %s", hash, err)
		return
	}
	log.Infof("Added block %s", hash)
}

// canHandleChainChangedMsg checks whether we have all the necessary data
// to successfully handle a ChainChangedMsg.
func canHandleChainChangedMsg(chainChanged *jsonrpc.ChainChangedMsg) (bool, error) {
	dbTx, err := database.DB()
	if err != nil {
		return false, err
	}

	// Collect all the referenced block hashes
	hashes := make(map[string]struct{})
	for _, removedHash := range chainChanged.RemovedChainBlockHashes {
		hashes[removedHash.String()] = struct{}{}
	}
	for _, addedBlock := range chainChanged.AddedChainBlocks {
		hashes[addedBlock.Hash.String()] = struct{}{}
		for _, acceptedBlock := range addedBlock.AcceptedBlocks {
			hashes[acceptedBlock.Hash.String()] = struct{}{}
		}
	}

	// Make sure that all the hashes exist in the database
	for hash := range hashes {
		var dbBlock []models.Block
		dbResult := dbTx.
			Where(&models.Block{BlockHash: hash}).
			Find(&dbBlock)
		if utils.HasDBError(dbResult) {
			return false, utils.NewErrorFromDBErrors("failed to find block: ", dbResult.GetErrors())
		}
		if utils.HasDBRecordNotFoundError(dbResult) {
			return false, nil
		}
	}

	return true, nil
}

// handleChainChangedMsg handles onChainChanged messages
func handleChainChangedMsg(chainChanged *jsonrpc.ChainChangedMsg) {
	// Convert the data in chainChanged to something we can feed into
	// updateSelectedParentChain
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

	err := updateSelectedParentChain(removedHashes, addedBlocks)
	if err != nil {
		log.Warnf("Could not update selected parent chain: %s", err)
		return
	}
	log.Infof("Chain changed: removed &d blocks and added %d block",
		len(removedHashes), len(addedBlocks))
}
