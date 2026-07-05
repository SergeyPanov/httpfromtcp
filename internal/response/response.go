package response

import (
	"bytes"
	"io"
	"strconv"

	"github.com/SergeyPanov/httpfromtcp/internal/headers"
	"github.com/SergeyPanov/httpfromtcp/internal/request"
)

type HandlerError struct {
	Status  StatusCode
	Message string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case Success:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return err
	case BadRequest:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return err
	case ServerError:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return err
	default:
		_, err := w.Write([]byte("HTTP/1.1 500\r\n"))
		return err
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers["Content-Length"] = strconv.Itoa(contentLen)
	headers["Connection"] = "closed"
	headers["Content-Type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	var b bytes.Buffer
	for k, v := range headers {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(v)
		b.WriteString("\r\n")
	}
	// Blank line terminating the header section.
	b.WriteString("\r\n")

	_, err := w.Write(b.Bytes())

	return err
}
