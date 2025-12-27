# putback
A minimal library with wrappers for common I/O interfaces, adding the ability to return read bytes back to the stream for subsequent reading.

[Split example:](https://github.com/asciimoth/putback/tree/master/examples/split)  
[![asciicast](https://asciinema.org/a/764258.svg)](https://asciinema.org/a/764258)  

Minimal [hello X](https://github.com/asciimoth/putback/tree/master/examples/hello/hello.go) example:
```go
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
```

## License
Files in this repository are distributed under the CC0 license.  

<p xmlns:dct="http://purl.org/dc/terms/">
  <a rel="license"
     href="http://creativecommons.org/publicdomain/zero/1.0/">
    <img src="http://i.creativecommons.org/p/zero/1.0/88x31.png" style="border-style: none;" alt="CC0" />
  </a>
  <br />
  To the extent possible under law,
  <a rel="dct:publisher"
     href="https://github.com/asciimoth">
    <span property="dct:title">ASCIIMoth</span></a>
  has waived all copyright and related or neighboring rights to
  <span property="dct:title">putback</span>.
</p>
