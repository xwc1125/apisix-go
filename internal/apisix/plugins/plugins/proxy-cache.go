// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(ProxyCache)
)

func init() {
	err := plugins.RegisterPlugin(&ProxyCache{
		name:     "proxy-cache",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin ProxyCache", "err", err)
	}
}

type ProxyCache struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type ProxyCacheConf struct {
	Disable          bool     `json:"disable"`
	CacheStrategy    string   `json:"cache_strategy"`
	CacheZone        string   `json:"cache_zone"`
	CacheKey         []string `json:"cache_key"`
	CacheBypass      []string `json:"cache_bypass"`
	CacheMethod      []string `json:"cache_method"`
	CacheHTTPStatus  []string `json:"cache_http_status"`
	HideCacheHeaders bool     `json:"hide_cache_headers"`
	CacheControl     bool     `json:"cache_control"`
	NoCache          []string `json:"no_cache"`
	CacheTTL         int      `json:"cache_ttl"`
}

func (p *ProxyCache) Name() string {
	return p.name
}

func (p *ProxyCache) Version() string {
	return p.version
}

func (p *ProxyCache) Priority() int64 {
	return p.priority
}

func (p *ProxyCache) ParseConf(in []byte) (interface{}, error) {
	conf := ProxyCacheConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *ProxyCache) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(ProxyCacheConf)
	if !ok {
		return ErrConfConvert
	}
	_ = config

	return nil
}
