// Package putback provides wrapper types around common io and net interfaces
// that add support for putting bytes or packets back so they can be read
// again.
package putback

import (
	"io"
	"net"
	"net/netip"
)

// PutBackReader wraps an io.Reader and allows bytes to be put back so they
// will be returned by subsequent Read calls.
type PutBackReader struct {
	io.Reader
	Buffer BackBuffer
}

// PutBack prepends bytes so they will be read before the underlying Reader.
func (pb *PutBackReader) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

// Read reads from the internal BackBuffer first and then from the underlying
// Reader once the buffer is exhausted.
func (pb *PutBackReader) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.Reader, p)
}

// PutBackReadCloser wraps an io.ReadCloser with put-back support. When closed,
// the internal buffer is wiped before closing the underlying ReadCloser.
type PutBackReadCloser struct {
	io.ReadCloser
	Buffer BackBuffer
}

// PutBack prepends bytes so they will be read before the underlying Reader.
func (pb *PutBackReadCloser) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

// Close wipes the internal buffer and then closes the underlying ReadCloser.
func (pb *PutBackReadCloser) Close() error {
	pb.Buffer.Wipe()
	return pb.ReadCloser.Close()
}

// Read reads from the internal BackBuffer first and then from the underlying
// ReadCloser once the buffer is exhausted.
func (pb *PutBackReadCloser) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.ReadCloser, p)
}

// PutBackReadWriter wraps an io.ReadWriter with put-back support for reads.
type PutBackReadWriter struct {
	io.ReadWriter
	Buffer BackBuffer
}

// PutBack prepends bytes so they will be read before the underlying Reader.
func (pb *PutBackReadWriter) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

// Read reads from the internal BackBuffer first and then from the underlying
// Reader once the buffer is exhausted.
func (pb *PutBackReadWriter) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.ReadWriter, p)
}

// PutBackReadWriteCloser wraps an io.ReadWriteCloser with put-back support for reads.
type PutBackReadWriteCloser struct {
	io.ReadWriteCloser
	Buffer BackBuffer
}

// PutBack prepends bytes so they will be read before the underlying Reader.
func (pb *PutBackReadWriteCloser) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

// Close wipes the internal buffer and then closes the underlying ReadCloser.
func (pb *PutBackReadWriteCloser) Close() error {
	pb.Buffer.Wipe()
	return pb.ReadWriteCloser.Close()
}

// Read reads from the internal BackBuffer first and then from the underlying
// Reader once the buffer is exhausted.
func (pb *PutBackReadWriteCloser) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.ReadWriteCloser, p)
}

// PutBackConn wraps a net.Conn with put-back support for reads.
type PutBackConn struct {
	net.Conn
	Buffer BackBuffer
}

// PutBack prepends bytes so they will be read before the underlying Conn.
func (pb *PutBackConn) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

// Close wipes the internal buffer and then closes the underlying Conn.
func (pb *PutBackConn) Close() error {
	pb.Buffer.Wipe()
	return pb.Conn.Close()
}

// Read reads from the internal BackBuffer first and then from the underlying
// Conn once the buffer is exhausted.
func (pb *PutBackConn) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.Conn, p)
}

// PutBackTCPConn wraps a net.TCPConn with put-back support for reads.
type PutBackTCPConn struct {
	TCPConn
	Buffer BackBuffer
}

// PutBack prepends bytes so they will be read before the underlying Conn.
func (pb *PutBackTCPConn) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

// Close wipes the internal buffer and then closes the underlying Conn.
func (pb *PutBackTCPConn) Close() error {
	pb.Buffer.Wipe()
	return pb.TCPConn.Close()
}

// CloseRead wipes the internal buffer and then half-closes the read side of
// the underlying TCPConn.
func (pb *PutBackTCPConn) CloseRead() error {
	pb.Buffer.Wipe()
	return pb.TCPConn.CloseRead()
}

// Read reads from the internal BackBuffer first and then from the underlying
// TCPConn once the buffer is exhausted.
func (pb *PutBackTCPConn) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.TCPConn, p)
}

// WriteTo implements io.WriterTo by copying data from the connection to w.
func (pb *PutBackTCPConn) WriteTo(w io.Writer) (int64, error) {
	return copyBuffer(w, pb)
}

