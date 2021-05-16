package server

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type walletUTXO struct {
	Outpoint  *externalapi.DomainOutpoint
	UTXOEntry externalapi.UTXOEntry
	address   *walletAddress
}

type walletAddress struct {
	index         uint32
	cosignerIndex uint32
	keyChain      uint8
}
