// Package proxy_http
//
// @author: xwc1125
package proxy

import (
	"github.com/valyala/fasthttp"
)

// compress 设置压缩比例
func compress(level int) func(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
		return fasthttp.CompressHandlerBrotliLevel(handler, level, level)
	}
}
