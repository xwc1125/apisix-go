package main

import (
	"log"

	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"github.com/xwc1125/apisix-go/internal/proxy"
	pool2 "github.com/xwc1125/apisix-go/internal/proxy/pool"
)

var (
	pool1 pool2.Pool
	err   error
)

// ProxyPoolHandler ...
func ProxyPoolHandler(ctx *fasthttp.RequestCtx) {
	proxyServer, err := pool1.Get("localhost:9090")
	if err != nil {
		log.Println("ProxyPoolHandler got an error: ", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	defer pool1.Put(proxyServer)
	proxyServer.ServeHTTP(ctx)
}

func factory(hostAddr string) (*proxy.Proxy, error) {
	p, _ := proxy.NewProxy(entity.Route{})
	return p, nil
}

func main() {
	initialCap, maxCap := 100, 1000
	pool1, err = pool2.NewChanPool(initialCap, maxCap, factory)
	if err := fasthttp.ListenAndServe(":8083", ProxyPoolHandler); err != nil {
		panic(err)
	}
}
