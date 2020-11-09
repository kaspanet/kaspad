package protocolerrors

import (
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

// ProtocolError is an error that signifies a violation
// of the peer-to-peer protocol
type ProtocolError struct {
	ShouldBan bool
	Cause     error
}

func (e *ProtocolError) Error() string {
	return e.Cause.Error()
}

func (e *ProtocolError) Unwrap() error {
	return e.Cause
}

// Errorf formats according to a format specifier and returns the string
// as a ProtocolError.
func Errorf(shouldBan bool, format string, args ...interface{}) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.Errorf(format, args...),
	}
}

// New returns a ProtocolError with the supplied message.
// New also records the stack trace at the point it was called.
func New(shouldBan bool, message string) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.New(message),
	}
}

// Wrap wraps the given error and returns it as a ProtocolError.
func Wrap(shouldBan bool, err error, message string) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.Wrap(err, message),
	}
}

// Wrapf wraps the given error with the given format and returns it as a ProtocolError.
func Wrapf(shouldBan bool, err error, format string, args ...interface{}) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.Wrapf(err, format, args...),
	}
}

func ConvertToProtocolErrorIfRuleError(err error, format string, args ...interface{}) error {
	if !errors.As(err, &ruleerrors.RuleError{}) {
		return err
	}

	return Wrapf(true, err, format, args...)
}
