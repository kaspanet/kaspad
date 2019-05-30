// Copyright (c) 2014-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcjson

import "encoding/json"

// GetBlockHeaderVerboseResult models the data from the getblockheader command when
// the verbose flag is set.  When the verbose flag is not set, getblockheader
// returns a hex-encoded string.
type GetBlockHeaderVerboseResult struct {
	Hash                 string   `json:"hash"`
	Confirmations        uint64   `json:"confirmations"`
	Height               uint64   `json:"height"`
	Version              int32    `json:"version"`
	VersionHex           string   `json:"versionHex"`
	HashMerkleRoot       string   `json:"hashMerkleRoot"`
	AcceptedIDMerkleRoot string   `json:"acceptedIdMerkleRoot"`
	Time                 int64    `json:"time"`
	Nonce                uint64   `json:"nonce"`
	Bits                 string   `json:"bits"`
	Difficulty           float64  `json:"difficulty"`
	ParentHashes         []string `json:"parentHashes,omitempty"`
	NextHashes           []string `json:"nextHashes,omitempty"`
}

// GetBlockVerboseResult models the data from the getblock command when the
// verbose flag is set.  When the verbose flag is not set, getblock returns a
// hex-encoded string.
type GetBlockVerboseResult struct {
	Hash                 string        `json:"hash"`
	Confirmations        uint64        `json:"confirmations"`
	Size                 int32         `json:"size"`
	Height               uint64        `json:"height"`
	Version              int32         `json:"version"`
	VersionHex           string        `json:"versionHex"`
	HashMerkleRoot       string        `json:"hashMerkleRoot"`
	AcceptedIDMerkleRoot string        `json:"acceptedIdMerkleRoot"`
	Tx                   []string      `json:"tx,omitempty"`
	RawTx                []TxRawResult `json:"rawRx,omitempty"`
	Time                 int64         `json:"time"`
	Nonce                uint64        `json:"nonce"`
	Bits                 string        `json:"bits"`
	Difficulty           float64       `json:"difficulty"`
	ParentHashes         []string      `json:"parentHashes"`
	NextHashes           []string      `json:"nextHashes,omitempty"`
}

// CreateMultiSigResult models the data returned from the createmultisig
// command.
type CreateMultiSigResult struct {
	Address      string `json:"address"`
	RedeemScript string `json:"redeemScript"`
}

// DecodeScriptResult models the data returned from the decodescript command.
type DecodeScriptResult struct {
	Asm       string   `json:"asm"`
	Type      string   `json:"type"`
	ReqSigs   int32    `json:"reqSigs,omitempty"`
	Addresses []string `json:"addresses,omitempty"`
	P2sh      string   `json:"p2sh,omitempty"`
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
	Fee     uint64  `json:"fee"`
	SigOps  int64   `json:"sigOps"`
}

// GetBlockTemplateResultAux models the coinbaseaux field of the
// getblocktemplate command.
type GetBlockTemplateResultAux struct {
	Flags string `json:"flags"`
}

