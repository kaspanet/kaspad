package externalapi

// DomainCoinbaseData contains data by which a coinbase transaction
// is built
type DomainCoinbaseData struct {
	scriptPublicKey []byte
	extraData       []byte
}
