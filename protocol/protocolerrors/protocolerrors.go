package protocolerrors

import "github.com/pkg/errors"

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

func Errorf(shouldBan bool, format string, args ...interface{}) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.Errorf(format, args...),
	}
}

func New(shouldBan bool, message string) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.New(message),
	}
}

func Wrap(shouldBan bool, err error, message string) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.Wrap(err, message),
	}
}

func Wrapf(shouldBan bool, err error, format string, args ...interface{}) error {
	return &ProtocolError{
		ShouldBan: shouldBan,
		Cause:     errors.Wrapf(err, format, args...),
	}
}
