// Copyright (c) 2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC commands that are supported by
// a wallet server with btcwallet extensions.

package btcjson

// CreateNewAccountCmd defines the createNewAccount JSON-RPC command.
type CreateNewAccountCmd struct {
	Account string
}

// NewCreateNewAccountCmd returns a new instance which can be used to issue a
// createNewAccount JSON-RPC command.
func NewCreateNewAccountCmd(account string) *CreateNewAccountCmd {
	return &CreateNewAccountCmd{
		Account: account,
	}
}

// DumpWalletCmd defines the dumpWallet JSON-RPC command.
type DumpWalletCmd struct {
	Filename string
}

// NewDumpWalletCmd returns a new instance which can be used to issue a
// dumpWallet JSON-RPC command.
func NewDumpWalletCmd(filename string) *DumpWalletCmd {
	return &DumpWalletCmd{
		Filename: filename,
	}
}

// ImportAddressCmd defines the importAddress JSON-RPC command.
type ImportAddressCmd struct {
	Address string
	Rescan  *bool `jsonrpcdefault:"true"`
}

// NewImportAddressCmd returns a new instance which can be used to issue an
// importAddress JSON-RPC command.
func NewImportAddressCmd(address string, rescan *bool) *ImportAddressCmd {
	return &ImportAddressCmd{
		Address: address,
		Rescan:  rescan,
	}
}

// ImportPubKeyCmd defines the importPubKey JSON-RPC command.
type ImportPubKeyCmd struct {
	PubKey string
	Rescan *bool `jsonrpcdefault:"true"`
}

// NewImportPubKeyCmd returns a new instance which can be used to issue an
// importPubKey JSON-RPC command.
func NewImportPubKeyCmd(pubKey string, rescan *bool) *ImportPubKeyCmd {
	return &ImportPubKeyCmd{
		PubKey: pubKey,
		Rescan: rescan,
	}
}

// ImportWalletCmd defines the importWallet JSON-RPC command.
type ImportWalletCmd struct {
	Filename string
}

// NewImportWalletCmd returns a new instance which can be used to issue a
// importWallet JSON-RPC command.
func NewImportWalletCmd(filename string) *ImportWalletCmd {
	return &ImportWalletCmd{
		Filename: filename,
	}
}

// RenameAccountCmd defines the renameAccount JSON-RPC command.
type RenameAccountCmd struct {
	OldAccount string
	NewAccount string
}

// NewRenameAccountCmd returns a new instance which can be used to issue a
// renameAccount JSON-RPC command.
func NewRenameAccountCmd(oldAccount, newAccount string) *RenameAccountCmd {
	return &RenameAccountCmd{
		OldAccount: oldAccount,
		NewAccount: newAccount,
	}
}

func init() {
	// The commands in this file are only usable with a wallet server.
	flags := UFWalletOnly

	MustRegisterCmd("createNewAccount", (*CreateNewAccountCmd)(nil), flags)
	MustRegisterCmd("dumpWallet", (*DumpWalletCmd)(nil), flags)
	MustRegisterCmd("importAddress", (*ImportAddressCmd)(nil), flags)
	MustRegisterCmd("importPubKey", (*ImportPubKeyCmd)(nil), flags)
	MustRegisterCmd("importWallet", (*ImportWalletCmd)(nil), flags)
	MustRegisterCmd("renameAccount", (*RenameAccountCmd)(nil), flags)
}
