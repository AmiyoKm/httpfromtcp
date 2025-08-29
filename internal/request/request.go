package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	b := make([]byte, 4096)
	n, err := reader.Read(b)
	if err != nil {
		return nil, err
	}
	line := string(b[:n])
	headers := strings.Split(line, "\r\n")

	parts := strings.Split(headers[0], " ")

	if len(parts) != 3 {
		return nil, fmt.Errorf("one or more of the part missing from request line header , parts leghth : %d", len(parts))
	}

	method := parts[0]
	if !isValidMethod(method) {
		return nil, fmt.Errorf("not a valid method , method : %s", method)
	}

	target := parts[1]
	if !isValidTarget(target) {
		return nil, fmt.Errorf("invalid request target: %s", target)
	}

	httpProtocol := strings.Split(parts[2], "/")
	version := httpProtocol[1]
	if !isValidHttpVersion(version) {
		return nil, fmt.Errorf("not a valid http version , verison : %s", version)
	}

	requestLine := RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}

	return &Request{
		RequestLine: requestLine,
	}, nil
}
