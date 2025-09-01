package main

import (
	"fmt"
	"net"
	"os"

	"github.com/AmiyoKm/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println("could not start listener:", err)
		os.Exit(1)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("could not accept connection:", err)
			continue
		}
		fmt.Println("tcp connection accepted")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println("error :", err)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Printf("Headers: \n")
		req.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)
		})

		fmt.Println("tcp connection closed")
	}
}
