package client

import (
	"net"
	"net/http"
	"net/http/httputil"
)

func NewUnixTransport(socketPath string) *http.Transport {
	unixTransport := &http.Transport{}
	unixTransport.RegisterProtocol("unix", NewUnixRoundTripper(socketPath))

	return unixTransport
}

func NewUnixRoundTripper(path string) *UnixRoundTripper {
	return &UnixRoundTripper{path: path}
}

type UnixRoundTripper struct {
	path string
	conn httputil.ClientConn
}

func (rt *UnixRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	conn, err := net.Dial("unix", rt.path)
	if err != nil {
		return nil, err
	}

	socket := httputil.NewClientConn(conn, nil)
	defer socket.Close()

	return socket.Do(req)
}
