package eval

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

func newHTTPClient() *http.Client {
	// N.B. We need to use HTTP2 if we want to support bidirectional streaming
	//http.DefaultClient,
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				// Use the standard Dial function to create a plain TCP connection
				return net.Dial(network, addr)
			},
		},
	}
}
