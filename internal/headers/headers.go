package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type ValueProcessingState int

const (
	SkippingPrefixWhitespace ValueProcessingState = iota
	SkippingPostfixWhitespace
	ConsumingValue
)

type Headers map[string]string

var allowedSpecialChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

const CR byte = '\r'
const LF byte = '\n'

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Not enough data yet to tell whether this is the terminating CRLF.
	if len(data) < 2 {
		return 0, false, nil
	}

	// A leading CRLF marks the end of the headers section.
	if isFinished(string(data)) {
		return 2, true, nil
	}

	// We need a complete header line (terminated by CRLF) before parsing.
	idx := bytes.Index(data, []byte{CR, LF})
	if idx < 0 {
		// Incomplete line: ask the caller for more data.
		return 0, false, nil
	}

	// The line plus its trailing CRLF, which extractValue expects.
	line := data[:idx+2]

	key, consumedK, err := extractKey(string(line))
	if err != nil {
		return 0, false, err
	}
	key = strings.ToLower(key)

	// field-name + ":"
	val, consumedV, err := extractValue(string(line[consumedK+1:]))
	if err != nil {
		return 0, false, err
	}

	v, ok := h[key]
	if !ok {
		h[key] = val
	} else {
		h[key] = strings.Join([]string{v, val}, ", ")
	}

	return consumedK + 1 + consumedV, false, nil
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	v, ok := h[key]
	return v, ok
}

func isFinished(data string) bool {
	return len(data) >= 2 && data[0] == CR && data[1] == LF
}

func extractKey(data string) (string, int, error) {
	key := ""

	for i, ch := range data {
		if ch == ':' {

			return key, i, nil

		} else if unicode.IsUpper(rune(ch)) ||
			unicode.IsLower(rune(ch)) ||
			unicode.IsDigit(rune(ch)) ||
			bytes.ContainsAny([]byte(string(ch)), string(allowedSpecialChars)) {

			key += string(ch)

		} else {
			return "", -1, fmt.Errorf("invalid character in the header key: %c", ch)
		}
	}

	return "", -1, errors.New("invalid header format")
}

func extractValue(data string) (string, int, error) {
	state := SkippingPrefixWhitespace
	var value string = ""

	for i, ch := range data {
		if i > 0 && byte(ch) == LF && data[i-1] == CR {
			return value, i + 1, nil
		}

		switch state {
		case SkippingPrefixWhitespace:
			if !unicode.IsSpace(ch) {
				state = ConsumingValue
				value += string(ch)
			}
			break
		case ConsumingValue:
			if !unicode.IsSpace(ch) {
				value += string(ch)
			} else {
				state = SkippingPostfixWhitespace
			}
			break
		case SkippingPostfixWhitespace:
			if !unicode.IsSpace(ch) {
				return "", -1, errors.New("invalid value")
			}
			break
		}
	}

	return "", -1, errors.New("invalid value")
}
