package putback

type BackBuffer struct {
	Bytes   []byte // May be nil
	Pointer int
	Pool    BufferPool // May be nil
}

func (b *BackBuffer) Wipe() {
	if b == nil {
		return
	}
	b.Pointer = 0
	if b.Pool != nil {
		bytes := b.Bytes
		b.Bytes = nil
		b.Pool.PutBuffer(bytes)
	}
}

func (b *BackBuffer) BytesLeft() int {
	if b == nil || b.Bytes == nil {
		return 0
	}
	if b.Pointer < 0 {
		// defensive — should not happen, but guard anyway
		b.Pointer = 0
	}
	if b.Pointer > len(b.Bytes) {
		// also defensive — clamp
		b.Pointer = len(b.Bytes)
	}
	return len(b.Bytes) - b.Pointer
}

func (b *BackBuffer) PutBack(bytes []byte) {
	if len(bytes) == 0 {
		return
	}

	// If no existing buffer, allocate fresh and place data at start.
	if b.Bytes == nil {
		b.Pointer = 0
		if b.Pool != nil {
			b.Bytes = b.Pool.GetBuffer(len(bytes))
			copy(b.Bytes, bytes)
			return
		}
		b.Bytes = append([]byte(nil), bytes...)
		return
	}

	free := b.Pointer // bytes available on the left side
	if free >= len(bytes) {
		// We can fit the new bytes into existing left free space.
		start := b.Pointer - len(bytes)
		copy(b.Bytes[start:b.Pointer], bytes)
		b.Pointer = start
		return
	}

	// Not enough free space: allocate a new buffer containing
	// (len(bytes) + existing data)
	existing := b.Bytes[b.Pointer:]
	newLen := len(bytes) + len(existing)
	var newBuf []byte
	if b.Pool != nil {
		newBuf = b.Pool.GetBuffer(newLen)
	} else {
		newBuf = make([]byte, newLen)
	}
	// layout: [bytes... | existing...]
	copy(newBuf[0:len(bytes)], bytes)
	copy(newBuf[len(bytes):], existing)

	// return old backing buffer to pool if present
	if b.Pool != nil {
		// give the whole slice back
		b.Pool.PutBuffer(b.Bytes)
	}

	b.Bytes = newBuf
	b.Pointer = 0
}

// Never returns error
func (b *BackBuffer) Read(p []byte) (n int, err error) {
	if b == nil || b.Bytes == nil || len(b.Bytes) == 0 {
		return 0, nil
	}

	if b.Pointer >= len(b.Bytes) {
		b.Wipe()
		return 0, nil
	}

	avail := b.BytesLeft()
	if avail == 0 {
		return 0, nil
	}

	toCopy := min(len(p), avail)

	copy(p[:toCopy], b.Bytes[b.Pointer:b.Pointer+toCopy])
	b.Pointer += toCopy
	if b.Pointer >= len(b.Bytes) {
		// consumed all data -> free backing buffer if possible
		if b.Pool != nil {
			b.Pool.PutBuffer(b.Bytes)
		}
		b.Bytes = nil
		b.Pointer = 0
	}
	return toCopy, nil
}

type Packet[T any] struct {
	Buffer []byte
	Assoc  T
}

type BackPacketBuffer[T any] struct {
	Packets []Packet[T] // May be nil
	Pool    BufferPool  // May be nil
}

func (b *BackPacketBuffer[T]) Wipe() {
	if b == nil {
		return
	}
	if b.Pool != nil {
		for _, packet := range b.Packets {
			b.Pool.PutBuffer(packet.Buffer)
		}
	}
	b.Packets = nil
}

func (b *BackPacketBuffer[T]) PacketsLeft() int {
	if b == nil {
		return 0
	}
	return len(b.Packets)
}

func (b *BackPacketBuffer[T]) PutBack(bytes []byte, Assoc T) {
	if b == nil {
		return
	}
	b.Packets = append(b.Packets, Packet[T]{
		Buffer: bytes,
		Assoc:  Assoc,
	})
}

// Never returns error
func (b *BackPacketBuffer[T]) ReadFrom(p []byte) (n int, assoc T, err error) {
	if b == nil {
		return
	}
	if len(b.Packets) == 0 {
		return
	}
	packet := b.Packets[len(b.Packets)-1]
	b.Packets[len(b.Packets)-1] = Packet[T]{}
	b.Packets = b.Packets[:len(b.Packets)-1]
	n = copy(p, packet.Buffer)
	assoc = packet.Assoc
	if b.Pool != nil {
		b.Pool.PutBuffer(packet.Buffer)
	}
	return
}
