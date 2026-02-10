package events

func NewService[T any](perChannelBufferSize uint8) *Broker[T] {
	return &Broker[T]{
		events:               make(map[string][]chan T),
		perChannelBufferSize: perChannelBufferSize,
	}
}

func (b *Broker[T]) Subscribe(
	Event string,
) <-chan T {
	b.lock.Lock()

	defer b.lock.Unlock()

	ch := make(chan T, b.perChannelBufferSize)
	b.events[Event] = append(b.events[Event], ch)
	return ch
}

func (b *Broker[T]) Publish(
	Event string,
	Payload T,
) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if channels, found := b.events[Event]; found {
		for _, channel := range channels {
			select {
            case channel <- Payload:
                // Successfully sent
            default:
                // Channel buffer is full, decide what to do:
                // Option 1: Skip this subscriber (shown here)
                // Option 2: Log a warning
                // Option 3: Use a goroutine (see below)
            }
		}
	}
}
