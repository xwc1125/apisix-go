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
	_ plugins.Plugin = new(RedirectRegex)
)

func init() {
	err := plugins.RegisterPlugin(&RedirectRegex{
		name:     "redirect-regex",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin RedirectRegex", "err", err)
	}
}

type RedirectRegex struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type RedirectRegexConf struct {
	Disable  bool                             `json:"disable"`
	Variants []PluginRedirectRegexVariantType `json:"variants"`
}
type PluginRedirectRegexVariantType struct {
	Pattern    string `json:"pattern"`
	Replace    string `json:"replace"`
	StatusCode int    `json:"status_code"`
}

func (p *RedirectRegex) Name() string {
	return p.name
}

func (p *RedirectRegex) Version() string {
	return p.version
}

func (p *RedirectRegex) Priority() int64 {
	return p.priority
}

func (p *RedirectRegex) ParseConf(in []byte) (interface{}, error) {
	conf := RedirectRegexConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *RedirectRegex) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(RedirectRegexConf)
	if !ok {
		return ErrConfConvert
	}
	_ = config

	return nil
}
