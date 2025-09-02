package request

import (
	"io"
	"strings"
	"testing"

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

			r, err := RequestFromReader(reader)

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

func TestParseHeaders(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

	val, _ := r.Headers.Get("host")
	assert.Equal(t, "localhost:42069", val)

	val, _ = r.Headers.Get("user-agent")
	assert.Equal(t, "curl/7.81.0", val)

	val, _ = r.Headers.Get("accept")
	assert.Equal(t, "*/*", val)

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestParseBody(t *testing.T) {
	// Test: Standard Body
	ttb := []struct {
		name        string
		data        string
		chunkSize   int
		expectError bool
		expectedBody string
	}{
		{
			name: "Standard Body",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			chunkSize: 3,
			expectError: false,
			expectedBody: "hello world!\n",
		},
		{
			name: "Empty Body, 0 reported content length",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 0\r\n" +
				"\r\n",
			chunkSize: 3,
			expectError: false,
			expectedBody: "",
		},
		{
			name: "Empty Body, no reported content length",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			chunkSize: 3,
			expectError: false,
			expectedBody: "",
		},
		{
			name: "Body shorter than reported content length",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content",
			chunkSize: 3,
			expectError: true,
			expectedBody: "",
		},
		{
			name: "No Content-Length but Body Exists",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n" +
				"body without length",
			chunkSize: 3,
			expectError: false,
			expectedBody: "body without length",
		},
	}

	for _, tc := range ttb {
		t.Run(tc.name, func(t *testing.T) {
			reader := &chunkReader{
				data: tc.data,
				numBytesPerRead: tc.chunkSize,
			}
			r, err := RequestFromReader(reader)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, r)
				assert.Equal(t, tc.expectedBody, string(r.Body))
			}
		})
	}
}
