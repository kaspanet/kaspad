package externalapi

// CoinbaseData contains data by which a coinbase transaction
// is built
type CoinbaseData struct {
	scriptPublicKey []byte
	extraData       []byte
}
