package main

import (
	"net"
	"net/http"
)

var (
	localIP = net.IPv4(127, 0, 0, 1)
)

// Middleware The type of our middleware consists of the original handler we want to wrap and a message
type Middleware struct {
	next   http.Handler
	signal chan string
}

// NewMiddleware Make a constructor for our middleware type since its fields are not exported (in lowercase)
func NewMiddleware(next http.Handler, signal chan string) *Middleware {
	return &Middleware{next: next, signal: signal}
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	userIP := net.ParseIP(ip)

	if userIP.Equal(localIP) {
		switch r.RequestURI {
		case "/service/stop/":
			middlewareServiceStop(w)
			m.signal <- "stop"
			return
		}

		m.next.ServeHTTP(w, r)
		return
	}

	m.next.ServeHTTP(w, r)
}

func middlewareServiceStop(w http.ResponseWriter) {
	serviceStatus.stop()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Сервис завершает работу…"))
}
