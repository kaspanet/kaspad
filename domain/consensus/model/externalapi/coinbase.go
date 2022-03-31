package externalapi

import "bytes"

// DomainCoinbaseData contains data by which a coinbase transaction
// is built
type DomainCoinbaseData struct {
	ScriptPublicKey *ScriptPublicKey
	ExtraData       []byte
}

// Clone returns a clone of DomainCoinbaseData
func (dcd *DomainCoinbaseData) Clone() *DomainCoinbaseData {

	scriptPubKeyClone := make([]byte, len(dcd.ScriptPublicKey.Script))
	copy(scriptPubKeyClone, dcd.ScriptPublicKey.Script)

	extraDataClone := make([]byte, len(dcd.ExtraData))
	copy(extraDataClone, dcd.ExtraData)

	return &DomainCoinbaseData{
		ScriptPublicKey: &ScriptPublicKey{Script: scriptPubKeyClone, Version: dcd.ScriptPublicKey.Version},
		ExtraData:       extraDataClone,
	}
}

// Equal returns whether dcd equals to other
func (dcd *DomainCoinbaseData) Equal(other *DomainCoinbaseData) bool {
	if dcd == nil || other == nil {
		return dcd == other
	}

	if !bytes.Equal(dcd.ExtraData, other.ExtraData) {
		return false
	}

	return dcd.ScriptPublicKey.Equal(other.ScriptPublicKey)
}
