// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
)

// ScriptClass is an enumeration for the list of standard types of script.
type ScriptClass byte

// Classes of script payment known about in the blockDAG.
const (
	NonStandardTy ScriptClass = iota // None of the recognized forms.
	PubKeyHashTy                     // Pay pubkey hash.
	ScriptHashTy                     // Pay to script hash.
)

// scriptClassToName houses the human-readable strings which describe each
// script class.
var scriptClassToName = []string{
	NonStandardTy: "nonstandard",
	PubKeyHashTy:  "pubkeyhash",
	ScriptHashTy:  "scripthash",
}

// String implements the Stringer interface by returning the name of
// the enum script class. If the enum is invalid then "Invalid" will be
// returned.
func (t ScriptClass) String() string {
	if int(t) > len(scriptClassToName) || int(t) < 0 {
		return "Invalid"
	}
	return scriptClassToName[t]
}

// isPubkeyHash returns true if the script passed is a pay-to-pubkey-hash
// transaction, false otherwise.
func isPubkeyHash(pops []parsedOpcode) bool {
	return len(pops) == 5 &&
		pops[0].opcode.value == OpDup &&
		pops[1].opcode.value == OpHash160 &&
		pops[2].opcode.value == OpData20 &&
		pops[3].opcode.value == OpEqualVerify &&
		pops[4].opcode.value == OpCheckSig

}

// scriptType returns the type of the script being inspected from the known
// standard types.
func typeOfScript(pops []parsedOpcode) ScriptClass {
	if isPubkeyHash(pops) {
		return PubKeyHashTy
	} else if isScriptHash(pops) {
		return ScriptHashTy
	}
	return NonStandardTy
}

// GetScriptClass returns the class of the script passed.
//
// NonStandardTy will be returned when the script does not parse.
func GetScriptClass(script []byte) ScriptClass {
	pops, err := parseScript(script)
	if err != nil {
		return NonStandardTy
	}
	return typeOfScript(pops)
}

// expectedInputs returns the number of arguments required by a script.
// If the script is of unknown type such that the number can not be determined
// then -1 is returned. We are an internal function and thus assume that class
// is the real class of pops (and we can thus assume things that were determined
// while finding out the type).
func expectedInputs(pops []parsedOpcode, class ScriptClass) int {
	switch class {

	case PubKeyHashTy:
		return 2

	case ScriptHashTy:
		// Not including script. That is handled by the caller.
		return 1

	default:
		return -1
	}
}

// ScriptInfo houses information about a script pair that is determined by
// CalcScriptInfo.
type ScriptInfo struct {
	// ScriptPubKeyClass is the class of the public key script and is equivalent
	// to calling GetScriptClass on it.
	ScriptPubKeyClass ScriptClass

	// NumInputs is the number of inputs provided by the public key script.
	NumInputs int

	// ExpectedInputs is the number of outputs required by the signature
	// script and any pay-to-script-hash scripts. The number will be -1 if
	// unknown.
	ExpectedInputs int

	// SigOps is the number of signature operations in the script pair.
	SigOps int
}

// CalcScriptInfo returns a structure providing data about the provided script
// pair. It will error if the pair is in someway invalid such that they can not
// be analysed, i.e. if they do not parse or the scriptPubKey is not a push-only
// script
func CalcScriptInfo(sigScript, scriptPubKey []byte, isP2SH bool) (*ScriptInfo, error) {
	sigPops, err := parseScript(sigScript)
	if err != nil {
		return nil, err
	}

	scriptPubKeyPops, err := parseScript(scriptPubKey)
	if err != nil {
		return nil, err
	}

	// Push only sigScript makes little sense.
	si := new(ScriptInfo)
	si.ScriptPubKeyClass = typeOfScript(scriptPubKeyPops)

	// Can't have a signature script that doesn't just push data.
	if !isPushOnly(sigPops) {
		return nil, scriptError(ErrNotPushOnly,
			"signature script is not push only")
	}

	si.ExpectedInputs = expectedInputs(scriptPubKeyPops, si.ScriptPubKeyClass)

	// All entries pushed to stack (or are OP_RESERVED and exec will fail).
	si.NumInputs = len(sigPops)

	if si.ScriptPubKeyClass == ScriptHashTy && isP2SH {
		// The pay-to-hash-script is the final data push of the
		// signature script.
		script := sigPops[len(sigPops)-1].data
		shPops, err := parseScript(script)
		if err != nil {
			return nil, err
		}

		shInputs := expectedInputs(shPops, typeOfScript(shPops))
		if shInputs == -1 {
			si.ExpectedInputs = -1
		} else {
			si.ExpectedInputs += shInputs
		}
		si.SigOps = getSigOpCount(shPops, true)
	} else {
		si.SigOps = getSigOpCount(scriptPubKeyPops, true)
	}

	return si, nil
}

