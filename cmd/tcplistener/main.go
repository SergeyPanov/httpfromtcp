package main

import (
	"fmt"
	"log"
	"net"

	"github.com/SergeyPanov/httpfromtcp/internal/request"
)

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
		req, err := request.RequestFromReader(conn)

		fmt.Printf("%+v\n", req.RequestLine)
		fmt.Printf("%+v\n", req.Headers)

	}

	defer l.Close()
}