// GetBlockTemplateResult models the data returned from the getblocktemplate
// command.
type GetBlockTemplateResult struct {
	// Base fields from BIP 0022.  CoinbaseAux is optional.  One of
	// CoinbaseTxn or CoinbaseValue must be specified, but not both.
	Bits                 string                     `json:"bits"`
	CurTime              int64                      `json:"curTime"`
	Height               uint64                     `json:"height"`
	ParentHashes         []string                   `json:"parentHashes"`
	SigOpLimit           int64                      `json:"sigOpLimit,omitempty"`
	SizeLimit            int64                      `json:"sizeLimit,omitempty"`
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
	SubmitOld   *bool  `json:"submitOld,omitempty"`

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

// GetMempoolEntryResult models the data returned from the getmempoolentry
// command.
type GetMempoolEntryResult struct {
	Size             int32    `json:"size"`
	Fee              float64  `json:"fee"`
	ModifiedFee      float64  `json:"modifiedFee"`
	Time             int64    `json:"time"`
	Height           uint64   `json:"height"`
	StartingPriority float64  `json:"startingPriority"`
	CurrentPriority  float64  `json:"currentPriority"`
	DescendantCount  int64    `json:"descendantCount"`
	DescendantSize   int64    `json:"descendantSize"`
	DescendantFees   float64  `json:"descendantFees"`
	AncestorCount    int64    `json:"ancestorCount"`
	AncestorSize     int64    `json:"ancestorSize"`
	AncestorFees     float64  `json:"ancestorFees"`
	Depends          []string `json:"depends"`
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
// command when the verbose flag is set.  When the verbose flag is not set,
// getrawmempool returns an array of transaction hashes.
type GetRawMempoolVerboseResult struct {
	Size             int32    `json:"size"`
	Fee              float64  `json:"fee"`
	Time             int64    `json:"time"`
	Height           uint64   `json:"height"`
	StartingPriority float64  `json:"startingPriority"`
	CurrentPriority  float64  `json:"currentPriority"`
	Depends          []string `json:"depends"`
}

// ScriptPubKeyResult models the scriptPubKey data of a tx script.  It is
// defined separately since it is used by multiple commands.
type ScriptPubKeyResult struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex,omitempty"`
	Type      string   `json:"type"`
	ReqSigs   int32    `json:"reqSigs,omitempty"`
	Addresses []string `json:"addresses,omitempty"`
}

// GetSubnetworkResult models the data from the getSubnetwork command.
type GetSubnetworkResult struct {
	GasLimit uint64 `json:"gasLimit"`
}

// GetTxOutResult models the data from the gettxout command.
type GetTxOutResult struct {
	BestBlock     string             `json:"bestBlock"`
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

// ScriptSig models a signature script.  It is defined separately since it only
// applies to non-coinbase.  Therefore the field in the Vin structure needs
// to be a pointer.
type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// Vin models parts of the tx data.  It is defined separately since
// getrawtransaction, decoderawtransaction, and searchrawtransaction use the
// same structure.
type Vin struct {
	Coinbase  string     `json:"coinbase"`
	TxID      string     `json:"txId"`
	Vout      uint32     `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig"`
	Sequence  uint64     `json:"sequence"`
}

// IsCoinBase returns a bool to show if a Vin is a Coinbase one or not.
func (v *Vin) IsCoinBase() bool {
	return len(v.Coinbase) > 0
}

// MarshalJSON provides a custom Marshal method for Vin.
func (v *Vin) MarshalJSON() ([]byte, error) {
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
	Addresses []string `json:"addresses,omitempty"`
	Value     float64  `json:"value"`
}

// VinPrevOut is like Vin except it includes PrevOut.  It is used by searchrawtransaction
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

// Vout models parts of the tx data.  It is defined separately since both
// getrawtransaction and decoderawtransaction use the same structure.
type Vout struct {
	Value        float64            `json:"value"`
	N            uint32             `json:"n"`
	ScriptPubKey ScriptPubKeyResult `json:"scriptPubKey"`
}

// GetMiningInfoResult models the data from the getmininginfo command.
type GetMiningInfoResult struct {
	Blocks           int64   `json:"blocks"`
	CurrentBlockSize uint64  `json:"currentBlockSize"`
	CurrentBlockTx   uint64  `json:"currentBlockTx"`
	Difficulty       float64 `json:"difficulty"`
	Errors           string  `json:"errors"`
	Generate         bool    `json:"generate"`
	GenProcLimit     int32   `json:"genProcLimit"`
	HashesPerSec     int64   `json:"hashesPerSec"`
	NetworkHashPS    int64   `json:"networkHashPs"`
	PooledTx         uint64  `json:"pooledTx"`
	TestNet          bool    `json:"testNet"`
	DevNet           bool    `json:"devNet"`
}

// GetWorkResult models the data from the getwork command.
type GetWorkResult struct {
	Data     string `json:"data"`
	Hash1    string `json:"hash1"`
	Midstate string `json:"midstate"`
	Target   string `json:"target"`
}

// InfoDAGResult models the data returned by the dag server getinfo command.
type InfoDAGResult struct {
	Version         int32   `json:"version"`
	ProtocolVersion int32   `json:"protocolVersion"`
	Blocks          uint64  `json:"blocks"`
	TimeOffset      int64   `json:"timeOffset"`
	Connections     int32   `json:"connections"`
	Proxy           string  `json:"proxy"`
	Difficulty      float64 `json:"difficulty"`
	TestNet         bool    `json:"testNet"`
	DevNet          bool    `json:"devNet"`
	RelayFee        float64 `json:"relayFee"`
	Errors          string  `json:"errors"`
}

// TxRawResult models the data from the getrawtransaction command.
type TxRawResult struct {
	Hex           string  `json:"hex"`
	TxID          string  `json:"txId"`
	Hash          string  `json:"hash,omitempty"`
	Size          int32   `json:"size,omitempty"`
	Version       int32   `json:"version"`
	LockTime      uint64  `json:"lockTime"`
	Subnetwork    string  `json:"subnetwork"`
	Gas           uint64  `json:"gas"`
	PayloadHash   string  `json:"payloadHash"`
	Payload       string  `json:"payload"`
	Vin           []Vin   `json:"vin"`
	Vout          []Vout  `json:"vout"`
	BlockHash     string  `json:"blockHash,omitempty"`
	Confirmations *uint64 `json:"confirmations,omitempty"`
	AcceptedBy    *string `json:"acceptedBy,omitempty"`
	IsInMempool   bool    `json:"isInMempool"`
	Time          uint64  `json:"time,omitempty"`
	BlockTime     uint64  `json:"blockTime,omitempty"`
}

// SearchRawTransactionsResult models the data from the searchrawtransaction
// command.
type SearchRawTransactionsResult struct {
	Hex           string       `json:"hex,omitempty"`
	TxID          string       `json:"txId"`
	Hash          string       `json:"hash"`
	Size          string       `json:"size"`
	Version       int32        `json:"version"`
	LockTime      uint64       `json:"lockTime"`
	Vin           []VinPrevOut `json:"vin"`
	Vout          []Vout       `json:"vout"`
	BlockHash     string       `json:"blockHash,omitempty"`
	Confirmations *uint64      `json:"confirmations,omitempty"`
	IsInMempool   bool         `json:"isInMempool"`
	Time          uint64       `json:"time,omitempty"`
	Blocktime     uint64       `json:"blockTime,omitempty"`
}

// TxRawDecodeResult models the data from the decoderawtransaction command.
type TxRawDecodeResult struct {
	TxID     string `json:"txId"`
	Version  int32  `json:"version"`
	Locktime uint64 `json:"lockTime"`
	Vin      []Vin  `json:"vin"`
	Vout     []Vout `json:"vout"`
}

// ValidateAddressResult models the data returned by the dag server
// validateaddress command.
type ValidateAddressResult struct {
	IsValid bool   `json:"isValid"`
	Address string `json:"address,omitempty"`
}
