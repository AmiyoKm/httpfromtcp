package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/AmiyoKm/httpfromtcp/internal/request"
	"github.com/AmiyoKm/httpfromtcp/internal/response"
	"github.com/AmiyoKm/httpfromtcp/internal/server"
)

const port = 42069

var handler = server.Handler(func(w *response.Writer, r *request.Request) {
	headers := response.GetDefaultHeaders(0)
	headers.Replace("Content-Type", "text/html")
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		body := respond400()
		w.WriteStatusLine(response.StatusBadRequest)
		headers.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeaders(*headers)
		w.WriteBody(respond400())
	case "/myproblem":
		body := respond500()

		w.WriteStatusLine(response.StatusInternalServerError)
		headers.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeaders(*headers)
		w.WriteBody(respond500())
	default:
		body := respond200()
		w.WriteStatusLine(response.StatusOK)
		headers.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeaders(*headers)
		w.WriteBody(respond200())
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
