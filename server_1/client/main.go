package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

// Function to measure time until connection is closed
func measureConnectionTime(address string, timeout time.Duration) {
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Sending request to the server
	request := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	_, err = conn.Write([]byte(request))
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Reading response from the server
	buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			break
		}
	}

	duration := time.Since(start)
	fmt.Printf("Connection to %s closed after %v\n", address, duration)
}

func main() {
	// Addresses of the servers to test
	addr := "localhost"
	connectTimeoutServer := addr + ":8081"
	sendTimeoutServer := addr + ":8082"
	receiveTimeoutServer := addr + ":8083"

	// Timeout for connecting to the server
	connectTimeout := 5 * time.Second

	// Measure connection time for connect timeout server
	fmt.Println("Testing connect timeout server...")
	measureConnectionTime(connectTimeoutServer, connectTimeout)

	// Measure connection time for send timeout server
	fmt.Println("Testing send timeout server...")
	measureConnectionTime(sendTimeoutServer, connectTimeout)

	// Measure connection time for receive timeout server
	fmt.Println("Testing receive timeout server...")
	measureConnectionTime(receiveTimeoutServer, connectTimeout)
}