// PutBackPacketConn wraps a net.PacketConn and allows received packets to be
// put back so they will be returned by subsequent ReadFrom calls.
type PutBackPacketConn struct {
	net.PacketConn
	Buffer BackPacketBuffer[net.Addr]
}

// PutBack pushes a packet back so it will be returned by the next ReadFrom.
func (pb *PutBackPacketConn) PutBack(bytes []byte, addr net.Addr) {
	pb.Buffer.PutBack(bytes, addr)
}

// Close wipes the internal packet buffer and then closes the underlying
// PacketConn.
func (pb *PutBackPacketConn) Close() error {
	pb.Buffer.Wipe()
	return pb.PacketConn.Close()
}

// ReadFrom first attempts to read a packet from the internal buffer. If none
// are available it delegates to the underlying PacketConn.
func (pb *PutBackPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = pb.Buffer.ReadFrom(p)
	if n != 0 {
		return
	}
	return pb.PacketConn.ReadFrom(p)
}

// PutBackUDPConn wraps a UDPConn and allows UDP packets to be put back and
// re-read. It supports the common UDP read variants provided by net.UDPConn.
type PutBackUDPConn struct {
	UDPConn
	Buffer BackPacketBuffer[*net.UDPAddr]
}

// PutBack pushes a packet back so it will be returned by the next reads.
func (pb *PutBackUDPConn) PutBack(bytes []byte, addr *net.UDPAddr) {
	pb.Buffer.PutBack(bytes, addr)
}

// Close wipes the internal packet buffer and then closes the underlying
// UDPConn.
func (pb *PutBackUDPConn) Close() error {
	pb.Buffer.Wipe()
	return pb.UDPConn.Close()
}

// ReadFromUDP reads a packet from the internal buffer first and falls back to
// the underlying UDPConn if no buffered packets are available.
func (pb *PutBackUDPConn) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	n, addr, err = pb.Buffer.ReadFrom(b)
	if n != 0 {
		return
	}
	return pb.UDPConn.ReadFromUDP(b)
}

// ReadMsgUDP is implemented in terms of ReadFromUDP. OOB data is not
// supported and oobn and flags are always zero.
func (pb *PutBackUDPConn) ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error) {
	n, addr, err = pb.ReadFromUDP(b)
	return
}

// ReadFrom implements net.PacketConn by delegating to ReadFromUDP.
func (pb *PutBackUDPConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return pb.ReadFromUDP(p)
}

// ReadFromUDPAddrPort reads a packet returning a netip.AddrPort. Buffered
// packets are returned first, falling back to the underlying UDPConn.
func (pb *PutBackUDPConn) ReadFromUDPAddrPort(b []byte) (n int, addr netip.AddrPort, err error) {
	var ua *net.UDPAddr
	n, ua, err = pb.Buffer.ReadFrom(b)
	if n != 0 {
		addr = udpAddrToAddrPort(ua)
		return
	}
	return pb.UDPConn.ReadFromUDPAddrPort(b)
}

// ReadMsgUDPAddrPort is implemented in terms of ReadFromUDPAddrPort. OOB data
// is not supported and oobn and flags are always zero.
func (pb *PutBackUDPConn) ReadMsgUDPAddrPort(b, oob []byte) (n, oobn, flags int, addr netip.AddrPort, err error) {
	n, addr, err = pb.ReadFromUDPAddrPort(b)
	return
}

// Read implements io.Reader by reading a UDP packet and discarding the source
// address.
func (pb *PutBackUDPConn) Read(p []byte) (n int, err error) {
	n, _, err = pb.ReadFromUDP(p)
	return
}

// WrapConn wraps a net.Conn with a put-back capable connection. Any initial
// bytes are made available for reading before data from the connection. If
// the provided connection already supports put-back, its existing buffer is
// reused as the parent. TCP connections are wrapped with PutBackTCPConn to
// preserve TCP-specific methods.
func WrapConn(conn net.Conn, bytes []byte, pool BufferPool) net.Conn {
	var parent WithBackBuffer
	if p, ok := conn.(WithBackBuffer); ok {
		parent = p
	}
	buf := NewBackBuffer(pool, parent, bytes)
	if tcp, ok := conn.(TCPConn); ok {
		return &PutBackTCPConn{
			TCPConn: tcp,
			Buffer:  buf,
		}
	}
	return &PutBackConn{
		Conn:   conn,
		Buffer: buf,
	}
}
