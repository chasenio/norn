package types

import "errors"

var (
	ErrInvalidOptions = errors.New("invalid parameter, please check your request")
	ErrConflict       = errors.New("conflict")

	NotFound = errors.New("not found")

	ErrUnknownProvider = errors.New("unknown provider")
)
