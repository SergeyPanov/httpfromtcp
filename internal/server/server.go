package server

import (
	"bytes"
	"log"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/SergeyPanov/httpfromtcp/internal/request"
	"github.com/SergeyPanov/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handlerF response.Handler
	wg       sync.WaitGroup
}

func Serve(port int, handler response.Handler) (*Server, error) {
	p := strconv.Itoa(port)
	l, err := net.Listen("tcp", ":"+p)

	if err != nil {
		return nil, err
	}

	server := &Server{listener: l, closed: atomic.Bool{}, handlerF: handler, wg: sync.WaitGroup{}}

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

	req, err := request.RequestFromReader(conn)

	if err != nil {
		log.Println(err)
		return
	}

	var buff bytes.Buffer

	handlerError := s.handlerF(&buff, &req)

	if handlerError != nil {
		msgBytes := []byte(handlerError.Message)
		headers := response.GetDefaultHeaders(len(msgBytes))

		response.WriteStatusLine(conn, handlerError.Status)
		response.WriteHeaders(conn, headers)

		conn.Write(msgBytes)

	} else {
		respBytes := buff.Bytes()
		headers := response.GetDefaultHeaders(len(respBytes))

		response.WriteStatusLine(conn, response.Success)
		response.WriteHeaders(conn, headers)

		conn.Write(respBytes)
	}

}

func (s *Server) Close() error {
	s.closed.Store(true)
	err := s.listener.Close()
	s.wg.Wait()
	return err
}
