package utils

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
)

// FormatKas takes the amount of sompis as uint64, and returns amount of KAS with 8 decimal places
func FormatKas(amount uint64) string {
	res := "                   "
	if amount > 0 {
		res = fmt.Sprintf("%19.8f", float64(amount)/constants.SompiPerKaspa)
	}
	return res
}

// FormatUtxos takes the number of UTXOs as uint64, and returns a string
// of 8 places that is enough to express an integer up to 10^8-1
func FormatUtxos(amount uint64) string {
	res := "        "
	if amount > 0 {
		res = fmt.Sprintf("%8d", amount)
	}
	return res
}
