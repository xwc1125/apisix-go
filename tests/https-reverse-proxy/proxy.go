package main

import (
	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	proxy2 "github.com/xwc1125/apisix-go/internal/proxy"
)

var (
	proxyServer, _ = proxy2.NewProxy(entity.Route{},
		proxy2.WithTLS("tests/https-reverse-proxy/selfsigned.crt", "tests/https-reverse-proxy/selfsigned.key"))
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	requestURI := string(ctx.RequestURI())
	logger.Info("a request incoming", "requestURI", requestURI)
	proxyServer.ServeHTTP(ctx)
}

func main() {
	if err := fasthttp.ListenAndServe("0.0.0.0:8081", ProxyHandler); err != nil {
		logger.Fatal(err)
	}
}
