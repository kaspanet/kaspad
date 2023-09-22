package utils

import (
	"regexp"

	"github.com/pkg/errors"
)

/**
 * 1. May be an integer (no decimal components)
 * 2. May be float with up to 8 decimal places
 */
func ValidateAmountFormat(amount string) error {
	// Check whether it's an integer, or a float with max 8 digits
	match, err := regexp.MatchString("^([1-9]\\d{0,11}|0)(\\.\\d{0,8})?$", amount)

	if !match {
		return errors.Errorf("Invalid send amount")
	}

	if err != nil {
		return err
	}

	return nil
}
