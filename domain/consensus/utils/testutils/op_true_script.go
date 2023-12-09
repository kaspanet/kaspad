package testutils

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/constants"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/txscript"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/constants"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/txscript"
)

// OpTrueScript returns a P2SH script paying to an anyone-can-spend address,
// The second return value is a redeemScript to be used with txscript.PayToScriptHashSignatureScript
func OpTrueScript() (*externalapi.ScriptPublicKey, []byte) {
	var err error
	redeemScript := []byte{txscript.OpTrue}
	scriptPublicKeyScript, err := txscript.PayToScriptHashScript(redeemScript)
	if err != nil {
		panic(errors.Wrapf(err, "Couldn't parse opTrueScript. This should never happen"))
	}
	scriptPublicKey := &externalapi.ScriptPublicKey{Script: scriptPublicKeyScript, Version: constants.MaxScriptPublicKeyVersion}
	return scriptPublicKey, redeemScript
}
