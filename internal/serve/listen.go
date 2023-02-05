// Package server
//
// @author: xwc1125
package serve

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chain5j/logger"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"github.com/xwc1125/apisix-go/internal/apisix/core/storage"
	"github.com/xwc1125/apisix-go/internal/apisix/core/store"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
	"github.com/xwc1125/apisix-go/internal/apisix/utils/iputils"
	"github.com/xwc1125/apisix-go/internal/apisix/utils/reg_uri"
	"github.com/xwc1125/apisix-go/internal/models"
	"github.com/xwc1125/apisix-go/internal/proxy"
)

type ProxyServe struct {
	log logger.Logger

	schema    gjson.Result
	confCache *plugins.ConfCache
	testLocal bool
}

func NewProxyServe() (*ProxyServe, error) {
	confCache := plugins.InitConfCache(time.Minute * 60)
	p := &ProxyServe{
		log:       logger.Log("proxy"),
		confCache: confCache,
		testLocal: viper.GetBool("test_local"),
	}
	var etcdConfig storage.EtcdConfig
	err := viper.UnmarshalKey("etcd", &etcdConfig)
	if err != nil {
		p.log.Error("unmarshal etcd config err", "err", err)
		return nil, err
	}
	err = storage.InitETCDClient(&etcdConfig)
	if err != nil {
		p.log.Error("init etcd client err", "err", err)
		return nil, err
	}

	p.initSchema(".")
	err = store.InitStores(p.schema, etcdConfig, map[store.HubKey]store.WatchEvent{
		store.HubKeyRoute: NewWatchRoute(confCache),
	})
	if err != nil {
		p.log.Error("init stores err", "err", err)
		return nil, err
	}
	return p, nil
}

var (
	testLocal = true
)

// ProxyHandler ...
func (p *ProxyServe) ProxyHandler(ctx *fasthttp.RequestCtx) {
	var route = new(entity.Route)
	if p.testLocal {
		// 本地测试
		bytes, err := os.ReadFile("conf/test-route.json")
		if err != nil {
			logger.Error("read json err", "err", err)
			return
		}
		err = json.Unmarshal(bytes, route)
		if err != nil {
			logger.Error("unmarshal route err", "err", err)
			return
		}
	} else {
		routeStore := store.GetStore(store.HubKeyRoute)
		output, err := routeStore.List(ctx, store.ListInput{
			Predicate: func(obj interface{}) bool {
				route, ok := obj.(*entity.Route)
				if !ok {
					return false
				}
				return MatchRoute(ctx, route)
			},
			Format: func(obj interface{}) interface{} {
				route, ok := obj.(*entity.Route)
				if !ok {
					return obj
				}
				if route.Upstream != nil && route.Upstream.Nodes != nil {
					route.Upstream.Nodes = entity.NodesFormat(route.Upstream.Nodes)
				}
				return route
			},
			Less: func(i, j interface{}) bool {
				routeI, ok := i.(*entity.Route)
				routeJ, ok1 := j.(*entity.Route)
				if !ok || !ok1 {
					return true
				}
				if routeI.Priority < routeJ.Priority {
					return false
				}
				if routeI.CreateTime < routeJ.CreateTime {
					return false
				}
				if routeI.UpdateTime < routeJ.UpdateTime {
					return false
				}
				return true
			},
			PageSize:   10,
			PageNumber: 0,
		})
		if err != nil {
			ctx.Error("parse uri err", http.StatusInternalServerError)
			return
		}
		if len(output.Rows) == 0 {
			ctx.Error(models.Response{}.SetErrMsg("404 Route Not Found").String(), http.StatusNotFound)
			return
		}
		logger.Debug("route match len", "len", len(output.Rows))
		route = output.Rows[0].(*entity.Route)
	}
	if route == nil {
		ctx.Error("Not found", fasthttp.StatusNotFound)
		return
	}
	newProxy, err := proxy.NewProxy(*route)
	if err != nil {
		ctx.Error("New proxy err:"+err.Error(), 500)
		return
	}
	newProxy.ServeHTTP(ctx)
}

func MatchRoute(ctx *fasthttp.RequestCtx, route *entity.Route) bool {
	// status
	if route.Status != 1 {
		return false
	}
	// uri
	{
		reqUri := string(ctx.RequestURI())
		if len(route.Uris) > 0 {
			match := false
			for _, uri := range route.Uris {
				if reg_uri.KeyMatch4(reqUri, uri) {
					match = true
					break
				}
			}
			if !match {
				return false
			}
		}
		if len(route.URI) > 0 && !reg_uri.KeyMatch4(reqUri, route.URI) {
			return false
		}
	}
	// host
	{
		reqHost := string(ctx.Host())
		split := strings.Split(reqHost, ":")
		domain := split[0]
		if len(route.Hosts) > 0 {
			match := false

			for _, host := range route.Hosts {
				if reg_uri.DomainMatch(domain, host) {
					match = true
					break
				}
			}
			if !match {
				return false
			}
		}
		if len(route.Host) > 0 && !reg_uri.DomainMatch(domain, route.Host) {
			return false
		}
	}
	// remoteAddr
	{
		remoteAddr := ctx.RemoteAddr().String()
		split := strings.Split(remoteAddr, ":")
		remoteIp := split[0]
		if len(route.RemoteAddrs) > 0 {
			match := false
			for _, addr := range route.RemoteAddrs {
				if iputils.Match(remoteIp, addr) {
					match = true
					break
				}
			}
			if !match {
				return false
			}
		}
		if len(route.RemoteAddr) > 0 && !iputils.Match(remoteIp, route.RemoteAddr) {
			return false
		}
	}
	// method
	{
		reqMethod := string(ctx.Method())
		if len(route.Methods) > 0 {
			match := false
			for _, method := range route.Methods {
				if strings.EqualFold(reqMethod, method) {
					match = true
					break
				} else if strings.EqualFold(method, "ALL") {
					match = true
					break
				}
			}
			if !match {
				return false
			}
		}
	}
	return true
}

var (
	_ store.WatchEvent = new(WatchRoute)
)

type WatchRoute struct {
	confCache *plugins.ConfCache
}

func NewWatchRoute(confCache *plugins.ConfCache) *WatchRoute {
	return &WatchRoute{
		confCache: confCache,
	}
}

func (w *WatchRoute) WatchEventPut(key string, objPtr interface{}) {
	w.confCache.Delete(key)
}

func (w *WatchRoute) WatchEventDelete(key string) {
	w.confCache.Delete(key)
}
