package externalapi

// CoinbaseData contains data by which a coinbase transaction
// is built in a new block
type CoinbaseData struct {
	scriptPublicKey []byte
	extraData       []byte
}
