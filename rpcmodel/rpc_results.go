// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcmodel

import "encoding/json"

// GetBlockHeaderVerboseResult models the data from the getblockheader command when
// the verbose flag is set. When the verbose flag is not set, getblockheader
// returns a hex-encoded string.
type GetBlockHeaderVerboseResult struct {
	Hash                 string   `json:"hash"`
	Confirmations        uint64   `json:"confirmations"`
	Version              int32    `json:"version"`
	VersionHex           string   `json:"versionHex"`
	HashMerkleRoot       string   `json:"hashMerkleRoot"`
	AcceptedIDMerkleRoot string   `json:"acceptedIdMerkleRoot"`
	Time                 int64    `json:"time"`
	Nonce                uint64   `json:"nonce"`
	Bits                 string   `json:"bits"`
	Difficulty           float64  `json:"difficulty"`
	ParentHashes         []string `json:"parentHashes,omitempty"`
	SelectedParentHash   string   `json:"selectedParentHash"`
	ChildHashes          []string `json:"childHashes,omitempty"`
}

// GetBlockVerboseResult models the data from the getblock command when the
// verbose flag is set. When the verbose flag is not set, getblock returns a
// hex-encoded string.
type GetBlockVerboseResult struct {
	Hash                 string        `json:"hash"`
	Confirmations        uint64        `json:"confirmations"`
	Size                 int32         `json:"size"`
	BlueScore            uint64        `json:"blueScore"`
	IsChainBlock         bool          `json:"isChainBlock"`
	Version              int32         `json:"version"`
	VersionHex           string        `json:"versionHex"`
	HashMerkleRoot       string        `json:"hashMerkleRoot"`
	AcceptedIDMerkleRoot string        `json:"acceptedIdMerkleRoot"`
	UTXOCommitment       string        `json:"utxoCommitment"`
	Tx                   []string      `json:"tx,omitempty"`
	RawTx                []TxRawResult `json:"rawRx,omitempty"`
	Time                 int64         `json:"time"`
	Nonce                uint64        `json:"nonce"`
	Bits                 string        `json:"bits"`
	Difficulty           float64       `json:"difficulty"`
	ParentHashes         []string      `json:"parentHashes"`
	SelectedParentHash   string        `json:"selectedParentHash,omitempty"`
	ChildHashes          []string      `json:"childHashes,omitempty"`
}

// CreateMultiSigResult models the data returned from the createmultisig
// command.
type CreateMultiSigResult struct {
	Address      string `json:"address"`
	RedeemScript string `json:"redeemScript"`
}

// DecodeScriptResult models the data returned from the decodescript command.
type DecodeScriptResult struct {
	Asm     string  `json:"asm"`
	Type    string  `json:"type"`
	Address *string `json:"address,omitempty"`
	P2sh    string  `json:"p2sh,omitempty"`
}

// GetManualNodeInfoResultAddr models the data of the addresses portion of the
// getmanualnodeinfo command.
type GetManualNodeInfoResultAddr struct {
	Address   string `json:"address"`
	Connected string `json:"connected"`
}

// GetManualNodeInfoResult models the data from the getmanualnodeinfo command.
type GetManualNodeInfoResult struct {
	ManualNode string                         `json:"manualNode"`
	Connected  *bool                          `json:"connected,omitempty"`
	Addresses  *[]GetManualNodeInfoResultAddr `json:"addresses,omitempty"`
}

// SoftForkDescription describes the current state of a soft-fork which was
// deployed using a super-majority block signalling.
type SoftForkDescription struct {
	ID      string `json:"id"`
	Version uint32 `json:"version"`
	Reject  struct {
		Status bool `json:"status"`
	} `json:"reject"`
}

// Bip9SoftForkDescription describes the current state of a defined BIP0009
// version bits soft-fork.
type Bip9SoftForkDescription struct {
	Status    string `json:"status"`
	Bit       uint8  `json:"bit"`
	StartTime int64  `json:"startTime"`
	Timeout   int64  `json:"timeout"`
	Since     int32  `json:"since"`
}

// GetBlockDAGInfoResult models the data returned from the getblockdaginfo
// command.
type GetBlockDAGInfoResult struct {
	DAG                  string                              `json:"dag"`
	Blocks               uint64                              `json:"blocks"`
	Headers              uint64                              `json:"headers"`
	TipHashes            []string                            `json:"tipHashes"`
	Difficulty           float64                             `json:"difficulty"`
	MedianTime           int64                               `json:"medianTime"`
	UTXOCommitment       string                              `json:"utxoCommitment"`
	VerificationProgress float64                             `json:"verificationProgress,omitempty"`
	Pruned               bool                                `json:"pruned"`
	PruneHeight          uint64                              `json:"pruneHeight,omitempty"`
	DAGWork              string                              `json:"dagWork,omitempty"`
	SoftForks            []*SoftForkDescription              `json:"softForks"`
	Bip9SoftForks        map[string]*Bip9SoftForkDescription `json:"bip9SoftForks"`
}

