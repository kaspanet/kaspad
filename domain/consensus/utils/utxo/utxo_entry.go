package utxo

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type utxoEntry struct {
	amount          uint64
	scriptPublicKey *externalapi.ScriptPublicKey
	blockBlueScore  uint64
	isCoinbase      bool
}

// NewUTXOEntry creates a new utxoEntry representing the given txOut
func NewUTXOEntry(amount uint64, scriptPubKey *externalapi.ScriptPublicKey, isCoinbase bool, blockBlueScore uint64) externalapi.UTXOEntry {
	scriptPubKeyClone := externalapi.ScriptPublicKey{make([]byte, len(scriptPubKey.Script)), scriptPubKey.Version}
	copy(scriptPubKeyClone.Script, scriptPubKey.Script)
	return &utxoEntry{
		amount:          amount,
		scriptPublicKey: &scriptPubKeyClone,
		blockBlueScore:  blockBlueScore,
		isCoinbase:      isCoinbase,
	}
}

func (u *utxoEntry) Amount() uint64 {
	return u.amount
}

func (u *utxoEntry) ScriptPublicKey() *externalapi.ScriptPublicKey {
	clone := externalapi.ScriptPublicKey{make([]byte, len(u.scriptPublicKey.Script)), u.scriptPublicKey.Version}
	copy(clone.Script, u.scriptPublicKey.Script)
	return &clone
}

func (u *utxoEntry) BlockBlueScore() uint64 {
	return u.blockBlueScore
}

func (u *utxoEntry) IsCoinbase() bool {
	return u.isCoinbase
}
