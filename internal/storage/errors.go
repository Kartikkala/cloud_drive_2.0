package storage

import (
	"errors"
)

var (
	ErrNodeNotFound   = errors.New("node not found")
	ErrNodeIsDirectory = errors.New("node is a directory")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrNoObjectData   = errors.New("node has no object data")
)