// payToPubKeyHashScript creates a new script to pay a transaction
// output to a 20-byte pubkey hash. It is expected that the input is a valid
// hash.
func payToPubKeyHashScript(pubKeyHash []byte) ([]byte, error) {
	return NewScriptBuilder().AddOp(OpDup).AddOp(OpHash160).
		AddData(pubKeyHash).AddOp(OpEqualVerify).AddOp(OpCheckSig).
		Script()
}

// payToScriptHashScript creates a new script to pay a transaction output to a
// script hash. It is expected that the input is a valid hash.
func payToScriptHashScript(scriptHash []byte) ([]byte, error) {
	return NewScriptBuilder().AddOp(OpHash160).AddData(scriptHash).
		AddOp(OpEqual).Script()
}

// PayToAddrScript creates a new script to pay a transaction output to a the
// specified address.
func PayToAddrScript(addr util.Address) (*externalapi.ScriptPublicKey, error) {
	const nilAddrErrStr = "unable to generate payment script for nil address"
	switch addr := addr.(type) {
	case *util.AddressPubKeyHash:
		if addr == nil {
			return nil, scriptError(ErrUnsupportedAddress,
				nilAddrErrStr)
		}
		script, err := payToPubKeyHashScript(addr.ScriptAddress())
		if err != nil {
			return nil, err
		}
		return &externalapi.ScriptPublicKey{script, constants.ScriptPublicKeyVersion}, err

	case *util.AddressScriptHash:
		if addr == nil {
			return nil, scriptError(ErrUnsupportedAddress,
				nilAddrErrStr)
		}
		script, err := payToScriptHashScript(addr.ScriptAddress())
		if err != nil {
			return nil, err
		}
		return &externalapi.ScriptPublicKey{script, constants.ScriptPublicKeyVersion}, err
	}

	str := fmt.Sprintf("unable to generate payment script for unsupported "+
		"address type %T", addr)
	return nil, scriptError(ErrUnsupportedAddress, str)
}

// PayToScriptHashScript takes a script and returns an equivalent pay-to-script-hash script
func PayToScriptHashScript(redeemScript []byte) ([]byte, error) {
	redeemScriptHash := util.Hash160(redeemScript)
	script, err := NewScriptBuilder().
		AddOp(OpHash160).AddData(redeemScriptHash).
		AddOp(OpEqual).Script()
	if err != nil {
		return nil, err
	}
	return script, nil
}

// PayToScriptHashSignatureScript generates a signature script that fits a pay-to-script-hash script
func PayToScriptHashSignatureScript(redeemScript []byte, signature []byte) ([]byte, error) {
	redeemScriptAsData, err := NewScriptBuilder().AddData(redeemScript).Script()
	if err != nil {
		return nil, err
	}
	signatureScript := make([]byte, len(signature)+len(redeemScriptAsData))
	copy(signatureScript, signature)
	copy(signatureScript[len(signature):], redeemScriptAsData)
	return signatureScript, nil
}

// PushedData returns an array of byte slices containing any pushed data found
// in the passed script. This includes OP_0, but not OP_1 - OP_16.
func PushedData(script []byte) ([][]byte, error) {
	pops, err := parseScript(script)
	if err != nil {
		return nil, err
	}

	var data [][]byte
	for _, pop := range pops {
		if pop.data != nil {
			data = append(data, pop.data)
		} else if pop.opcode.value == Op0 {
			data = append(data, nil)
		}
	}
	return data, nil
}

