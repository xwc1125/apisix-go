package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/chain5j/logger"
	"github.com/chain5j/logger/zap"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
	_ "github.com/xwc1125/apisix-go/internal/apisix/plugins/plugins"
	"github.com/xwc1125/apisix-go/internal/proxy"
)

// ProxyHandler ...
func ProxyHandler(ctx *fasthttp.RequestCtx) {
	requestURI := string(ctx.RequestURI())

	routerJson, err := os.ReadFile("cmd/proxy/example/example_route.json")
	if err != nil {
		logger.Fatal(err)
	}
	var route entity.Route
	if err := json.Unmarshal([]byte(routerJson), &route); err != nil {
		logger.Error("unmarshal route json err", "err", err)
		logger.Fatal(err)
		return
	}

	logger.Info("new http into", "requestURI", requestURI)
	newProxy, _ := proxy.NewProxy(route)
	newProxy.ServeHTTP(ctx)
}

func main() {
	log := zap.InitWithConfig(&logger.LogConfig{
		Console: logger.ConsoleLogConfig{
			Level:    4,
			Modules:  "*",
			ShowPath: false,
			UseColor: true,
			Console:  true,
		},
	})

	plugins.InitConfCache(time.Second * 600)
	if err := fasthttp.ListenAndServe(":8081", ProxyHandler); err != nil {
		log.Fatal(err)
	}
}
