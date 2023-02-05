// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"
	"go/types"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(ResponseRewrite)
)

func init() {
	err := plugins.RegisterPlugin(&ResponseRewrite{
		name:     "response-rewrite",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin ResponseRewrite", "err", err)
	}
}

type ResponseRewrite struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type ResponseRewriteConf struct {
	Disable    bool      `json:"disable"`
	StatusCode int       `json:"status_code"`
	Body       string    `json:"body"`
	BodyBase64 bool      `json:"body_base64"`
	Headers    types.Map `json:"headers"`
	Vars       string    `json:"vars"`
}

func (p *ResponseRewrite) Name() string {
	return p.name
}

func (p *ResponseRewrite) Version() string {
	return p.version
}

func (p *ResponseRewrite) Priority() int64 {
	return p.priority
}

func (p *ResponseRewrite) ParseConf(in []byte) (interface{}, error) {
	conf := ResponseRewriteConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *ResponseRewrite) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(ResponseRewriteConf)
	if !ok {
		return ErrConfConvert
	}
	_ = config

	return nil
}
