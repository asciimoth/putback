package putback

import (
	"io"
	"net"
	"net/netip"
	"os"
	"syscall"
	"time"
)

var (
	_ TCPConn = &net.TCPConn{}
	_ UDPConn = &net.UDPConn{}
)

type BufferPool interface {
	// Should return byte buffer with exactly provided length and any capacity
	GetBuffer(length int) []byte
	// Putted buffer may have arbitrary length and capacity
	PutBuffer(buf []byte)
}

type TCPConn interface {
	net.Conn
	ReadFrom(r io.Reader) (int64, error)
	CloseRead() error
	CloseWrite() error
	File() (f *os.File, err error)
	MultipathTCP() (bool, error)
	SetKeepAlive(keepalive bool) error
	SetKeepAliveConfig(config net.KeepAliveConfig) error
	SetKeepAlivePeriod(d time.Duration) error
	SetLinger(sec int) error
	SetNoDelay(noDelay bool) error
	SetReadBuffer(bytes int) error
	SetWriteBuffer(bytes int) error
	SyscallConn() (syscall.RawConn, error)
	WriteTo(w io.Writer) (int64, error)
}

type UDPConn interface {
	net.PacketConn
	net.Conn
	File() (f *os.File, err error)
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	ReadFromUDPAddrPort(b []byte) (n int, addr netip.AddrPort, err error)
	ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error)
	ReadMsgUDPAddrPort(b, oob []byte) (n, oobn, flags int, addr netip.AddrPort, err error)
	RemoteAddr() net.Addr
	SetWriteBuffer(bytes int) error
	SyscallConn() (syscall.RawConn, error)
	WriteMsgUDP(b, oob []byte, addr *net.UDPAddr) (n, oobn int, err error)
	WriteMsgUDPAddrPort(b, oob []byte, addr netip.AddrPort) (n, oobn int, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
	WriteToUDPAddrPort(b []byte, addr netip.AddrPort) (int, error)
}
