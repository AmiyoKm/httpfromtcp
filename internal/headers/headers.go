package headers

import (
	"bytes"
	"errors"
)

type Headers map[string]string

var errMalformedHeader = errors.New("header is malformed")

var rn = []byte("\r\n")

func NewHeaders() Headers {
	return Headers{}
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

func (h Headers) Parse(data []byte) (int, bool, error) {
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

		h[key] = val
		read += idx + len(rn)
	}
	return read, done, nil
}
