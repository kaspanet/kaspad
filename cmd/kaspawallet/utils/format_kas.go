package utils

import (
	"fmt"

	"github.com/zoomy-network/zoomyd/domain/consensus/utils/constants"
)

// FormatKas takes the amount of sompis as uint64, and returns amount of KAS with 8  decimal places
func FormatKas(amount uint64) string {
	res := "                   "
	if amount > 0 {
		res = fmt.Sprintf("%19.8f", float64(amount)/constants.SompiPerKaspa)
	}
	return res
}
