package testutils

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
)

// OpTrueScript returns a P2SH script paying to an anyone-can-spend address,
// The second return value is a redeemScript to be used with txscript.PayToScriptHashSignatureScript
func OpTrueScript() (scriptPublicKey, redeemScript []byte) {
	var err error
	redeemScript = []byte{txscript.OpTrue}
	scriptPublicKey, err = txscript.PayToScriptHashScript(redeemScript)
	if err != nil {
		panic(errors.Wrapf(err, "Couldn't parse opTrueScript. This should never happen"))
	}
	return scriptPublicKey, redeemScript
}
