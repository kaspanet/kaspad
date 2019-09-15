package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/apiserver/database"
	"github.com/daglabs/btcd/apiserver/jsonrpc"
	"github.com/daglabs/btcd/apiserver/models"
	"github.com/daglabs/btcd/btcjson"
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
	db, err := database.DB()
	if err != nil {
		return err
	}

	mostRecentBlockHash := findMostRecentBlockHash(db)
	err = SyncBlocks(client, db, mostRecentBlockHash)
	if err != nil {
		return err
	}

	err = SyncSelectedParentChain(client, db, mostRecentBlockHash)
	if err != nil {
		return err
	}

loop:
	for {
		select {
		case blockAdded := <-client.OnBlockAdded:
			hash := blockAdded.Header.BlockHash()
			block, rawBlock, err := fetchBlock(client, hash)
			if err != nil {
				log.Warnf("Could not fetch block %s: %s", hash, err)
				continue
			}
			err = insertBlock(client, db, block, *rawBlock)
			if err != nil {
				log.Warnf("Could not insert block %s: %s", hash, err)
				continue
			}
			log.Infof("Added block %s", hash)
		case chainChanged := <-client.OnChainChanged:
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
			err := updateSelectedParentChain(db, removedHashes, addedBlocks)
			if err != nil {
				log.Warnf("Could not update selected parent chain: %s", err)
			}
		case <-doneChan:
			log.Infof("blockLoop stopped")
			break loop
		}
	}
	return nil
}

func findMostRecentBlockHash(db *gorm.DB) *string {
	var block models.Block
	db.Order("blue_score DESC").First(&block)

	if block.ID == 0 {
		return nil
	}
	return &block.BlockHash
}

func SyncBlocks(client *jsonrpc.Client, db *gorm.DB, startHash *string) error {
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

	return insertBlocks(client, db, blocks, rawBlocks)
}

func SyncSelectedParentChain(client *jsonrpc.Client, db *gorm.DB, startHash *string) error {
	for {
		chainFromBlockResult, err := client.GetChainFromBlock(false, startHash)
		if err != nil {
			return err
		}
		if len(chainFromBlockResult.AddedChainBlocks) == 0 {
			break
		}

		startHash = &chainFromBlockResult.AddedChainBlocks[len(chainFromBlockResult.AddedChainBlocks)].Hash
		err = updateSelectedParentChain(db,
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

func insertBlocks(client *jsonrpc.Client, db *gorm.DB, blocks []string, rawBlocks []btcjson.GetBlockVerboseResult) error {
	for i, rawBlock := range rawBlocks {
		block := blocks[i]
		err := insertBlock(client, db, block, rawBlock)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertBlock(client *jsonrpc.Client, db *gorm.DB, block string, rawBlock btcjson.GetBlockVerboseResult) error {
	db = db.Begin()

	// Insert the block
	var dbBlock models.Block
	db.Where(&models.Block{BlockHash: rawBlock.Hash}).First(&dbBlock)
	if dbBlock.ID == 0 {
		bits, err := strconv.ParseUint(rawBlock.Bits, 16, 32)
		if err != nil {
			return err
		}
		dbBlock = models.Block{
			BlockHash:            rawBlock.Hash,
			Version:              rawBlock.Version,
			HashMerkleRoot:       rawBlock.HashMerkleRoot,
			AcceptedIDMerkleRoot: rawBlock.AcceptedIDMerkleRoot,
			UTXOCommitment:       rawBlock.UTXOCommitment,
			Timestamp:            time.Unix(rawBlock.Time, 0),
			Bits:                 uint32(bits),
			Nonce:                rawBlock.Nonce,
			BlueScore:            rawBlock.BlueScore,
			IsChainBlock:         rawBlock.IsChainBlock,
		}
		db.Create(&dbBlock)
	}

	// Insert the block parents
	for _, parentHash := range rawBlock.ParentHashes {
		var dbParent models.Block
		db.Where(&models.Block{BlockHash: parentHash}).First(&dbParent)
		if dbParent.ID == 0 {
			return fmt.Errorf("missing parent for hash: %s", parentHash)
		}

		var dbParentBlock models.ParentBlock
		db.Where(&models.ParentBlock{BlockID: dbBlock.ID, ParentBlockID: dbParent.ID}).First(&dbParentBlock)
		if dbParentBlock.BlockID == 0 {
			dbParentBlock = models.ParentBlock{
				BlockID:       dbBlock.ID,
				ParentBlockID: dbParent.ID,
			}
			db.Create(&dbParentBlock)
		}
	}

	// Insert the block data
	var dbRawBlock models.RawBlock
	db.Where(&models.RawBlock{BlockID: dbBlock.ID}).First(&dbRawBlock)
	if dbRawBlock.BlockID == 0 {
		blockData, err := hex.DecodeString(block)
		if err != nil {
			return err
		}
		dbRawBlock = models.RawBlock{
			BlockID:   dbBlock.ID,
			BlockData: blockData,
		}
		db.Create(&dbRawBlock)
	}

	for i, transaction := range rawBlock.RawTx {
		// Insert the subnetwork
		var dbSubnetwork models.Subnetwork
		db.Where(&models.Subnetwork{SubnetworkID: transaction.Subnetwork}).First(&dbSubnetwork)
		if dbSubnetwork.ID == 0 {
			subnetwork, err := client.GetSubnetwork(transaction.Subnetwork)
			if err != nil {
				return err
			}
			dbSubnetwork = models.Subnetwork{
				SubnetworkID: transaction.Subnetwork,
				GasLimit:     subnetwork.GasLimit,
			}
			db.Create(&dbSubnetwork)
		}

		// Insert the transaction
		var dbTransaction models.Transaction
		db.Where(&models.Transaction{TransactionID: transaction.TxID}).First(&dbTransaction)
		if dbTransaction.ID == 0 {
			payload, err := hex.DecodeString(transaction.Payload)
			if err != nil {
				return err
			}
			dbTransaction = models.Transaction{
				TransactionHash: transaction.Hash,
				TransactionID:   transaction.TxID,
				LockTime:        transaction.LockTime,
				SubnetworkID:    dbSubnetwork.ID,
				Gas:             transaction.Gas,
				PayloadHash:     transaction.PayloadHash,
				Payload:         payload,
			}
			db.Create(&dbTransaction)
		}

		// Insert the transaction block
		var dbTransactionBlock models.TransactionBlock
		db.Where(&models.TransactionBlock{TransactionID: dbTransaction.ID, BlockID: dbBlock.ID}).First(&dbTransactionBlock)
		if dbTransactionBlock.TransactionID == 0 {
			dbTransactionBlock = models.TransactionBlock{
				TransactionID: dbTransaction.ID,
				BlockID:       dbBlock.ID,
				Index:         uint32(i),
			}
			db.Create(&dbTransactionBlock)
		}

		// Check whether this transaction is coinbase
		subnetwork, err := subnetworkid.NewFromStr(transaction.Subnetwork)
		if err != nil {
			return err
		}
		isCoinbase := subnetwork.IsEqual(subnetworkid.SubnetworkIDCoinbase)

		// Insert the transaction inputs
		if !isCoinbase {
			for _, input := range transaction.Vin {
				var dbOutputTransaction models.Transaction
				db.Where(&models.Transaction{TransactionID: input.TxID}).First(&dbOutputTransaction)
				if dbOutputTransaction.ID == 0 {
					return fmt.Errorf("missing output transaction for txID: %s", input.TxID)
				}

				var dbOutputTransactionOutput models.TransactionOutput
				db.Where(&models.TransactionOutput{TransactionID: dbOutputTransaction.ID, Index: input.Vout}).First(&dbOutputTransactionOutput)
				if dbOutputTransactionOutput.ID == 0 {
					return fmt.Errorf("missing output transaction output for txID: %s and index: %d", input.TxID, input.Vout)
				}

				var dbTransactionInput models.TransactionInput
				db.Where(models.TransactionInput{TransactionID: dbTransaction.ID, TransactionOutputID: dbOutputTransactionOutput.ID}).First(&dbTransactionInput)
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
					db.Create(&dbTransactionInput)
				}
			}
		}

		// Insert the transaction outputs
		for _, output := range transaction.Vout {
			var dbTransactionOutput models.TransactionOutput
			db.Where(&models.TransactionOutput{TransactionID: dbTransaction.ID, Index: output.N}).First(&dbTransactionOutput)
			if dbTransactionOutput.TransactionID == 0 {
				scriptPubKey, err := hex.DecodeString(output.ScriptPubKey.Hex)
				if err != nil {
					return err
				}
				dbTransactionOutput = models.TransactionOutput{
					TransactionID: dbTransaction.ID,
					Index:         output.N,
					Value:         output.Value,
					IsSpent:       false,
					ScriptPubKey:  scriptPubKey,
				}
				db.Create(&dbTransactionOutput)
			}
		}
	}

	db.Commit()
	return nil
}

func updateSelectedParentChain(db *gorm.DB, removedChainHashes []string, addedChainBlocks []btcjson.ChainBlock) error {
	db = db.Begin()
	for _, removedHash := range removedChainHashes {
		var dbBlock models.Block
		db.Where(&models.Block{BlockHash: removedHash}).First(&dbBlock)
		if dbBlock.ID == 0 {
			return fmt.Errorf("missing block for hash: %s", removedHash)
		}

		var dbTransactions []models.Transaction
		db.Where(&models.Transaction{AcceptingBlockID: &dbBlock.ID}).Preload("TransactionInputs").Find(&dbTransactions)
		for _, dbTransaction := range dbTransactions {
			for _, dbTransactionInput := range dbTransaction.TransactionInputs {
				var dbTransactionOutput models.TransactionOutput
				db.Where(&models.TransactionOutput{ID: dbTransactionInput.TransactionOutputID}).First(&dbTransactionOutput)
				if dbTransactionOutput.ID == 0 {
					return fmt.Errorf("missing transaction output for transaction: %s index: %d", dbTransaction.TransactionID, dbTransactionInput.Index)
				}
				if dbTransactionOutput.IsSpent == false {
					return fmt.Errorf("cannot de-spend an unspent transaction output")
				}

				dbTransactionOutput.IsSpent = false
				db.Save(&dbTransactionOutput)
			}

			dbTransaction.AcceptingBlockID = nil
			db.Save(&dbTransaction)
		}
	}
	for _, addedBlock := range addedChainBlocks {
		for _, acceptedBlock := range addedBlock.AcceptedBlocks {
			var dbAcceptingBlock models.Block
			db.Where(&models.Block{BlockHash: acceptedBlock.Hash}).First(dbAcceptingBlock)
			if dbAcceptingBlock.ID == 0 {
				return fmt.Errorf("missing block for hash: %s", acceptedBlock.Hash)
			}

			for _, acceptedTxID := range acceptedBlock.AcceptedTxIDs {
				var dbTransaction models.Transaction
				db.Where(&models.Transaction{TransactionID: acceptedTxID}).First(&dbTransaction)
				if dbTransaction.ID == 0 {
					return fmt.Errorf("missing transaction for txID: %s", acceptedTxID)
				}

				var dbTransactionInputs []models.TransactionInput
				db.Where(&models.TransactionInput{TransactionID: dbTransaction.ID}).Preload("TransactionInputs").Find(&dbTransactionInputs)
				for _, dbTransactionInput := range dbTransactionInputs {
					var dbTransactionOutput models.TransactionOutput
					db.Where(&models.TransactionOutput{ID: dbTransactionInput.TransactionOutputID}).First(&dbTransactionOutput)
					if dbTransactionOutput.ID == 0 {
						return fmt.Errorf("missing transaction output for transaction: %s index: %d", dbTransaction.TransactionID, dbTransactionInput.Index)
					}
					if dbTransactionOutput.IsSpent == true {
						return fmt.Errorf("cannot spend an already spent transaction output")
					}

					dbTransactionOutput.IsSpent = true
					db.Save(&dbTransactionOutput)
				}

				dbTransaction.AcceptingBlockID = &dbAcceptingBlock.ID
				db.Save(&dbTransaction)
			}
		}
	}

	db.Commit()
	return nil
}
