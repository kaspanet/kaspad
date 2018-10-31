// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC commands that are supported by
// a wallet server.

package btcjson

// AddMultisigAddressCmd defines the addmutisigaddress JSON-RPC command.
type AddMultisigAddressCmd struct {
	NRequired int
	Keys      []string
	Account   *string
}

// NewAddMultisigAddressCmd returns a new instance which can be used to issue a
// addMultisigAddress JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewAddMultisigAddressCmd(nRequired int, keys []string, account *string) *AddMultisigAddressCmd {
	return &AddMultisigAddressCmd{
		NRequired: nRequired,
		Keys:      keys,
		Account:   account,
	}
}

// CreateMultisigCmd defines the createMultisig JSON-RPC command.
type CreateMultisigCmd struct {
	NRequired int
	Keys      []string
}

// NewCreateMultisigCmd returns a new instance which can be used to issue a
// createMultisig JSON-RPC command.
func NewCreateMultisigCmd(nRequired int, keys []string) *CreateMultisigCmd {
	return &CreateMultisigCmd{
		NRequired: nRequired,
		Keys:      keys,
	}
}

// DumpPrivKeyCmd defines the dumpPrivKey JSON-RPC command.
type DumpPrivKeyCmd struct {
	Address string
}

// NewDumpPrivKeyCmd returns a new instance which can be used to issue a
// dumpPrivKey JSON-RPC command.
func NewDumpPrivKeyCmd(address string) *DumpPrivKeyCmd {
	return &DumpPrivKeyCmd{
		Address: address,
	}
}

// EncryptWalletCmd defines the encryptWallet JSON-RPC command.
type EncryptWalletCmd struct {
	Passphrase string
}

// NewEncryptWalletCmd returns a new instance which can be used to issue a
// encryptWallet JSON-RPC command.
func NewEncryptWalletCmd(passphrase string) *EncryptWalletCmd {
	return &EncryptWalletCmd{
		Passphrase: passphrase,
	}
}

// EstimateFeeCmd defines the estimateFee JSON-RPC command.
type EstimateFeeCmd struct {
	NumBlocks int64
}

// NewEstimateFeeCmd returns a new instance which can be used to issue a
// estimateFee JSON-RPC command.
func NewEstimateFeeCmd(numBlocks int64) *EstimateFeeCmd {
	return &EstimateFeeCmd{
		NumBlocks: numBlocks,
	}
}

// EstimatePriorityCmd defines the estimatePriority JSON-RPC command.
type EstimatePriorityCmd struct {
	NumBlocks int64
}

// NewEstimatePriorityCmd returns a new instance which can be used to issue a
// estimatePriority JSON-RPC command.
func NewEstimatePriorityCmd(numBlocks int64) *EstimatePriorityCmd {
	return &EstimatePriorityCmd{
		NumBlocks: numBlocks,
	}
}

// GetAccountCmd defines the getAccount JSON-RPC command.
type GetAccountCmd struct {
	Address string
}

// NewGetAccountCmd returns a new instance which can be used to issue a
// getAccount JSON-RPC command.
func NewGetAccountCmd(address string) *GetAccountCmd {
	return &GetAccountCmd{
		Address: address,
	}
}

// GetAccountAddressCmd defines the getAccountAddress JSON-RPC command.
type GetAccountAddressCmd struct {
	Account string
}

// NewGetAccountAddressCmd returns a new instance which can be used to issue a
// getAccountAddress JSON-RPC command.
func NewGetAccountAddressCmd(account string) *GetAccountAddressCmd {
	return &GetAccountAddressCmd{
		Account: account,
	}
}

// GetAddressesByAccountCmd defines the getAddressesByAccount JSON-RPC command.
type GetAddressesByAccountCmd struct {
	Account string
}

// NewGetAddressesByAccountCmd returns a new instance which can be used to issue
// a getAddressesByAccount JSON-RPC command.
func NewGetAddressesByAccountCmd(account string) *GetAddressesByAccountCmd {
	return &GetAddressesByAccountCmd{
		Account: account,
	}
}

