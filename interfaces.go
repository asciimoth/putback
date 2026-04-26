package putback

import (
	"io"
	"net"
	"net/netip"
	"time"
)

var (
	_ TCPConn = &net.TCPConn{}
	_ UDPConn = &net.UDPConn{}
)

type TCPConn interface {
	net.Conn

	// ReadFrom(r io.Reader) (int64, error)
	io.ReaderFrom

	// WriteTo(w io.Writer) (int64, error)
	io.WriterTo

	SetKeepAlive(keepalive bool) error
	SetKeepAliveConfig(config net.KeepAliveConfig) error
	SetKeepAlivePeriod(d time.Duration) error
	SetLinger(sec int) error
	SetNoDelay(noDelay bool) error

	CloseRead() error
	CloseWrite() error
}

type UDPConn interface {
	net.Conn

	// ReadFrom(p []byte) (n int, addr Addr, err error)
	// WriteTo(p []byte, addr Addr) (n int, err error)
	net.PacketConn

	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	ReadFromUDPAddrPort(b []byte) (n int, addr netip.AddrPort, err error)

	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
	WriteToUDPAddrPort(b []byte, addr netip.AddrPort) (int, error)

	ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error)
	ReadMsgUDPAddrPort(
		b, oob []byte,
	) (n, oobn, flags int, addr netip.AddrPort, err error)

	WriteMsgUDP(b, oob []byte, addr *net.UDPAddr) (n, oobn int, err error)
	WriteMsgUDPAddrPort(
		b, oob []byte,
		addr netip.AddrPort,
	) (n, oobn int, err error)
}
