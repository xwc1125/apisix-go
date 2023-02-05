// Package fasthttp
//
// @author: xwc1125
package fasthttp

import (
	"context"
	"net"
	"net/url"

	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/xgateway"
)

var (
	_ xgateway.Request = new(Request)
)

type Request struct {
	ctx *fasthttp.RequestCtx
	*fasthttp.Request
}

func (r *Request) ID() uint64 {
	// TODO implement me
	panic("implement me")
}

func (r *Request) Context() context.Context {
	// TODO implement me
	panic("implement me")
}

func (r *Request) Header() xgateway.Header {
	// TODO implement me
	panic("implement me")
}

func (r *Request) Body() ([]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (r *Request) Cookies() []xgateway.Cookie {
	// TODO implement me
	panic("implement me")
}

func (r *Request) Cookie(name string) (xgateway.Cookie, error) {
	// TODO implement me
	panic("implement me")
}

func (r *Request) AddCookie(c xgateway.Cookie) {
	// TODO implement me
	panic("implement me")
}

func (r *Request) RemoteIP() net.IP {
	// TODO implement me
	panic("implement me")
}

func (r *Request) Method() string {
	// TODO implement me
	panic("implement me")
}

func (r *Request) Path() []byte {
	// TODO implement me
	panic("implement me")
}

func (r *Request) SetPath(bytes []byte) {
	// TODO implement me
	panic("implement me")
}

func (r *Request) Args() url.Values {
	// TODO implement me
	panic("implement me")
}

func NewRequest(ctx *fasthttp.RequestCtx) *Request {
	r := &Request{
		ctx:     ctx,
		Request: fasthttp.AcquireRequest(),
	}
	ctx.Request.CopyTo(r.Request)
	return r
}
