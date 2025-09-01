package headers

import (
	"bytes"
	"errors"
	"strings"
)

var errMalformedHeader = errors.New("header is malformed")

var rn = []byte("\r\n")

func isValidToken(key string) bool {
	if len(key) == 0 {
		return false
	}
	for _, c := range key {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case strings.ContainsRune("!#$%&'*+-.^_`|~", c):
		default:
			return false
		}
	}
	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", errMalformedHeader
	}

	key := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(key, []byte(" ")) {
		return "", "", errMalformedHeader
	}

	return string(key), string(value), nil
}

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(key string) string {
	return h.headers[strings.ToLower(key)]
}

func (h *Headers) Set(key string, val string) {
	key = strings.ToLower(key)
	if v, ok := h.headers[key]; ok {
		h.headers[key] = v + "," + val
	} else {
		h.headers[key] = val
	}
}

func (h *Headers) ForEach(cb func(key, value string)) {
	for key, value := range h.headers {
		cb(key, value)
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			read += len(rn)
			break
		}

		key, val, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isValidToken(key) {
			return 0, false, errMalformedHeader
		}

		h.Set(key, val)
		read += idx + len(rn)
	}
	return read, done, nil
}
