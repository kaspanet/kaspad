package utxoindex

type UTXOIndex struct {
	store *utxoIndexStore
}

func New() *UTXOIndex {
	store := newUTXOIndexStore()
	return &UTXOIndex{
		store: store,
	}
}

func (ui *UTXOIndex) Update() {

}
