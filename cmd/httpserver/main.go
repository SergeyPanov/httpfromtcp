package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SergeyPanov/httpfromtcp/internal/request"
	"github.com/SergeyPanov/httpfromtcp/internal/response"
	"github.com/SergeyPanov/httpfromtcp/internal/server"
)

const port = 42069

func handler(w io.Writer, req *request.Request) *response.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &response.HandlerError{
			Status:  response.BadRequest,
			Message: "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &response.HandlerError{
			Status:  response.ServerError,
			Message: "Woopsie, my bad\n",
		}

	default:
		w.Write([]byte("All good, frfr\n"))
		return nil
	}
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
