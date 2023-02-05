package main

import (
	"log"
	"sync"

	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"github.com/xwc1125/apisix-go/internal/proxy"
)

var (
	proxyServer *proxy.Proxy
	once        sync.Once
)

// ProxyHandler ... fasthttp.RequestHandler func
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	once.Do(func() {
		var err error
		proxyServer, _ = proxy.NewProxy(entity.Route{})
		if err != nil {
			panic(err)
		}
	})

	switch string(ctx.Path()) {
	case "/echo":
		proxyServer.ServeHTTP(ctx)
	case "/":
		fasthttp.ServeFileUncompressed(ctx, "./index.html")
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func main() {
	log.Println("serving on: 8081")
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