// GetBalanceCmd defines the getBalance JSON-RPC command.
type GetBalanceCmd struct {
	Account *string
	MinConf *int `jsonrpcdefault:"1"`
}

// NewGetBalanceCmd returns a new instance which can be used to issue a
// getBalance JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewGetBalanceCmd(account *string, minConf *int) *GetBalanceCmd {
	return &GetBalanceCmd{
		Account: account,
		MinConf: minConf,
	}
}

// GetNewAddressCmd defines the getNewAddress JSON-RPC command.
type GetNewAddressCmd struct {
	Account *string
}

// NewGetNewAddressCmd returns a new instance which can be used to issue a
// getNewAddress JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewGetNewAddressCmd(account *string) *GetNewAddressCmd {
	return &GetNewAddressCmd{
		Account: account,
	}
}

// GetRawChangeAddressCmd defines the getRawChangeAddress JSON-RPC command.
type GetRawChangeAddressCmd struct {
	Account *string
}

// NewGetRawChangeAddressCmd returns a new instance which can be used to issue a
// getRawChangeAddress JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewGetRawChangeAddressCmd(account *string) *GetRawChangeAddressCmd {
	return &GetRawChangeAddressCmd{
		Account: account,
	}
}

// GetReceivedByAccountCmd defines the getReceivedByAccount JSON-RPC command.
type GetReceivedByAccountCmd struct {
	Account string
	MinConf *int `jsonrpcdefault:"1"`
}

// NewGetReceivedByAccountCmd returns a new instance which can be used to issue
// a getReceivedByAccount JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewGetReceivedByAccountCmd(account string, minConf *int) *GetReceivedByAccountCmd {
	return &GetReceivedByAccountCmd{
		Account: account,
		MinConf: minConf,
	}
}

// GetReceivedByAddressCmd defines the getReceivedByAddress JSON-RPC command.
type GetReceivedByAddressCmd struct {
	Address string
	MinConf *int `jsonrpcdefault:"1"`
}

// NewGetReceivedByAddressCmd returns a new instance which can be used to issue
// a getReceivedByAddress JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewGetReceivedByAddressCmd(address string, minConf *int) *GetReceivedByAddressCmd {
	return &GetReceivedByAddressCmd{
		Address: address,
		MinConf: minConf,
	}
}

// GetTransactionCmd defines the getTransaction JSON-RPC command.
type GetTransactionCmd struct {
	Txid             string
	IncludeWatchOnly *bool `jsonrpcdefault:"false"`
}

// NewGetTransactionCmd returns a new instance which can be used to issue a
// getTransaction JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewGetTransactionCmd(txHash string, includeWatchOnly *bool) *GetTransactionCmd {
	return &GetTransactionCmd{
		Txid:             txHash,
		IncludeWatchOnly: includeWatchOnly,
	}
}

// GetWalletInfoCmd defines the getWalletInfo JSON-RPC command.
type GetWalletInfoCmd struct{}

// NewGetWalletInfoCmd returns a new instance which can be used to issue a
// getWalletInfo JSON-RPC command.
func NewGetWalletInfoCmd() *GetWalletInfoCmd {
	return &GetWalletInfoCmd{}
}

// ImportPrivKeyCmd defines the importPrivKey JSON-RPC command.
type ImportPrivKeyCmd struct {
	PrivKey string
	Label   *string
	Rescan  *bool `jsonrpcdefault:"true"`
}

// NewImportPrivKeyCmd returns a new instance which can be used to issue a
// importPrivKey JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewImportPrivKeyCmd(privKey string, label *string, rescan *bool) *ImportPrivKeyCmd {
	return &ImportPrivKeyCmd{
		PrivKey: privKey,
		Label:   label,
		Rescan:  rescan,
	}
}

