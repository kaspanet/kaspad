package protocolerrors

import "github.com/pkg/errors"

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
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(shouldBan bool, format string, args ...interface{}) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.Errorf(format, args...),
	}
}

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(shouldBan bool, message string) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.New(message),
	}
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
func Wrap(shouldBan bool, err error, message string) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.Wrap(err, message),
	}
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is called, and the format specifier.
func Wrapf(shouldBan bool, err error, format string, args ...interface{}) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.Wrapf(err, format, args...),
	}
}
