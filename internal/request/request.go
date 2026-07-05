package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/SergeyPanov/httpfromtcp/internal/headers"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	RequestLine RequestLine `validate:"required"`
	Headers     headers.Headers
	Body        []byte
	State       ParserState
}

type RequestLine struct {
	Method        string `validate:"required,oneof=GET POST PUT PATCH DELETE"`
	RequestTarget string `validate:"required"`
	HttpVersion   string `validate:"required,eq=1.1"`
}

type ParserState int

const (
	initialized ParserState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	done
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := Request{
		State:   initialized,
		Headers: headers.NewHeaders(),
	}

	for request.State != done {
		if readToIndex >= len(buf) {
			dbuf := make([]byte, 2*len(buf))
			copy(dbuf, buf)
			buf = dbuf
		}

		n, err := reader.Read(buf[readToIndex:])
		readToIndex += n

		if err != nil && err != io.EOF {
			return Request{}, err
		}

		// Parse as much as possible from what we have so far.
		for request.State != done {
			consumed, parseErr := request.parse(buf[:readToIndex])
			if parseErr != nil {
				return Request{}, parseErr
			}

			if consumed == 0 {
				// Need more data.
				break
			}

			copy(buf, buf[consumed:readToIndex])
			readToIndex -= consumed
		}

		if err == io.EOF && request.State != done {
			return Request{}, errors.New("unexpected EOF: incomplete request")
		}
	}

	// Report no headers as a nil map rather than an empty one.
	if len(request.Headers) == 0 {
		request.Headers = nil
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(request)

	return request, err
}

func (r *Request) parse(data []byte) (int, error) {
	clrf := []byte("\r\n")

	switch r.State {
	case done:
		return 0, errors.New("error: trying to read data in a done state")

	case initialized:
		consumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if consumed == 0 {
			return 0, nil
		}

		requestLineBytes := data[:consumed-len(clrf)]
		parts := bytes.Split(requestLineBytes, []byte(" "))

		if len(parts) != 3 {
			return 0, errors.New("error: invalid request line")
		}

		r.RequestLine = RequestLine{
			Method:        string(parts[0]),
			RequestTarget: string(parts[1]),
			HttpVersion:   strings.Split(string(parts[2]), "/")[1],
		}

		r.State = requestStateParsingHeaders

		return consumed, nil

	case requestStateParsingHeaders:
		consumed, doneParsing, err := r.Headers.Parse(data)

		if err != nil {
			return consumed, err
		}

		if doneParsing {
			r.State = requestStateParsingBody
		}

		return consumed, nil

	case requestStateParsingBody:
		contentLength, ok := r.Headers.Get("Content-Length")

		// No body expected.
		if !ok {
			r.State = done
			return 0, nil
		}

		length, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, err
		}

		// Wait until the full body has arrived.
		if len(data) < length {
			return 0, nil
		}

		r.Body = data[:length]
		r.State = done

		return length, nil

	default:
		return 0, errors.New("error: unknown state")

	}
}

func parseRequestLine(data []byte) (int, error) {
	clrf := []byte("\r\n")
	idx := bytes.Index(data, clrf)
	if idx < 0 {
		return 0, nil
	}

	return idx + len(clrf), nil

}