// KeyPoolRefillCmd defines the keyPoolRefill JSON-RPC command.
type KeyPoolRefillCmd struct {
	NewSize *uint `jsonrpcdefault:"100"`
}

// NewKeyPoolRefillCmd returns a new instance which can be used to issue a
// keyPoolRefill JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewKeyPoolRefillCmd(newSize *uint) *KeyPoolRefillCmd {
	return &KeyPoolRefillCmd{
		NewSize: newSize,
	}
}

// ListAccountsCmd defines the listAccounts JSON-RPC command.
type ListAccountsCmd struct {
	MinConf *int `jsonrpcdefault:"1"`
}

// NewListAccountsCmd returns a new instance which can be used to issue a
// listAccounts JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewListAccountsCmd(minConf *int) *ListAccountsCmd {
	return &ListAccountsCmd{
		MinConf: minConf,
	}
}

// ListAddressGroupingsCmd defines the listAddressGroupings JSON-RPC command.
type ListAddressGroupingsCmd struct{}

// NewListAddressGroupingsCmd returns a new instance which can be used to issue
// a listaddressgroupoings JSON-RPC command.
func NewListAddressGroupingsCmd() *ListAddressGroupingsCmd {
	return &ListAddressGroupingsCmd{}
}

// ListLockUnspentCmd defines the listLockUnspent JSON-RPC command.
type ListLockUnspentCmd struct{}

// NewListLockUnspentCmd returns a new instance which can be used to issue a
// listLockUnspent JSON-RPC command.
func NewListLockUnspentCmd() *ListLockUnspentCmd {
	return &ListLockUnspentCmd{}
}

// ListReceivedByAccountCmd defines the listReceivedByAccount JSON-RPC command.
type ListReceivedByAccountCmd struct {
	MinConf          *int  `jsonrpcdefault:"1"`
	IncludeEmpty     *bool `jsonrpcdefault:"false"`
	IncludeWatchOnly *bool `jsonrpcdefault:"false"`
}

// NewListReceivedByAccountCmd returns a new instance which can be used to issue
// a listReceivedByAccount JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewListReceivedByAccountCmd(minConf *int, includeEmpty, includeWatchOnly *bool) *ListReceivedByAccountCmd {
	return &ListReceivedByAccountCmd{
		MinConf:          minConf,
		IncludeEmpty:     includeEmpty,
		IncludeWatchOnly: includeWatchOnly,
	}
}

// ListReceivedByAddressCmd defines the listReceivedByAddress JSON-RPC command.
type ListReceivedByAddressCmd struct {
	MinConf          *int  `jsonrpcdefault:"1"`
	IncludeEmpty     *bool `jsonrpcdefault:"false"`
	IncludeWatchOnly *bool `jsonrpcdefault:"false"`
}

// NewListReceivedByAddressCmd returns a new instance which can be used to issue
// a listReceivedByAddress JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewListReceivedByAddressCmd(minConf *int, includeEmpty, includeWatchOnly *bool) *ListReceivedByAddressCmd {
	return &ListReceivedByAddressCmd{
		MinConf:          minConf,
		IncludeEmpty:     includeEmpty,
		IncludeWatchOnly: includeWatchOnly,
	}
}

// ListSinceBlockCmd defines the listSinceBlock JSON-RPC command.
type ListSinceBlockCmd struct {
	BlockHash           *string
	TargetConfirmations *int  `jsonrpcdefault:"1"`
	IncludeWatchOnly    *bool `jsonrpcdefault:"false"`
}

// NewListSinceBlockCmd returns a new instance which can be used to issue a
// listSinceBlock JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewListSinceBlockCmd(blockHash *string, targetConfirms *int, includeWatchOnly *bool) *ListSinceBlockCmd {
	return &ListSinceBlockCmd{
		BlockHash:           blockHash,
		TargetConfirmations: targetConfirms,
		IncludeWatchOnly:    includeWatchOnly,
	}
}

