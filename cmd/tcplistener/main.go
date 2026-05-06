package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string)

	go func(ch chan string) {
		var err error
		b := make([]byte, 8)
		var currentLineBuffer []byte
		n := 0

		for err != io.EOF {
			n, err = f.Read(b)
			var parts [][]byte

			if n > 0 {
				parts = bytes.Split(b, []byte("\n"))
				currentLineBuffer = append(currentLineBuffer, parts[0]...)
			}

			if len(parts) > 1 {
				ch <- string(currentLineBuffer)
				currentLineBuffer = make([]byte, len(parts[1]))
				copy(currentLineBuffer, parts[1])
			}

			clear(b)
		}

		if len(currentLineBuffer) > 0 {
			out <- string(currentLineBuffer)
		}

		close(ch)
	}(out)

	return out
}

func main() {
	l, err := net.Listen("tcp", ":42069")

	if err != nil {
		log.Fatal("error while readin the file: ", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
		}

		log.Println("the connection accepted")
		ch := getLinesChannel(conn)

		for line := range ch {
			fmt.Println(line)
		}
	}

	defer l.Close()
}
