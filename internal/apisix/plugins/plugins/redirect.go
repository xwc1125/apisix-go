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
	_ plugins.Plugin = new(Redirect)
)

func init() {
	err := plugins.RegisterPlugin(&Redirect{
		name:     "redirect",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin Redirect", "err", err)
	}
}

type Redirect struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type RedirectConf struct {
	Disable           bool     `json:"disable"`
	HTTPToHTTPS       bool     `json:"http_to_https"`
	URI               string   `json:"uri"`
	RegexUri          []string `json:"regex_uri"`
	RetCode           int      `json:"ret_code"`
	EncodeURI         bool     `json:"encode_uri"`
	AppendQueryString bool     `json:"append_query_string"`
}

func (p *Redirect) Name() string {
	return p.name
}

func (p *Redirect) Version() string {
	return p.version
}

func (p *Redirect) Priority() int64 {
	return p.priority
}

func (p *Redirect) ParseConf(in []byte) (interface{}, error) {
	conf := RedirectConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *Redirect) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(RedirectConf)
	if !ok {
		return ErrConfConvert
	}
	_ = config

	return nil
}
