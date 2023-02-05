package pool

import (
	"errors"

	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/proxy"
)

var (
	// errClosed is the error resulting if the pool is closed via pool.Close().
	errClosed = errors.New("pool is closed")
)

// Proxier can be HTTP or WebSocket proxier
// TODO:
type Proxier interface {
	ServeHTTP(ctx *fasthttp.RequestCtx)
	// ?
	SetClient(addr string) Proxier

	// Reset .
	Reset()

	// Close .
	Close()
}

// Pool interface ...
// this interface ref to: https://github.com/fatih/pool/blob/master/pool.go
type Pool interface {
	// Get returns a new Proxy from the pool.
	Get(string) (*proxy.Proxy, error)

	// Put Reseting the Proxy puts it back to the Pool.
	Put(*proxy.Proxy) error

	// Close closes the pool and all its connections. After Close() the pool is
	// no longer usable.
	Close()

	// Len returns the current number of connections of the pool.
	Len() int
}
