package tss

import "fmt"

// ErrorKind represents a specific kind of TSS error.
type ErrorKind string

// Error returns the string representation of the ErrorKind.
func (e ErrorKind) Error() string {
	return string(e)
}

// Predefined error kinds
var (
	ErrParseError            = ErrorKind("parse error")
	ErrInvalidLength         = ErrorKind("invalid length")
	ErrInvalidOrder          = ErrorKind("invalid order")
	ErrGenerateKeyPairFailed = ErrorKind("generate key pair failed")
	ErrPrivateKeyZero        = ErrorKind("private key zero")
	ErrNotOnCurve            = ErrorKind("not on curve")
	ErrInvalidSecretShare    = ErrorKind("invalid secret share")
	ErrValidSecretShare      = ErrorKind("valid secret share")
	ErrInvalidSignature      = ErrorKind("invalid signature")
	ErrNotInOrder            = ErrorKind("not in group order")
	ErrInvalidPubkeyFormat   = ErrorKind("invalid pubkey format")
	ErrRandomError           = ErrorKind("random error")
)

// Error represents a TSS error.
type Error struct {
	err  error
	desc string
}

// NewError creates a new TSS error with the given wrapped error and description.
func NewError(err error, format string, args ...interface{}) *Error {
	return &Error{
		err:  err,
		desc: fmt.Sprintf(format, args...),
	}
}

// Unwrap returns the underlying wrapped error.
func (e *Error) Unwrap() error {
	return e.err
}

// Error returns the string representation of the error.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.desc, e.err)
}