// GetBlockTemplateResultTx models the transactions field of the
// getblocktemplate command.
type GetBlockTemplateResultTx struct {
	Data    string  `json:"data"`
	ID      string  `json:"id"`
	Depends []int64 `json:"depends"`
	Mass    uint64  `json:"mass"`
	Fee     uint64  `json:"fee"`
}

// GetBlockTemplateResultAux models the coinbaseaux field of the
// getblocktemplate command.
type GetBlockTemplateResultAux struct {
	Flags string `json:"flags"`
}

// GetBlockTemplateResult models the data returned from the getblocktemplate
// command.
type GetBlockTemplateResult struct {
	// Base fields from BIP 0022. CoinbaseAux is optional. One of
	// CoinbaseTxn or CoinbaseValue must be specified, but not both.
	Bits                 string                     `json:"bits"`
	CurTime              int64                      `json:"curTime"`
	Height               uint64                     `json:"height"`
	ParentHashes         []string                   `json:"parentHashes"`
	MassLimit            int64                      `json:"massLimit,omitempty"`
	Transactions         []GetBlockTemplateResultTx `json:"transactions"`
	AcceptedIDMerkleRoot string                     `json:"acceptedIdMerkleRoot"`
	UTXOCommitment       string                     `json:"utxoCommitment"`
	Version              int32                      `json:"version"`
	CoinbaseAux          *GetBlockTemplateResultAux `json:"coinbaseAux,omitempty"`
	CoinbaseTxn          *GetBlockTemplateResultTx  `json:"coinbaseTxn,omitempty"`
	CoinbaseValue        *uint64                    `json:"coinbaseValue,omitempty"`
	WorkID               string                     `json:"workId,omitempty"`

	// Optional long polling from BIP 0022.
	LongPollID  string `json:"longPollId,omitempty"`
	LongPollURI string `json:"longPollUri,omitempty"`

	// Basic pool extension from BIP 0023.
	Target  string `json:"target,omitempty"`
	Expires int64  `json:"expires,omitempty"`

	// Mutations from BIP 0023.
	MaxTime    int64    `json:"maxTime,omitempty"`
	MinTime    int64    `json:"minTime,omitempty"`
	Mutable    []string `json:"mutable,omitempty"`
	NonceRange string   `json:"nonceRange,omitempty"`

	// Block proposal from BIP 0023.
	Capabilities  []string `json:"capabilities,omitempty"`
	RejectReasion string   `json:"rejectReason,omitempty"`
}

// GetMempoolEntryResult models the data returned from the getMempoolEntry
// command.
type GetMempoolEntryResult struct {
	Fee   uint64      `json:"fee"`
	Time  int64       `json:"time"`
	RawTx TxRawResult `json:"rawTx"`
}

// GetMempoolInfoResult models the data returned from the getmempoolinfo
// command.
type GetMempoolInfoResult struct {
	Size  int64 `json:"size"`
	Bytes int64 `json:"bytes"`
}

// NetworksResult models the networks data from the getnetworkinfo command.
type NetworksResult struct {
	Name                      string `json:"name"`
	Limited                   bool   `json:"limited"`
	Reachable                 bool   `json:"reachable"`
	Proxy                     string `json:"proxy"`
	ProxyRandomizeCredentials bool   `json:"proxyRandomizeCredentials"`
}

// LocalAddressesResult models the localaddresses data from the getnetworkinfo
// command.
type LocalAddressesResult struct {
	Address string `json:"address"`
	Port    uint16 `json:"port"`
	Score   int32  `json:"score"`
}

// GetNetworkInfoResult models the data returned from the getnetworkinfo
// command.
type GetNetworkInfoResult struct {
	Version         int32                  `json:"version"`
	SubVersion      string                 `json:"subVersion"`
	ProtocolVersion int32                  `json:"protocolVersion"`
	LocalServices   string                 `json:"localServices"`
	LocalRelay      bool                   `json:"localRelay"`
	TimeOffset      int64                  `json:"timeOffset"`
	Connections     int32                  `json:"connections"`
	NetworkActive   bool                   `json:"networkActive"`
	Networks        []NetworksResult       `json:"networks"`
	RelayFee        float64                `json:"relayFee"`
	IncrementalFee  float64                `json:"incrementalFee"`
	LocalAddresses  []LocalAddressesResult `json:"localAddresses"`
	Warnings        string                 `json:"warnings"`
}

