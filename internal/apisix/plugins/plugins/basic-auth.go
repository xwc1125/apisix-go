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
	_ plugins.Plugin = new(BasicAuth)
)

func init() {
	err := plugins.RegisterPlugin(&BasicAuth{
		name:     "basic-auth",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin BasicAuth", "err", err)
	}
}

type BasicAuth struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type BasicAuthConf struct {
	Disable bool `json:"disable"`
}

func (p *BasicAuth) Name() string {
	return p.name
}

func (p *BasicAuth) Version() string {
	return p.version
}

func (p *BasicAuth) Priority() int64 {
	return p.priority
}

func (p *BasicAuth) ParseConf(in []byte) (interface{}, error) {
	conf := BasicAuthConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *BasicAuth) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(BasicAuthConf)
	if !ok {
		return ErrConfConvert
	}
	_ = config

	return nil
}
