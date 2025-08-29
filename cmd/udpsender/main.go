package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Println("could not find udp address:", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("could not dial udp address:", err)
		os.Exit(1)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("error cloding the udp address:", err)
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("could not read the input:", err)
		}

		n, err := conn.Write([]byte(line))
		if err != nil {
			fmt.Println("could not write the line", err)
		}
		fmt.Printf("wrote %d bytes\n", n)
	}

}
