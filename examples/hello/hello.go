package main

import (
	"fmt"
	"io"
	"net"

	"github.com/asciimoth/putback"
)

func main() {
	// echo "WORLD" | nc -c 127.0.0.1 3333
	l, err := net.Listen("tcp4", "127.0.0.1:3333")
	if err != nil {
		panic(err)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}

		c = putback.WrapConn(c, []byte("hello "), nil)

		b, _ := io.ReadAll(c)

		fmt.Println(string(b))
	}
}
