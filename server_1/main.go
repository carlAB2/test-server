package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	connStartTimes sync.Map
)

type loggingHandler struct {
	handler http.Handler
}

func (h *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remoteAddr := r.RemoteAddr
	startTime, ok := connStartTimes.Load(remoteAddr)
	if ok {
		duration := time.Since(startTime.(time.Time))
		log.Printf("Request from %s, connection duration: %v", remoteAddr, duration)
	}

	h.handler.ServeHTTP(w, r)
}

type loggingListener struct {
	net.Listener
}

func (l *loggingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	remoteAddr := conn.RemoteAddr().String()
	connStartTimes.Store(remoteAddr, time.Now())
	log.Printf("New connection from %s", remoteAddr)

	return &loggingConn{Conn: conn, remoteAddr: remoteAddr}, nil
}

type loggingConn struct {
	net.Conn
	remoteAddr string
}

func (c *loggingConn) Close() error {
	startTime, ok := connStartTimes.Load(c.remoteAddr)
	if ok {
		duration := time.Since(startTime.(time.Time))
		log.Printf("Connection from %s closed after %v", c.remoteAddr, duration)
		connStartTimes.Delete(c.remoteAddr)
	}
	return c.Conn.Close()
}

// Simulate sending timeout
func handleSendTimeout(conn net.Conn) {
	defer conn.Close()
	log.Println("Connection established, simulating send timeout...")

	// Read from the connection to simulate receiving data but do not send ACK
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Printf("Error reading from connection: %v", err)
		return
	}

	// Simulate delay without sending a response
	time.Sleep(30 * time.Minute) // Adjust the time as needed
}

// Simulate receiving timeout
func handleReceiveTimeout(conn net.Conn) {
	defer conn.Close()
	log.Println("Connection established, simulating receive timeout...")

	// Read from the connection to simulate receiving data
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Printf("Error reading from connection: %v", err)
		return
	}

	// Simulate delay without sending a full response
	time.Sleep(30 * time.Minute) // Adjust the time as needed
}

func startTCPServer(addr string, handler func(net.Conn)) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}
	defer listener.Close()
	log.Printf("TCP server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handler(conn)
	}
}

func main() {
	handler := &loggingHandler{handler: http.DefaultServeMux}

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Create a custom listener that wraps net.Listener
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error creating listener: %v", err)
	}

	loggingLn := &loggingListener{Listener: ln}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	log.Println("Starting server on :8080")
	go startTCPServer(":8083", handleSendTimeout)   // Server for simulating send timeout
	go startTCPServer(":8082", handleReceiveTimeout) // Server for simulating receive timeout
	log.Fatal(server.Serve(loggingLn))
}
