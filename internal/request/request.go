package request

import (
	"errors"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Request struct {
	RequestLine RequestLine `validate:"required"`
}

type RequestLine struct {
	Method        string `validate:"required,oneof=GET POST PUT PATCH DELETE"`
	RequestTarget string `validate:"required"`
	HttpVersion   string `validate:"required,eq=1.1"`
}

func RequestFromReader(reader io.Reader) (Request, error) {
	requestLine, err := parseRequestLine(reader)

	if err != nil {
		return Request{}, err
	}
	req := Request{
		RequestLine: *requestLine,
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(req)

	return req, err
}

func parseRequestLine(reader io.Reader) (*RequestLine, error) {
	content, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(content), "\r\n")

	if len(parts) <= 0 {
		return nil, errors.New("no request line in the request")
	}

	requestLineParts := strings.Split(parts[0], " ")

	if len(requestLineParts) != 3 {
		return nil, errors.New("invalid request line; request-line  = method SP request-target SP HTTP-version")
	}

	return &RequestLine{
		Method:        requestLineParts[0],
		RequestTarget: requestLineParts[1],
		HttpVersion:   strings.Split(requestLineParts[2], "/")[1],
	}, nil

}
