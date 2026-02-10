package events

import (
	"sync"

	"github.com/google/uuid"
)

type Broker[T any] struct{
	events               map[string][]chan T
	perChannelBufferSize uint8
	lock                 sync.RWMutex
}

type Job struct {
	NodeID uuid.UUID
}