// ExtractScriptPubKeyAddress returns the type of script and its addresses.
// Note that it only works for 'standard' transaction script types. Any data such
// as public keys which are invalid will return a nil address.
func ExtractScriptPubKeyAddress(scriptPubKey *externalapi.ScriptPublicKey, dagParams *dagconfig.Params) (ScriptClass, util.Address, error) {
	if scriptPubKey.Version > constants.ScriptPublicKeyVersion {
		return NonStandardTy, nil, errors.Errorf("Script version is unkown.")
	}
	// No valid address if the script doesn't parse.
	pops, err := parseScript(scriptPubKey.Script)
	if err != nil {
		return NonStandardTy, nil, err
	}

	scriptClass := typeOfScript(pops)
	switch scriptClass {
	case PubKeyHashTy:
		// A pay-to-pubkey-hash script is of the form:
		//  OP_DUP OP_HASH160 <hash> OP_EQUALVERIFY OP_CHECKSIG
		// Therefore the pubkey hash is the 3rd item on the stack.
		// If the pubkey hash is invalid for some reason, return a nil address.
		addr, err := util.NewAddressPubKeyHash(pops[2].data,
			dagParams.Prefix)
		if err != nil {
			return scriptClass, nil, nil
		}
		return scriptClass, addr, nil

	case ScriptHashTy:
		// A pay-to-script-hash script is of the form:
		//  OP_HASH160 <scripthash> OP_EQUAL
		// Therefore the script hash is the 2nd item on the stack.
		// If the script hash ss invalid for some reason, return a nil address.
		addr, err := util.NewAddressScriptHashFromHash(pops[1].data,
			dagParams.Prefix)
		if err != nil {
			return scriptClass, nil, nil
		}
		return scriptClass, addr, nil

	case NonStandardTy:
		// Don't attempt to extract addresses or required signatures for
		// nonstandard transactions.
		return NonStandardTy, nil, nil
	}

	return NonStandardTy, nil, errors.Errorf("Cannot handle script class %s", scriptClass)
}

// AtomicSwapDataPushes houses the data pushes found in atomic swap contracts.
type AtomicSwapDataPushes struct {
	RecipientHash160 [20]byte
	RefundHash160    [20]byte
	SecretHash       [32]byte
	SecretSize       int64
	LockTime         uint64
}

// ExtractAtomicSwapDataPushes returns the data pushes from an atomic swap
// contract. If the script is not an atomic swap contract,
// ExtractAtomicSwapDataPushes returns (nil, nil). Non-nil errors are returned
// for unparsable scripts.
//
// NOTE: Atomic swaps are not considered standard script types by the dcrd
// mempool policy and should be used with P2SH. The atomic swap format is also
// expected to change to use a more secure hash function in the future.
//
// This function is only defined in the txscript package due to API limitations
// which prevent callers using txscript to parse nonstandard scripts.
func ExtractAtomicSwapDataPushes(version uint16, scriptPubKey []byte) (*AtomicSwapDataPushes, error) {
	pops, err := parseScript(scriptPubKey)
	if err != nil {
		return nil, err
	}

	if len(pops) != 20 {
		return nil, nil
	}
	isAtomicSwap := pops[0].opcode.value == OpIf &&
		pops[1].opcode.value == OpSize &&
		canonicalPush(pops[2]) &&
		pops[3].opcode.value == OpEqualVerify &&
		pops[4].opcode.value == OpSHA256 &&
		pops[5].opcode.value == OpData32 &&
		pops[6].opcode.value == OpEqualVerify &&
		pops[7].opcode.value == OpDup &&
		pops[8].opcode.value == OpHash160 &&
		pops[9].opcode.value == OpData20 &&
		pops[10].opcode.value == OpElse &&
		canonicalPush(pops[11]) &&
		pops[12].opcode.value == OpCheckLockTimeVerify &&
		pops[13].opcode.value == OpDrop &&
		pops[14].opcode.value == OpDup &&
		pops[15].opcode.value == OpHash160 &&
		pops[16].opcode.value == OpData20 &&
		pops[17].opcode.value == OpEndIf &&
		pops[18].opcode.value == OpEqualVerify &&
		pops[19].opcode.value == OpCheckSig
	if !isAtomicSwap {
		return nil, nil
	}

	pushes := new(AtomicSwapDataPushes)
	copy(pushes.SecretHash[:], pops[5].data)
	copy(pushes.RecipientHash160[:], pops[9].data)
	copy(pushes.RefundHash160[:], pops[16].data)
	if pops[2].data != nil {
		locktime, err := makeScriptNum(pops[2].data, 5)
		if err != nil {
			return nil, nil
		}
		pushes.SecretSize = int64(locktime)
	} else if op := pops[2].opcode; isSmallInt(op) {
		pushes.SecretSize = int64(asSmallInt(op))
	} else {
		return nil, nil
	}
	if pops[11].data != nil {
		locktime, err := makeScriptNum(pops[11].data, 5)
		if err != nil {
			return nil, nil
		}
		pushes.LockTime = uint64(locktime)
	} else if op := pops[11].opcode; isSmallInt(op) {
		pushes.LockTime = uint64(asSmallInt(op))
	} else {
		return nil, nil
	}
	return pushes, nil
}
