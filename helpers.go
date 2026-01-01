package putback

import (
	"io"
	"net"
	"net/netip"
)

func concatCopy[T any](a, b []T) []T {
	if a == nil && b == nil {
		return nil
	}
	c := make([]T, len(a)+len(b))
	copy(c, a)
	copy(c[len(a):], b)
	return c
}

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

// copyBuffer is like io.Copy but
// without specific logic for io.WriteTo and io.ReadFrom
func copyBuffer(dst io.Writer, src io.Reader) (written int64, err error) {
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	buf := make([]byte, size)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = io.ErrNoProgress
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