// GetPeerInfoResult models the data returned from the getpeerinfo command.
type GetPeerInfoResult struct {
	ID          int32   `json:"id"`
	Addr        string  `json:"addr"`
	Services    string  `json:"services"`
	RelayTxes   bool    `json:"relayTxes"`
	LastSend    int64   `json:"lastSend"`
	LastRecv    int64   `json:"lastRecv"`
	BytesSent   uint64  `json:"bytesSent"`
	BytesRecv   uint64  `json:"bytesRecv"`
	ConnTime    int64   `json:"connTime"`
	TimeOffset  int64   `json:"timeOffset"`
	PingTime    float64 `json:"pingTime"`
	PingWait    float64 `json:"pingWait,omitempty"`
	Version     uint32  `json:"version"`
	SubVer      string  `json:"subVer"`
	Inbound     bool    `json:"inbound"`
	SelectedTip string  `json:"selectedTip,omitempty"`
	BanScore    int32   `json:"banScore"`
	FeeFilter   int64   `json:"feeFilter"`
	SyncNode    bool    `json:"syncNode"`
}

// GetRawMempoolVerboseResult models the data returned from the getrawmempool
// command when the verbose flag is set. When the verbose flag is not set,
// getrawmempool returns an array of transaction hashes.
type GetRawMempoolVerboseResult struct {
	Size    int32    `json:"size"`
	Fee     float64  `json:"fee"`
	Time    int64    `json:"time"`
	Depends []string `json:"depends"`
}

// ScriptPubKeyResult models the scriptPubKey data of a tx script. It is
// defined separately since it is used by multiple commands.
type ScriptPubKeyResult struct {
	Asm     string  `json:"asm"`
	Hex     string  `json:"hex,omitempty"`
	Type    string  `json:"type"`
	Address *string `json:"address,omitempty"`
}

// GetSubnetworkResult models the data from the getSubnetwork command.
type GetSubnetworkResult struct {
	GasLimit *uint64 `json:"gasLimit"`
}

// GetTxOutResult models the data from the gettxout command.
type GetTxOutResult struct {
	SelectedTip   string             `json:"selectedTip"`
	Confirmations *uint64            `json:"confirmations,omitempty"`
	IsInMempool   bool               `json:"isInMempool"`
	Value         float64            `json:"value"`
	ScriptPubKey  ScriptPubKeyResult `json:"scriptPubKey"`
	Coinbase      bool               `json:"coinbase"`
}

// GetNetTotalsResult models the data returned from the getnettotals command.
type GetNetTotalsResult struct {
	TotalBytesRecv uint64 `json:"totalBytesRecv"`
	TotalBytesSent uint64 `json:"totalBytesSent"`
	TimeMillis     int64  `json:"timeMillis"`
}

// ScriptSig models a signature script. It is defined separately since it only
// applies to non-coinbase. Therefore the field in the Vin structure needs
// to be a pointer.
type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// Vin models parts of the tx data.
type Vin struct {
	TxID      string     `json:"txId"`
	Vout      uint32     `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig"`
	Sequence  uint64     `json:"sequence"`
}

// MarshalJSON provides a custom Marshal method for Vin.
func (v *Vin) MarshalJSON() ([]byte, error) {
	txStruct := struct {
		TxID      string     `json:"txId"`
		Vout      uint32     `json:"vout"`
		ScriptSig *ScriptSig `json:"scriptSig"`
		Sequence  uint64     `json:"sequence"`
	}{
		TxID:      v.TxID,
		Vout:      v.Vout,
		ScriptSig: v.ScriptSig,
		Sequence:  v.Sequence,
	}
	return json.Marshal(txStruct)
}

// PrevOut represents previous output for an input Vin.
type PrevOut struct {
	Address *string `json:"address,omitempty"`
	Value   float64 `json:"value"`
}

// VinPrevOut is like Vin except it includes PrevOut.
type VinPrevOut struct {
	Coinbase  string     `json:"coinbase"`
	TxID      string     `json:"txId"`
	Vout      uint32     `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig"`
	PrevOut   *PrevOut   `json:"prevOut"`
	Sequence  uint64     `json:"sequence"`
}

// IsCoinBase returns a bool to show if a Vin is a Coinbase one or not.
func (v *VinPrevOut) IsCoinBase() bool {
	return len(v.Coinbase) > 0
}