// ListTransactionsCmd defines the listTransactions JSON-RPC command.
type ListTransactionsCmd struct {
	Account          *string
	Count            *int  `jsonrpcdefault:"10"`
	From             *int  `jsonrpcdefault:"0"`
	IncludeWatchOnly *bool `jsonrpcdefault:"false"`
}

// NewListTransactionsCmd returns a new instance which can be used to issue a
// listTransactions JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewListTransactionsCmd(account *string, count, from *int, includeWatchOnly *bool) *ListTransactionsCmd {
	return &ListTransactionsCmd{
		Account:          account,
		Count:            count,
		From:             from,
		IncludeWatchOnly: includeWatchOnly,
	}
}

// ListUnspentCmd defines the listUnspent JSON-RPC command.
type ListUnspentCmd struct {
	MinConf   *int `jsonrpcdefault:"1"`
	MaxConf   *int `jsonrpcdefault:"9999999"`
	Addresses *[]string
}

// NewListUnspentCmd returns a new instance which can be used to issue a
// listUnspent JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewListUnspentCmd(minConf, maxConf *int, addresses *[]string) *ListUnspentCmd {
	return &ListUnspentCmd{
		MinConf:   minConf,
		MaxConf:   maxConf,
		Addresses: addresses,
	}
}

// LockUnspentCmd defines the lockUnspent JSON-RPC command.
type LockUnspentCmd struct {
	Unlock       bool
	Transactions []TransactionInput
}

// NewLockUnspentCmd returns a new instance which can be used to issue a
// lockUnspent JSON-RPC command.
func NewLockUnspentCmd(unlock bool, transactions []TransactionInput) *LockUnspentCmd {
	return &LockUnspentCmd{
		Unlock:       unlock,
		Transactions: transactions,
	}
}

// MoveCmd defines the move JSON-RPC command.
type MoveCmd struct {
	FromAccount string
	ToAccount   string
	Amount      float64 // In BTC
	MinConf     *int    `jsonrpcdefault:"1"`
	Comment     *string
}

// NewMoveCmd returns a new instance which can be used to issue a move JSON-RPC
// command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewMoveCmd(fromAccount, toAccount string, amount float64, minConf *int, comment *string) *MoveCmd {
	return &MoveCmd{
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Amount:      amount,
		MinConf:     minConf,
		Comment:     comment,
	}
}

// SendFromCmd defines the sendFrom JSON-RPC command.
type SendFromCmd struct {
	FromAccount string
	ToAddress   string
	Amount      float64 // In BTC
	MinConf     *int    `jsonrpcdefault:"1"`
	Comment     *string
	CommentTo   *string
}

// NewSendFromCmd returns a new instance which can be used to issue a sendFrom
// JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewSendFromCmd(fromAccount, toAddress string, amount float64, minConf *int, comment, commentTo *string) *SendFromCmd {
	return &SendFromCmd{
		FromAccount: fromAccount,
		ToAddress:   toAddress,
		Amount:      amount,
		MinConf:     minConf,
		Comment:     comment,
		CommentTo:   commentTo,
	}
}

// SendManyCmd defines the sendMany JSON-RPC command.
type SendManyCmd struct {
	FromAccount string
	Amounts     map[string]float64 `jsonrpcusage:"{\"address\":amount,...}"` // In BTC
	MinConf     *int               `jsonrpcdefault:"1"`
	Comment     *string
}

// NewSendManyCmd returns a new instance which can be used to issue a sendMany
// JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewSendManyCmd(fromAccount string, amounts map[string]float64, minConf *int, comment *string) *SendManyCmd {
	return &SendManyCmd{
		FromAccount: fromAccount,
		Amounts:     amounts,
		MinConf:     minConf,
		Comment:     comment,
	}
}

// SendToAddressCmd defines the sendToAddress JSON-RPC command.
type SendToAddressCmd struct {
	Address   string
	Amount    float64
	Comment   *string
	CommentTo *string
}

