package github

import "errors"

var ErrInvalidOptions = errors.New("invalid parameter, please check your request")

var NotFound = errors.New("not found")
