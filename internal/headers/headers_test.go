package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidSingleHeader(t *testing.T) {
	headers := NewHeaders()

	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	require.NotNil(t, headers)

	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestInvalidSpacingHeader(t *testing.T) {
	headers := NewHeaders()

	data := []byte("       Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)

	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestManyHeadersWithExistingHeaders(t *testing.T) {
	headers := NewHeaders()

	data := [][]byte{
		[]byte("Host: localhost:42069\r\n"),
		[]byte("Content-Type: text/html\r\n"),
		[]byte("Accept: application/json\r\n\r\n"),
	}
	for _, d := range data {
		headers.Parse(d)
	}

	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, "text/html", headers["Content-Type"])
	assert.Equal(t, "application/json", headers["Accept"])
}

func TestValidDone(t *testing.T) {
	headers := NewHeaders()

	data := []byte("\r\n\r\n")

	_, done, err := headers.Parse(data)

	assert.True(t, done)
	assert.NoError(t, err)
}
