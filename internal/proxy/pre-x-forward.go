// Package proxy_http
//
// @author: xwc1125
package proxy

import (
	"net"

	"github.com/valyala/fasthttp"
)

// xForwardFor 设置X-Forwarded-For
func xForwardFor(ctx *fasthttp.RequestCtx, req *fasthttp.Request) {
	if ip, _, err := net.SplitHostPort(ctx.RemoteAddr().String()); err == nil {
		req.Header.Add("X-Forwarded-For", ip)
	}
}
