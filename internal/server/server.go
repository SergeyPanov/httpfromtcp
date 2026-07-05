package server

import (
	"log"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/SergeyPanov/httpfromtcp/internal/request"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	wg       sync.WaitGroup
}

func Serve(port int) (*Server, error) {
	p := strconv.Itoa(port)
	l, err := net.Listen("tcp", ":"+p)

	if err != nil {
		return nil, err
	}

	server := &Server{listener: l, closed: atomic.Bool{}, wg: sync.WaitGroup{}}

	server.wg.Go(func() {
		server.listen()
	})

	return server, nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Println(err)
			continue
		}

		s.wg.Go(func() {
			s.handle(conn)
		})
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	resp := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 12\r\n" +
		"\r\n" +
		"Hello World!"

	request.RequestFromReader(conn)

	// fmt.Println(resp)
	conn.Write([]byte(resp))

}

func (s *Server) Close() error {
	s.closed.Store(true)
	err := s.listener.Close()
	s.wg.Wait()
	return err
}
