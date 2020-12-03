package utxo

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type utxoEntry struct {
	amount          uint64
	scriptPublicKey []byte
	blockBlueScore  uint64
	isCoinbase      bool
}

// NewUTXOEntry creates a new utxoEntry representing the given txOut
func NewUTXOEntry(amount uint64, scriptPubKey []byte, isCoinbase bool, blockBlueScore uint64) externalapi.UTXOEntry {
	scriptPubKeyClone := make([]byte, len(scriptPubKey))
	copy(scriptPubKeyClone, scriptPubKey)
	return &utxoEntry{
		amount:          amount,
		scriptPublicKey: scriptPubKeyClone,
		blockBlueScore:  blockBlueScore,
		isCoinbase:      isCoinbase,
	}
}

func (u *utxoEntry) Amount() uint64 {
	return u.amount
}

func (u *utxoEntry) ScriptPublicKey() []byte {
	clone := make([]byte, len(u.scriptPublicKey))
	copy(clone, u.scriptPublicKey)
	return clone
}

func (u *utxoEntry) BlockBlueScore() uint64 {
	return u.blockBlueScore
}

func (u *utxoEntry) IsCoinbase() bool {
	return u.isCoinbase
}
