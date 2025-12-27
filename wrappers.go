package putback

import (
	"io"
	"net"
	"net/netip"
)

type PutBackReader struct {
	io.Reader
	Buffer BackBuffer
}

func (pb *PutBackReader) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackReader) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.Reader, p)
}

type PutBackReadCloser struct {
	io.ReadCloser
	Buffer BackBuffer
}

func (pb *PutBackReadCloser) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackReadCloser) Close() error {
	pb.Buffer.Wipe()
	return pb.ReadCloser.Close()
}

func (pb *PutBackReadCloser) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.ReadCloser, p)
}

type PutBackReadWriter struct {
	io.ReadWriter
	Buffer BackBuffer
}

func (pb *PutBackReadWriter) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackReadWriter) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.ReadWriter, p)
}

type PutBackReadWriteCloser struct {
	io.ReadWriteCloser
	Buffer BackBuffer
}

func (pb *PutBackReadWriteCloser) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackReadWriteCloser) Close() error {
	pb.Buffer.Wipe()
	return pb.ReadWriteCloser.Close()
}

func (pb *PutBackReadWriteCloser) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.ReadWriteCloser, p)
}

type PutBackConn struct {
	net.Conn
	Buffer BackBuffer
}

func (pb *PutBackConn) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackConn) Close() error {
	pb.Buffer.Wipe()
	return pb.Conn.Close()
}

func (pb *PutBackConn) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.Conn, p)
}

type PutBackTCPConn struct {
	TCPConn
	Buffer BackBuffer
}

func (pb *PutBackTCPConn) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackTCPConn) Close() error {
	pb.Buffer.Wipe()
	return pb.TCPConn.Close()
}

func (pb *PutBackTCPConn) CloseRead() error {
	pb.Buffer.Wipe()
	return pb.TCPConn.CloseRead()
}

func (pb *PutBackTCPConn) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.TCPConn, p)
}

func (pb *PutBackTCPConn) WriteTo(w io.Writer) (int64, error) {
	return copyBuffer(w, pb)
}

type PutBackPacketConn struct {
	net.PacketConn
	Buffer BackPacketBuffer[net.Addr]
}

func (pb *PutBackPacketConn) PutBack(bytes []byte, addr net.Addr) {
	pb.Buffer.PutBack(bytes, addr)
}

func (pb *PutBackPacketConn) Close() error {
	pb.Buffer.Wipe()
	return pb.PacketConn.Close()
}

func (pb *PutBackPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = pb.Buffer.ReadFrom(p)
	if n != 0 {
		return
	}
	return pb.PacketConn.ReadFrom(p)
}

type PutBackUDPConn struct {
	UDPConn
	Buffer BackPacketBuffer[*net.UDPAddr]
}

func (pb *PutBackUDPConn) PutBack(bytes []byte, addr *net.UDPAddr) {
	pb.Buffer.PutBack(bytes, addr)
}

func (pb *PutBackUDPConn) Close() error {
	pb.Buffer.Wipe()
	return pb.UDPConn.Close()
}

func (pb *PutBackUDPConn) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	n, addr, err = pb.Buffer.ReadFrom(b)
	if n != 0 {
		return
	}
	return pb.UDPConn.ReadFromUDP(b)
}

func (pb *PutBackUDPConn) ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error) {
	n, addr, err = pb.ReadFromUDP(b)
	return
}

func (pb *PutBackUDPConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return pb.ReadFromUDP(p)
}

func (pb *PutBackUDPConn) ReadFromUDPAddrPort(b []byte) (n int, addr netip.AddrPort, err error) {
	var ua *net.UDPAddr
	n, ua, err = pb.Buffer.ReadFrom(b)
	if n != 0 {
		addr = udpAddrToAddrPort(ua)
		return
	}
	return pb.UDPConn.ReadFromUDPAddrPort(b)
}

func (pb *PutBackUDPConn) ReadMsgUDPAddrPort(b, oob []byte) (n, oobn, flags int, addr netip.AddrPort, err error) {
	n, addr, err = pb.ReadFromUDPAddrPort(b)
	return
}

func (pb *PutBackUDPConn) Read(p []byte) (n int, err error) {
	n, _, err = pb.ReadFromUDP(p)
	return
}

func WrapConn(conn net.Conn, bytes []byte, pool BufferPool) net.Conn {
	buf := BackBuffer{
		Bytes: bytes,
		Pool:  pool,
	}
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
