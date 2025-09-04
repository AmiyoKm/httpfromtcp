package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/AmiyoKm/httpfromtcp/internal/request"
	"github.com/AmiyoKm/httpfromtcp/internal/response"
	"github.com/AmiyoKm/httpfromtcp/internal/server"
)

const port = 42069

var handler = server.Handler(func(w *response.Writer, r *request.Request) {
	headers := response.GetDefaultHeaders(0)
	headers.Replace("Content-Type", "text/html")
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/stream") {
		target := r.RequestLine.RequestTarget
		res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
		if err != nil {
			body := respond500()

			w.WriteStatusLine(response.StatusInternalServerError)
			headers.Replace("Content-Length", strconv.Itoa(len(body)))
			w.WriteHeaders(*headers)
			w.WriteBody(body)
			return
		}
		w.WriteStatusLine(response.StatusOK)

		headers.Delete("Content-Length")
		headers.Set("transfer-encoding", "chunked")
		headers.Replace("Content-Type", "text/plain")
		w.WriteHeaders(*headers)

		for {
			data := make([]byte, 32)
			n, err := res.Body.Read(data)
			if err != nil {
				break
			}
			w.WriteChunkedBody(data[:n])
		}
		w.WriteChunkedBodyDone()
		return
	}

	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		body := respond400()
		w.WriteStatusLine(response.StatusBadRequest)
		headers.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeaders(*headers)
		w.WriteBody(body)
	case "/myproblem":
		body := respond500()

		w.WriteStatusLine(response.StatusInternalServerError)
		headers.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeaders(*headers)
		w.WriteBody(body)
	default:
		body := respond200()
		w.WriteStatusLine(response.StatusOK)
		headers.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeaders(*headers)
		w.WriteBody(body)
	}
})

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
