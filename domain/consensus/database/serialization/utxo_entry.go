package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func UTXOEntryToDBUTXOEntry(utxoEntry *externalapi.UTXOEntry) *DbUtxoEntry {
	return &DbUtxoEntry{
		Amount:          utxoEntry.Amount,
		ScriptPublicKey: utxoEntry.ScriptPublicKey,
		BlockBlueScore:  utxoEntry.BlockBlueScore,
		IsCoinbase:      utxoEntry.IsCoinbase,
	}
}

func DBUTXOEntryToUTXOEntry(dbUtxoEntry *DbUtxoEntry) *externalapi.UTXOEntry {
	return &externalapi.UTXOEntry{
		Amount:          dbUtxoEntry.Amount,
		ScriptPublicKey: dbUtxoEntry.ScriptPublicKey,
		BlockBlueScore:  dbUtxoEntry.BlockBlueScore,
		IsCoinbase:      dbUtxoEntry.IsCoinbase,
	}
}
