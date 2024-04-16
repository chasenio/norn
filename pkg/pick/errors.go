package pick

import "errors"

var (
	ErrInvalidOptions = errors.New("invalid parameter, please check your request")
	ErrConflict       = errors.New("conflict")
)
