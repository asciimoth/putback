package putback

import (
	"io"
	"net"
	"net/netip"
)

func readJoin(a, b io.Reader, p []byte) (n int, err error) {
	if a != nil {
		n, err = a.Read(p)
		if n != 0 {
			return
		}
	}
	return b.Read(p)
}

func udpAddrToAddrPort(a *net.UDPAddr) netip.AddrPort {
	if a == nil {
		return netip.AddrPort{}
	}
	if a.IP == nil {
		return netip.AddrPort{}
	}

	// prefer the canonical 4-byte form for IPv4
	if v4 := a.IP.To4(); v4 != nil {
		na, ok := netip.AddrFromSlice(v4)
		if !ok {
			return netip.AddrPort{}
		}
		return netip.AddrPortFrom(na, uint16(a.Port))
	}

	// IPv6 (or 4-in-6 if the slice is 16 bytes)
	v16 := a.IP.To16()
	if v16 == nil {
		return netip.AddrPort{}
	}
	na, ok := netip.AddrFromSlice(v16)
	if !ok {
		return netip.AddrPort{}
	}
	na = na.WithZone(a.Zone) // keep IPv6 zone if present
	return netip.AddrPortFrom(na, uint16(a.Port))
}
