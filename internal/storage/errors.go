package storage

import (
	"errors"
)

var (
	ErrNodeNotFound   = errors.New("node not found")
	ErrParentNodeNotFound   = errors.New("parent node not found")
	ErrNodeIsDirectory = errors.New("node is a directory")
	ErrNodeIsFile = errors.New("node is a file")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrNoObjectData   = errors.New("node has no object data")
)
