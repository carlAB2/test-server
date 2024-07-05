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
	log.Fatal(server.Serve(loggingLn))
}
