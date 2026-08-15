package entity

import "errors"

var ErrEmptyKey = errors.New("rate limiter key cannot be empty")