// MarshalJSON provides a custom Marshal method for VinPrevOut.
func (v *VinPrevOut) MarshalJSON() ([]byte, error) {
	if v.IsCoinBase() {
		coinbaseStruct := struct {
			Coinbase string `json:"coinbase"`
			Sequence uint64 `json:"sequence"`
		}{
			Coinbase: v.Coinbase,
			Sequence: v.Sequence,
		}
		return json.Marshal(coinbaseStruct)
	}

	txStruct := struct {
		TxID      string     `json:"txId"`
		Vout      uint32     `json:"vout"`
		ScriptSig *ScriptSig `json:"scriptSig"`
		PrevOut   *PrevOut   `json:"prevOut,omitempty"`
		Sequence  uint64     `json:"sequence"`
	}{
		TxID:      v.TxID,
		Vout:      v.Vout,
		ScriptSig: v.ScriptSig,
		PrevOut:   v.PrevOut,
		Sequence:  v.Sequence,
	}
	return json.Marshal(txStruct)
}

// Vout models parts of the tx data
type Vout struct {
	Value        uint64             `json:"value"`
	N            uint32             `json:"n"`
	ScriptPubKey ScriptPubKeyResult `json:"scriptPubKey"`
}

// GetWorkResult models the data from the getwork command.
type GetWorkResult struct {
	Data     string `json:"data"`
	Hash1    string `json:"hash1"`
	Midstate string `json:"midstate"`
	Target   string `json:"target"`
}

// InfoDAGResult models the data returned by the kaspa rpc server getinfo command.
type InfoDAGResult struct {
	Version         string  `json:"version"`
	ProtocolVersion int32   `json:"protocolVersion"`
	Blocks          uint64  `json:"blocks"`
	Connections     int32   `json:"connections"`
	Proxy           string  `json:"proxy"`
	Difficulty      float64 `json:"difficulty"`
	Testnet         bool    `json:"testnet"`
	Devnet          bool    `json:"devnet"`
	RelayFee        float64 `json:"relayFee"`
	Errors          string  `json:"errors"`
}

// TxRawResult models transaction result data.
type TxRawResult struct {
	Hex         string  `json:"hex"`
	TxID        string  `json:"txId"`
	Hash        string  `json:"hash,omitempty"`
	Size        int32   `json:"size,omitempty"`
	Version     int32   `json:"version"`
	LockTime    uint64  `json:"lockTime"`
	Subnetwork  string  `json:"subnetwork"`
	Gas         uint64  `json:"gas"`
	PayloadHash string  `json:"payloadHash"`
	Payload     string  `json:"payload"`
	Vin         []Vin   `json:"vin"`
	Vout        []Vout  `json:"vout"`
	BlockHash   string  `json:"blockHash,omitempty"`
	AcceptedBy  *string `json:"acceptedBy,omitempty"`
	IsInMempool bool    `json:"isInMempool"`
	Time        uint64  `json:"time,omitempty"`
	BlockTime   uint64  `json:"blockTime,omitempty"`
}

// TxRawDecodeResult models the data from the decoderawtransaction command.
type TxRawDecodeResult struct {
	TxID     string `json:"txId"`
	Version  int32  `json:"version"`
	Locktime uint64 `json:"lockTime"`
	Vin      []Vin  `json:"vin"`
	Vout     []Vout `json:"vout"`
}

// ValidateAddressResult models the data returned by the kaspa rpc server
// validateaddress command.
type ValidateAddressResult struct {
	IsValid bool   `json:"isValid"`
	Address string `json:"address,omitempty"`
}

// ChainBlock models a block that is part of the selected parent chain.
type ChainBlock struct {
	Hash           string          `json:"hash"`
	AcceptedBlocks []AcceptedBlock `json:"acceptedBlocks"`
}

// AcceptedBlock models a block that is included in the blues of a selected
// chain block.
type AcceptedBlock struct {
	Hash          string   `json:"hash"`
	AcceptedTxIDs []string `json:"acceptedTxIds"`
}

// GetChainFromBlockResult models the data from the getChainFromBlock command.
type GetChainFromBlockResult struct {
	RemovedChainBlockHashes []string                `json:"removedChainBlockHashes"`
	AddedChainBlocks        []ChainBlock            `json:"addedChainBlocks"`
	Blocks                  []GetBlockVerboseResult `json:"blocks"`
}

// GetBlocksResult models the data from the getBlocks command.
type GetBlocksResult struct {
	Hashes        []string                `json:"hashes"`
	RawBlocks     []string                `json:"rawBlocks"`
	VerboseBlocks []GetBlockVerboseResult `json:"verboseBlocks"`
}

// VersionResult models objects included in the version response. In the actual
// result, these objects are keyed by the program or API name.
type VersionResult struct {
	VersionString string `json:"versionString"`
	Major         uint32 `json:"major"`
	Minor         uint32 `json:"minor"`
	Patch         uint32 `json:"patch"`
	Prerelease    string `json:"prerelease"`
	BuildMetadata string `json:"buildMetadata"`
}
