package main

import (
	"bytes"
	"encoding/hex"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/dbmodels"
	"github.com/daglabs/btcd/apiserver/jsonrpc"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/config"
	"github.com/daglabs/btcd/httpserverutils"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
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
	log.Infof("Finished syncing past data")

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
			for {
				<-blockAddedMsgHandledChan
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

	var block dbmodels.Block
	dbQuery := dbTx.Order("blue_score DESC")
	if mustBeChainBlock {
		dbQuery = dbQuery.Where(&dbmodels.Block{IsChainBlock: true})
	}
	dbResult := dbQuery.First(&block)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("failed to find hash of bluest block: ", dbErrors)
	}
	if httpserverutils.IsDBRecordNotFoundError(dbErrors) {
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

	blockMass := uint64(0)
	for i, transaction := range rawBlock.RawTx {
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

	dbTx.Commit()
	return nil
}

func insertBlock(dbTx *gorm.DB, rawBlock btcjson.GetBlockVerboseResult) (*dbmodels.Block, error) {
	bits, err := strconv.ParseUint(rawBlock.Bits, 16, 32)
	if err != nil {
		return nil, err
	}
	dbBlock := dbmodels.Block{
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

	// Set genesis block as the initial chain block
	if len(rawBlock.ParentHashes) == 0 {
		dbBlock.IsChainBlock = true
	}
	dbResult := dbTx.Create(&dbBlock)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return nil, httpserverutils.NewErrorFromDBErrors("failed to insert block: ", dbErrors)
	}
	return &dbBlock, nil
}

func insertBlockParents(dbTx *gorm.DB, rawBlock btcjson.GetBlockVerboseResult, dbBlock *dbmodels.Block) error {
	// Exit early if this is the genesis block
	if len(rawBlock.ParentHashes) == 0 {
		return nil
	}

	hashesIn := make([]string, len(rawBlock.ParentHashes))
	for i, parentHash := range rawBlock.ParentHashes {
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
	if len(dbParents) != len(rawBlock.ParentHashes) {
		return errors.Errorf("some parents are missing for block: %s", rawBlock.Hash)
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

func insertBlockData(dbTx *gorm.DB, block string, dbBlock *dbmodels.Block) error {
	blockData, err := hex.DecodeString(block)
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
		Where("`transactions`.`transactiond_id` = ? AND `transaction_outputs`.`index` = ?", input.TxID, input.Vout).
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
	_, addr, err := txscript.ExtractScriptPubKeyAddress(scriptPubKey, config.ActiveNetworkFlags.ActiveNetParams)
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

// handleBlockAddedMsg handles onBlockAdded messages
func handleBlockAddedMsg(client *jsonrpc.Client, blockAdded *jsonrpc.BlockAddedMsg) {
	hash := blockAdded.Header.BlockHash()
	log.Debugf("Got block %s from the RPC server", hash)
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

	// Collect all unique referenced block hashes
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
	hashesIn := make([]string, len(hashes))
	i := 0
	for hash := range hashes {
		hashesIn[i] = hash
		i++
	}
	var dbBlocksCount int
	dbResult := dbTx.
		Model(&dbmodels.Block{}).
		Where("block_hash in (?)", hashesIn).
		Count(&dbBlocksCount)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return false, httpserverutils.NewErrorFromDBErrors("failed to find block count: ", dbErrors)
	}
	if len(hashes) != dbBlocksCount {
		return false, nil
	}

	return true, nil
}

// handleChainChangedMsg handles onChainChanged messages
func handleChainChangedMsg(chainChanged *jsonrpc.ChainChangedMsg) {
	// Convert the data in chainChanged to something we can feed into
	// updateSelectedParentChain
	removedHashes, addedBlocks := convertChainChangedMsg(chainChanged)

	err := updateSelectedParentChain(removedHashes, addedBlocks)
	if err != nil {
		log.Warnf("Could not update selected parent chain: %s", err)
		return
	}
	log.Infof("Chain changed: removed %d blocks and added %d block",
		len(removedHashes), len(addedBlocks))
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
