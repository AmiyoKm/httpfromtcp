package response

import (
	"fmt"
	"io"
	"log/slog"
	"strconv"

	"github.com/AmiyoKm/httpfromtcp/internal/headers"
)

type StatusCode int

type Response struct {
}
type Writer struct {
	writer io.Writer
}

func NewWriter(wc io.WriteCloser) *Writer {
	return &Writer{
		writer: wc,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	line := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase[statusCode])

	_, err := w.writer.Write([]byte(line))
	if err != nil {
		return err
	}
	return nil
}
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	b := []byte{}

	headers.ForEach(func(key, value string) {
		slog.Info("WRITE#HEADERS ", "key", key, "value", value)
		b = fmt.Appendf(b, "%s: %s\r\n", key, value)
	})

	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	return n, err
}

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var reasonPhrase = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

func NewResponse(staus StatusCode) *Response {
	return &Response{}
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}
