package putback

// Static type assertion
var (
	_ WithBackBuffer            = &BackBuffer{}
	_ WithBackPacketBuffer[any] = &BackPacketBuffer[any]{}
)

// BackBuffer holds a byte slice that may be read from and to which bytes can
// be "put back" (prepended) so they will be returned by subsequent reads.
type BackBuffer struct {
	Bytes   []byte // May be nil
	Pointer int
	Pool    BufferPool // May be nil
}

// BackBuffer returns the receiver to satisfy the WithBackBuffer interface.
func (b *BackBuffer) BackBuffer() *BackBuffer {
	return b
}

// Wipe clears the buffer state. If a BufferPool is present, the backing
// slice is returned to the pool. Wipe is safe to call on a nil receiver.
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

// BytesLeft returns the number of unread bytes remaining in the buffer. If
// the receiver or its backing slice is nil, BytesLeft returns 0. The method
// defensively clamps Pointer to the valid range.
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

// PutBack prepends the provided bytes so they will be returned by subsequent
// Read calls. If there is sufficient unused space to the left of the current
// Pointer the bytes are copied into that space. Otherwise a new backing
// buffer is allocated (or obtained from Pool) and the existing unread bytes
// are appended after the new data. If the pool is used the old backing slice
// is returned to the pool.
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

// Read reads up to len(p) bytes from the unread portion of the buffer into p
// and advances the read position. Read never returns an error
// and is safe to call on a nil receiver. When all data is consumed
// the backing buffer is released to the Pool if present.
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

// Packet holds a single buffer and an associated value of type T. It is used
// by BackPacketBuffer to store packets that can be put back and read later.
type Packet[T any] struct {
	Buffer []byte
	Assoc  T
}

// BackPacketBuffer stores a stack of packets that can be pushed back and later
// read in LIFO order. If a BufferPool is present, packet buffers are returned
// to the pool when the packet is discarded.
type BackPacketBuffer[T any] struct {
	Packets []Packet[T] // May be nil
	Pool    BufferPool  // May be nil
}

// Wipe clears stored packets and returns their buffers to the pool when
// available. Wipe is safe to call on a nil receiver.
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

// PacketsLeft returns the number of packets currently stored in the buffer.
func (b *BackPacketBuffer[T]) PacketsLeft() int {
	if b == nil {
		return 0
	}
	return len(b.Packets)
}

// PutBack pushes a packet (buffer + associated value) onto the stack. The
// provided buffer slice is not copied; callers that need ownership should
// ensure the slice won't be modified after PutBack.
func (b *BackPacketBuffer[T]) PutBack(bytes []byte, Assoc T) {
	if b == nil {
		return
	}
	b.Packets = append(b.Packets, Packet[T]{
		Buffer: bytes,
		Assoc:  Assoc,
	})
}

// ReadFrom pops the most recently PutBack packet and copies its buffer into
// p (up to len(p)). It returns the number of bytes copied and the associated
// value. If no packets are available it returns zero values. When a pool is
// present the packet buffer is returned to the pool. ReadFrom never returns
// a non-nil error.
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

// BackPacketBuffer returns the receiver to satisfy the WithBackPacketBuffer[T] interface.
func (b *BackPacketBuffer[T]) BackPacketBuffer() *BackPacketBuffer[T] {
	return b
}

type WithBackBuffer interface {
	BackBuffer() *BackBuffer
}

type WithBackPacketBuffer[T any] interface {
	BackPacketBuffer() *BackPacketBuffer[T]
}

// NewBackBuffer constructs a BackBuffer optionally reusing data from parent
// and prepending any provided bufs in front of that data. The returned
// BackBuffer does not take ownership of a non-nil parent; it copies slices as
// necessary.
func NewBackBuffer(pool BufferPool, parent WithBackBuffer, bufs ...[]byte) BackBuffer {
	var bytes []byte
	if parent != nil {
		bytes = concatCopy(parent.BackBuffer().Bytes, bytes)
		if pool == nil {
			pool = parent.BackBuffer().Pool
		}
	}
	for _, b := range bufs {
		bytes = concatCopy(b, bytes)
	}
	return BackBuffer{
		Bytes: bytes,
		Pool:  pool,
	}
}

// NewBackPacketBuffer constructs a BackPacketBuffer optionally reusing
// packets from parent and appending the provided packets. The returned
func NewBackPacketBuffer[T any](pool BufferPool, parent WithBackPacketBuffer[T], packets ...Packet[T]) BackPacketBuffer[T] {
	var packs []Packet[T]
	if parent != nil {
		packs = concatCopy(parent.BackPacketBuffer().Packets, packs)
		if pool == nil {
			pool = parent.BackPacketBuffer().Pool
		}
	}
	packs = concatCopy(packets, packs)
	return BackPacketBuffer[T]{
		Packets: packs,
		Pool:    pool,
	}
}
