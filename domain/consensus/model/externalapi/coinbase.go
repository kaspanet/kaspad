package externalapi

// DomainCoinbaseData contains data by which a coinbase transaction
// is built
type DomainCoinbaseData struct {
	ScriptPublicKey []byte
	ExtraData       []byte
}

// Clone returns a clone of DomainCoinbaseData
func (dcd *DomainCoinbaseData) Clone() *DomainCoinbaseData {
	scriptPubKeyClone := make([]byte, len(dcd.ScriptPublicKey))
	copy(scriptPubKeyClone, dcd.ScriptPublicKey)

	extraDataClone := make([]byte, len(dcd.ExtraData))
	copy(extraDataClone, dcd.ExtraData)

	return &DomainCoinbaseData{
		ScriptPublicKey: scriptPubKeyClone,
		ExtraData:       extraDataClone,
	}
}
