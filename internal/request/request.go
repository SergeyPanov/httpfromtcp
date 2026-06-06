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

	for {
		n, err := reader.Read(buf[readToIndex:])
		readToIndex += n

		if err != nil && err != io.EOF {
			return Request{}, nil
		}

		if readToIndex >= len(buf) {
			dbuf := make([]byte, 2*len(buf))
			copy(dbuf, buf)
			buf = dbuf
		}

		if err == io.EOF {
			break
		}
	}

	request := Request{
		State: initialized,
	}

	for request.State != done {
		n, err := request.parse(buf)

		if err != nil {
			return Request{}, err
		}
		tmpBuf := make([]byte, len(buf)-n)
		copy(tmpBuf, buf[n:])
		buf = tmpBuf
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
		headers := headers.NewHeaders()
		consumed, doneParsing, err := headers.Parse(data)

		if err != nil {
			return consumed, err
		}

		if doneParsing {
			r.State = requestStateParsingBody
			return consumed, nil
		}

		r.Headers = headers

		return consumed, nil

	case requestStateParsingBody:
		trimmed := bytes.TrimRight(data, "\x00")

		contentLength, ok := r.Headers.Get("Content-Length")
		r.State = done

		if !ok {
			return 0, nil
		}

		length, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, err
		}

		if len(trimmed) != length+len(clrf) {
			return 0, errors.New("the body length doesn't match to the Content-Length")
		}

		r.Body = trimmed[len(clrf):]

		return len(data), nil

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
