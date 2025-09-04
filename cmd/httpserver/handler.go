package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/AmiyoKm/httpfromtcp/internal/headers"
	"github.com/AmiyoKm/httpfromtcp/internal/request"
	"github.com/AmiyoKm/httpfromtcp/internal/response"
	"github.com/AmiyoKm/httpfromtcp/internal/server"
)

var handler = server.Handler(func(w *response.Writer, r *request.Request) {
	h := response.GetDefaultHeaders(0)
	h.Replace("Content-Type", "text/html")
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/stream") {
		target := r.RequestLine.RequestTarget
		res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
		if err != nil {
			body := respond500()

			w.WriteStatusLine(response.StatusInternalServerError)
			h.Replace("Content-Length", strconv.Itoa(len(body)))
			w.WriteHeaders(*h)
			w.WriteBody(body)
			return
		}
		w.WriteStatusLine(response.StatusOK)

		h.Delete("Content-Length")
		h.Set("transfer-encoding", "chunked")
		h.Replace("Content-Type", "text/plain")
		h.Set("Trailer", "X-Content-SHA256")
		h.Set("Trailer", "X-Content-Length")
		w.WriteHeaders(*h)

		fullBody := []byte{}
		for {
			data := make([]byte, 32)
			n, err := res.Body.Read(data)
			if err != nil {
				break
			}

			fullBody = append(fullBody, data[:n]...)
			w.WriteChunkedBody(data[:n])
		}

		w.WriteChunkedBodyDone()

		trailers := headers.NewHeaders()
		encrypted := sha256.Sum256(fullBody)
		trailers.Set("X-Content-SHA256", toStr(encrypted[:]))
		trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

		w.WriteHeaders(*trailers)

		return
	}

	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		body := respond400()
		w.WriteStatusLine(response.StatusBadRequest)
		h.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeaders(*h)
		w.WriteBody(body)
	case "/myproblem":
		body := respond500()

		w.WriteStatusLine(response.StatusInternalServerError)
		h.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeaders(*h)
		w.WriteBody(body)
	default:
		body := respond200()
		w.WriteStatusLine(response.StatusOK)
		h.Replace("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeaders(*h)
		w.WriteBody(body)
	}
})
