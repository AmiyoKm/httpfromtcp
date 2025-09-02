package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/AmiyoKm/httpfromtcp/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

var ErrorMalformedRequestLine = fmt.Errorf("malformed request-line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")

var SEPERATOR = []byte("\r\n")

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
	state       parserState
}

func NewRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    []byte{},
	}
}

func getInt(headers *headers.Headers, key string, defaultValue int) int {
	valStr, ok := headers.Get(key)
	if !ok {
		return defaultValue
	}
	value, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState

		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)

			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				r.state = StateBody
			}
		case StateBody:
			length := getInt(r.Headers, "content-length", -1)

			if length >= 0 {
				// Content-Length specified - read exactly that many bytes
				remaining := min(length-len(r.Body), len(currentData))
				if remaining > 0 {
					r.Body = append(r.Body, currentData[:remaining]...)
					read += remaining

					if len(r.Body) >= length {
						r.state = StateDone
					}
				} else {
					r.state = StateDone
				}
			} else {
				// No Content-Length - consume all remaining data as body
				r.Body = append(r.Body, currentData...)
				read += len(currentData)
			}

		case StateDone:
			break outer
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}
func (r *Request) error() bool {
	return r.state == StateError
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPERATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPERATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, read, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, read, ErrorMalformedRequestLine
	}

	method := string(parts[0])
	if !isValidMethod(method) {
		return nil, 0, ErrorMalformedRequestLine
	}

	target := string(parts[1])
	if !isValidTarget(target) {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()
	buf := make([]byte, 4096)
	bufLen := 0

	for !request.done() && !request.error() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			if err == io.EOF {
				// Handle incomplete body if Content-Length is set
				contentLength := getInt(request.Headers, "content-length", -1)
				if contentLength > 0 && len(request.Body) < contentLength {
					return nil, fmt.Errorf("body shorter than reported content-length: got %d, expected %d",
						len(request.Body), contentLength)
				}
				// If headers and body are parsed, mark as done
				if request.state == StateBody {
					request.state = StateDone
				}
				break

			} else {
				return nil, err
			}
		}

		if n > 0 {
			bufLen += n
			readN, parseErr := request.parse(buf[:bufLen])
			if parseErr != nil {
				return nil, parseErr
			}
			copy(buf, buf[readN:bufLen])
			bufLen -= readN
		}
	}
	return request, nil
}
