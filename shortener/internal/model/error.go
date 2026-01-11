package model

import (
	"fmt"

	"github.com/pkg/errors"
)

type ErrorKind string

const (
	Validation ErrorKind = "Validation Error"
	Type       ErrorKind = "Type Error"
	NotFound   ErrorKind = "Not Found"
	Unknown    ErrorKind = "Unknown Error"
	Internal   ErrorKind = "Internal Server Error"
)

// NewError return wrapped dynamic errors
func NewError(kind ErrorKind, msg string) error {
	return errors.New(fmt.Sprintf("%s: %s", string(kind), msg))
}
