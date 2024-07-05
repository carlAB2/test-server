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
    conn, ok := r.Context().Value(http.LocalAddrContextKey).(net.Conn)
    if !ok {
        http.Error(w, "could not get connection", http.StatusInternalServerError)
        return
    }

    addr := conn.RemoteAddr().String()

    // Log the start time for this connection
    if _, exists := connStartTimes.LoadOrStore(addr, time.Now()); !exists {
        log.Printf("New connection from %s", addr)
    }

    h.handler.ServeHTTP(w, r)
}

func main() {
    handler := &loggingHandler{handler: http.DefaultServeMux}

    server := &http.Server{
        Addr:    ":8080",
        Handler: handler,
        ConnState: func(c net.Conn, cs http.ConnState) {
            addr := c.RemoteAddr().String()

            switch cs {
            case http.StateNew:
                log.Printf("New connection: %s", addr)
            case http.StateClosed:
                if startTime, ok := connStartTimes.Load(addr); ok {
                    duration := time.Since(startTime.(time.Time))
                    log.Printf("Connection from %s closed after %v", addr, duration)
                    connStartTimes.Delete(addr)
                }
            }
        },
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    })

    log.Println("Starting server on :8080")
    log.Fatal(server.ListenAndServe())
}
