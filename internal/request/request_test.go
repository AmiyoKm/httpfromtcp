package request_test

import (
	"io"
	"strings"
	"testing"

	"github.com/AmiyoKm/httpfromtcp/internal/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		useChunkReader  bool
		chunkSize       int
		expectError     bool
		expectedMethod  string
		expectedTarget  string
		expectedVersion string
	}{
		{
			name:            "Good GET Request line (chunked)",
			input:           "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			useChunkReader:  true,
			chunkSize:       3,
			expectError:     false,
			expectedMethod:  "GET",
			expectedTarget:  "/",
			expectedVersion: "1.1",
		},
		{
			name:            "Good GET Request line with path",
			input:           "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			useChunkReader:  true,
			chunkSize:       1,
			expectError:     false,
			expectedMethod:  "GET",
			expectedTarget:  "/coffee",
			expectedVersion: "1.1",
		},
		{
			name:            "Good POST Request with path",
			input:           "POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nContent-Type: application/json\r\nContent-Length: 22\r\n\r\n{\"flavor\":\"dark mode\"}",
			useChunkReader:  true,
			chunkSize:       1,
			expectError:     false,
			expectedMethod:  "POST",
			expectedTarget:  "/coffee",
			expectedVersion: "1.1",
		},
		{
			name:           "Invalid number of parts in request line",
			input:          "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			useChunkReader: true,
			chunkSize:      1,
			expectError:    true,
		},
		{
			name:           "Invalid method (out of order) Request line",
			input:          "/ GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			useChunkReader: true,
			chunkSize:      1,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var reader io.Reader

			if tc.useChunkReader {
				reader = &chunkReader{
					data:            tc.input,
					numBytesPerRead: tc.chunkSize,
				}
			} else {
				reader = strings.NewReader(tc.input)
			}

			r, err := request.RequestFromReader(reader)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Equal(t, tc.expectedMethod, r.RequestLine.Method)
			assert.Equal(t, tc.expectedTarget, r.RequestLine.RequestTarget)
			assert.Equal(t, tc.expectedVersion, r.RequestLine.HttpVersion)
		})
	}
}
