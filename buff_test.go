package putback_test

import (
	"bytes"
	"testing"

	"github.com/asciimoth/putback"
)

type mockPool struct{}

func (m *mockPool) GetBuffer(length int) []byte {
	return make([]byte, length, length+42)
}

func (m *mockPool) PutBuffer(buf []byte) {}

func TestBufferPool_PutBackAndRead_Simple(t *testing.T) {
	b := &putback.BackBuffer{}
	b.PutBack([]byte("world"))
	b.PutBack([]byte("hello "))

	buf := make([]byte, 11)
	n, err := b.Read(buf)
	if err != nil {
		t.Fatalf("unexpected non-nil error: %v", err)
	}
	if n != 11 {
		t.Fatalf("expected 11 bytes, got %d", n)
	}
	if !bytes.Equal(buf, []byte("hello world")) {
		t.Fatalf("unexpected data: %q", buf)
	}
	// further reads should return 0
	n2, err2 := b.Read(buf)
	if n2 != 0 || err2 != nil {
		t.Fatalf("expected 0,nil after drain; got %d,%v", n2, err2)
	}
}

func TestBufferPool_PartialReads(t *testing.T) {
	b := &putback.BackBuffer{}
	src := []byte("abcdef")
	b.PutBack(src)

	p := make([]byte, 3)
	n, _ := b.Read(p)
	if n != 3 || !bytes.Equal(p, []byte("abc")) {
		t.Fatalf("first read wrong: n=%d p=%q", n, p)
	}

	n2, _ := b.Read(p)
	if n2 != 3 || !bytes.Equal(p, []byte("def")) {
		t.Fatalf("second read wrong: n=%d p=%q", n2, p)
	}

	// drained
	n3, _ := b.Read(p)
	if n3 != 0 {
		t.Fatalf("expected drained, got %d", n3)
	}
}

func TestBufferPool_PushBackWithFreeSpace(t *testing.T) {
	// Simulate buffer with free space on left by creating Bytes with a
	// prefix of zeros and setting Pointer > 0.
	b := &putback.BackBuffer{}
	b.Bytes = append(make([]byte, 5), []byte("DATA")...)
	b.Pointer = 5

	b.PutBack([]byte("NEW"))
	// Now expected layout is Bytes[Pointer:] == "NEWDATA"
	if !bytes.Equal(b.Bytes[b.Pointer:], []byte("NEWDATA")) {
		t.Fatalf("unexpected layout: %q", b.Bytes[b.Pointer:])
	}
}
