// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"
	"fmt"
	"go/types"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(MultiResponseRewrite)
)

func init() {
	err := plugins.RegisterPlugin(&MultiResponseRewrite{
		name:     "multi-response-rewrite",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin MultiResponseRewrite", "err", err)
	}
}

type MultiResponseRewrite struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type MultiResponseRewriteConf struct {
	Disable  bool                                    `json:"disable"`
	Variants []PluginMultiResponseRewriteVariantType `json:"variants"`
}

type PluginMultiResponseRewriteVariantType struct {
	StatusCode int       `json:"status_code"`
	Body       string    `json:"body"`
	BodyBase64 bool      `json:"body_base64"`
	Headers    types.Map `json:"headers"`
	Vars       string    `json:"vars"`
}

func (p *MultiResponseRewrite) Name() string {
	return p.name
}

func (p *MultiResponseRewrite) Version() string {
	return p.version
}

func (p *MultiResponseRewrite) Priority() int64 {
	return p.priority
}

func (p *MultiResponseRewrite) ParseConf(in []byte) (interface{}, error) {
	conf := MultiResponseRewriteConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *MultiResponseRewrite) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(MultiResponseRewriteConf)
	if !ok {
		return fmt.Errorf("convert to MultiResponseRewrite conf err")
	}
	_ = config

	return nil
}