// NewSendToAddressCmd returns a new instance which can be used to issue a
// sendToAddress JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewSendToAddressCmd(address string, amount float64, comment, commentTo *string) *SendToAddressCmd {
	return &SendToAddressCmd{
		Address:   address,
		Amount:    amount,
		Comment:   comment,
		CommentTo: commentTo,
	}
}

// SetAccountCmd defines the setAccount JSON-RPC command.
type SetAccountCmd struct {
	Address string
	Account string
}

// NewSetAccountCmd returns a new instance which can be used to issue a
// setAccount JSON-RPC command.
func NewSetAccountCmd(address, account string) *SetAccountCmd {
	return &SetAccountCmd{
		Address: address,
		Account: account,
	}
}

// SetTxFeeCmd defines the setTxFee JSON-RPC command.
type SetTxFeeCmd struct {
	Amount float64 // In BTC
}

// NewSetTxFeeCmd returns a new instance which can be used to issue a setTxFee
// JSON-RPC command.
func NewSetTxFeeCmd(amount float64) *SetTxFeeCmd {
	return &SetTxFeeCmd{
		Amount: amount,
	}
}

// SignMessageCmd defines the signMessage JSON-RPC command.
type SignMessageCmd struct {
	Address string
	Message string
}

// NewSignMessageCmd returns a new instance which can be used to issue a
// signMessage JSON-RPC command.
func NewSignMessageCmd(address, message string) *SignMessageCmd {
	return &SignMessageCmd{
		Address: address,
		Message: message,
	}
}

// RawTxInput models the data needed for raw transaction input that is used in
// the SignRawTransactionCmd struct.
type RawTxInput struct {
	Txid         string `json:"txid"`
	Vout         uint32 `json:"vout"`
	ScriptPubKey string `json:"scriptPubKey"`
	RedeemScript string `json:"redeemScript"`
}

// SignRawTransactionCmd defines the signRawTransaction JSON-RPC command.
type SignRawTransactionCmd struct {
	RawTx    string
	Inputs   *[]RawTxInput
	PrivKeys *[]string
	Flags    *string `jsonrpcdefault:"\"ALL\""`
}

// NewSignRawTransactionCmd returns a new instance which can be used to issue a
// signRawTransaction JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewSignRawTransactionCmd(hexEncodedTx string, inputs *[]RawTxInput, privKeys *[]string, flags *string) *SignRawTransactionCmd {
	return &SignRawTransactionCmd{
		RawTx:    hexEncodedTx,
		Inputs:   inputs,
		PrivKeys: privKeys,
		Flags:    flags,
	}
}

// WalletLockCmd defines the walletLock JSON-RPC command.
type WalletLockCmd struct{}

// NewWalletLockCmd returns a new instance which can be used to issue a
// walletLock JSON-RPC command.
func NewWalletLockCmd() *WalletLockCmd {
	return &WalletLockCmd{}
}

// WalletPassphraseCmd defines the walletPassphrase JSON-RPC command.
type WalletPassphraseCmd struct {
	Passphrase string
	Timeout    int64
}

// NewWalletPassphraseCmd returns a new instance which can be used to issue a
// walletPassphrase JSON-RPC command.
func NewWalletPassphraseCmd(passphrase string, timeout int64) *WalletPassphraseCmd {
	return &WalletPassphraseCmd{
		Passphrase: passphrase,
		Timeout:    timeout,
	}
}

// WalletPassphraseChangeCmd defines the walletPassphrase JSON-RPC command.
type WalletPassphraseChangeCmd struct {
	OldPassphrase string
	NewPassphrase string
}

// NewWalletPassphraseChangeCmd returns a new instance which can be used to
// issue a walletPassphraseChange JSON-RPC command.
func NewWalletPassphraseChangeCmd(oldPassphrase, newPassphrase string) *WalletPassphraseChangeCmd {
	return &WalletPassphraseChangeCmd{
		OldPassphrase: oldPassphrase,
		NewPassphrase: newPassphrase,
	}
}

