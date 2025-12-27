package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"

	"github.com/asciimoth/putback"
)

type Config struct {
	Listen  string            // Addr to listen at
	Mapping map[string]string // proto type -> proxy addr
}

func readConfig() Config {
	listen := flag.String("listen", "127.0.0.1:3333", "addr to listen")
	httpProxy := flag.String("http", "127.0.0.1:80", "http proxying addr")
	sshProxy := flag.String("ssh", "127.0.0.1:22", "http proxying addr")
	socksProxy := flag.String("socks", "127.0.0.1:1080", "http proxying addr")

	flag.Parse()

	cfg := Config{
		Listen: *listen,
		Mapping: map[string]string{
			"http":  *httpProxy,
			"ssh":   *sshProxy,
			"socks": *socksProxy,
		},
	}
	return cfg
}

// accept accepts incoming connection and reads 3-byte header from it.
func accept(l net.Listener) (c net.Conn, h [3]byte) {
	var err error
	for {
		c, err = l.Accept()
		if err != nil {
			// This is a simple example so we can just panic
			panic(err)
		}

		_, err = io.ReadFull(c, h[:])
		if err != nil {
			_ = c.Close()
			continue
		}

		return
	}
}

// guessProto guesses protocol based on 3-byte header
func guessProto(h [3]byte) string {
	if h[0] == 4 || h[0] == 5 {
		fmt.Println("socks protocol with header:", h)
		return "socks"
	}

	str := string(h[:])
	lower := strings.ToLower(str)

	if lower == "ssh" {
		fmt.Println("ssh protocol with header:", str)
		return "ssh"
	}

	if slices.Contains([]string{
		"get", "head", "post", "put", "delete",
		"connect", "options", "trace", "patch",
		"pri", "http",
	}, lower) {
		fmt.Println("http protocol with header:", str)
		return "http"
	}

	fmt.Println("unknown protocol with header:", h)
	return "unknown"
}

// proxy dialing outgoing connection to addr and copy all bytes between it and in
func proxy(addr string, in net.Conn) {
	out, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		defer out.Close()
		_, _ = io.Copy(out, in)
	}()

	go func() {
		defer in.Close()
		_, _ = io.Copy(in, out)
	}()
}

func main() {
	cfg := readConfig()

	l, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		panic(err)
	}

	fmt.Println("listening at", cfg.Listen)

	for {
		c, header := accept(l)

		proto := guessProto(header)

		addr, ok := cfg.Mapping[proto]
		if !ok {
			fmt.Printf("there is no mapped addr for %s protocol; dropping\n", proto)
			_ = c.Close()
			continue
		}
		fmt.Println("proxying to", addr)

		// Putting header bytes back to conn
		c = putback.WrapConn(c, header[:], nil)

		proxy(addr, c)
	}
}
