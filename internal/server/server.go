package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/AmiyoKm/httpfromtcp/internal/request"
	"github.com/AmiyoKm/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

// Serve creates a new server listening on the specified port
func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}
	server.closed.Store(false)

	// Start listening in a goroutine
	go server.listen()

	return server, nil
}

// Close shuts down the server by closing the listener
func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

// listen accepts connections and handles them in separate goroutines
func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		// Handle each connection in a separate goroutine
		go s.handle(conn)
	}
}

// handle processes a single connection
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Connection received, reading request...")
	headers := response.GetDefaultHeaders(0)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerErr := s.handler(writer, req)

	var body []byte
	var status response.StatusCode = response.StatusOK
	if handlerErr != nil {
		status = handlerErr.StatusCode
		body = []byte(handlerErr.Message)
	} else {
		body = writer.Bytes()
	}

	response.WriteStatusLine(conn, status)
	headers.Replace("Content-Length", strconv.Itoa(len(body)))
	response.WriteHeaders(conn, headers)

	conn.Write(body)
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError
