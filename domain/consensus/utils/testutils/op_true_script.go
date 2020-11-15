package testutils

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
)

// OpTrueScript returns a P2SH script paying to an anyone-can-spend address
func OpTrueScript() []byte {
	opTrueScript, err := txscript.PayToScriptHashScript([]byte{txscript.OpTrue})
	if err != nil {
		panic(errors.Wrapf(err, "Couldn't parse opTrueScript. This should never happen"))
	}
	return opTrueScript
}
