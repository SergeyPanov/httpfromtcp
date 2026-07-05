package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/SergeyPanov/httpfromtcp/internal/request"
)

func main() {
	l, err := net.Listen("tcp", ":42069")

	if err != nil {
		log.Fatal("error while readin the file: ", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println(err)
			}

			log.Println("the connection accepted")
			req, err := request.RequestFromReader(conn)

			fmt.Printf("%+v\n", req.RequestLine)
			fmt.Printf("%+v\n", req.Headers)
			fmt.Printf("%+v\n", string(req.Body))

		}
	}()

	<-sigChan
	l.Close()
}