func init() {
	// The commands in this file are only usable with a wallet server.
	flags := UFWalletOnly

	MustRegisterCmd("addMultisigAddress", (*AddMultisigAddressCmd)(nil), flags)
	MustRegisterCmd("createMultisig", (*CreateMultisigCmd)(nil), flags)
	MustRegisterCmd("dumpPrivKey", (*DumpPrivKeyCmd)(nil), flags)
	MustRegisterCmd("encryptWallet", (*EncryptWalletCmd)(nil), flags)
	MustRegisterCmd("estimateFee", (*EstimateFeeCmd)(nil), flags)
	MustRegisterCmd("estimatePriority", (*EstimatePriorityCmd)(nil), flags)
	MustRegisterCmd("getAccount", (*GetAccountCmd)(nil), flags)
	MustRegisterCmd("getAccountAddress", (*GetAccountAddressCmd)(nil), flags)
	MustRegisterCmd("getAddressesByAccount", (*GetAddressesByAccountCmd)(nil), flags)
	MustRegisterCmd("getBalance", (*GetBalanceCmd)(nil), flags)
	MustRegisterCmd("getNewAddress", (*GetNewAddressCmd)(nil), flags)
	MustRegisterCmd("getRawChangeAddress", (*GetRawChangeAddressCmd)(nil), flags)
	MustRegisterCmd("getReceivedByAccount", (*GetReceivedByAccountCmd)(nil), flags)
	MustRegisterCmd("getReceivedByAddress", (*GetReceivedByAddressCmd)(nil), flags)
	MustRegisterCmd("getTransaction", (*GetTransactionCmd)(nil), flags)
	MustRegisterCmd("getWalletInfo", (*GetWalletInfoCmd)(nil), flags)
	MustRegisterCmd("importPrivKey", (*ImportPrivKeyCmd)(nil), flags)
	MustRegisterCmd("keyPoolRefill", (*KeyPoolRefillCmd)(nil), flags)
	MustRegisterCmd("listAccounts", (*ListAccountsCmd)(nil), flags)
	MustRegisterCmd("listAddressGroupings", (*ListAddressGroupingsCmd)(nil), flags)
	MustRegisterCmd("listLockUnspent", (*ListLockUnspentCmd)(nil), flags)
	MustRegisterCmd("listReceivedByAccount", (*ListReceivedByAccountCmd)(nil), flags)
	MustRegisterCmd("listReceivedByAddress", (*ListReceivedByAddressCmd)(nil), flags)
	MustRegisterCmd("listSinceBlock", (*ListSinceBlockCmd)(nil), flags)
	MustRegisterCmd("listTransactions", (*ListTransactionsCmd)(nil), flags)
	MustRegisterCmd("listUnspent", (*ListUnspentCmd)(nil), flags)
	MustRegisterCmd("lockUnspent", (*LockUnspentCmd)(nil), flags)
	MustRegisterCmd("move", (*MoveCmd)(nil), flags)
	MustRegisterCmd("sendFrom", (*SendFromCmd)(nil), flags)
	MustRegisterCmd("sendMany", (*SendManyCmd)(nil), flags)
	MustRegisterCmd("sendToAddress", (*SendToAddressCmd)(nil), flags)
	MustRegisterCmd("setAccount", (*SetAccountCmd)(nil), flags)
	MustRegisterCmd("setTxFee", (*SetTxFeeCmd)(nil), flags)
	MustRegisterCmd("signMessage", (*SignMessageCmd)(nil), flags)
	MustRegisterCmd("signRawTransaction", (*SignRawTransactionCmd)(nil), flags)
	MustRegisterCmd("walletLock", (*WalletLockCmd)(nil), flags)
	MustRegisterCmd("walletPassphrase", (*WalletPassphraseCmd)(nil), flags)
	MustRegisterCmd("walletPassphraseChange", (*WalletPassphraseChangeCmd)(nil), flags)
}
