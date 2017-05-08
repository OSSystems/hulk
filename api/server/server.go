package server

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

// NewListener creates a new listener
func NewListener(addr string) (net.Listener, error) {
	parts := strings.SplitN(addr, "://", 2)

	if len(parts) != 2 {
		return nil, errors.New("Invalid address")
	}

	return net.Listen(parts[0], parts[1])
}

// Listen listens on the address
func Listen(l net.Listener, handler http.Handler) error {
	server := &http.Server{
		Addr:    l.Addr().String(),
		Handler: handler,
	}

	return server.Serve(l)
}
