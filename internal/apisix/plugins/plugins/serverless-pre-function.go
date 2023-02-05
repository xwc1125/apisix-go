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
	_ plugins.Plugin = new(ServerlessPreFunction)
)

func init() {
	err := plugins.RegisterPlugin(&ServerlessPreFunction{
		name:     "serverless-pre-function",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin ServerlessPreFunction", "err", err)
	}
}

type ServerlessPreFunction struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type ServerlessPreFunctionConf struct {
	Disable   bool     `json:"disable"`
	Phase     string   `json:"phase"`
	Functions []string `json:"functions"`
}

func (p *ServerlessPreFunction) Name() string {
	return p.name
}

func (p *ServerlessPreFunction) Version() string {
	return p.version
}

func (p *ServerlessPreFunction) Priority() int64 {
	return p.priority
}

func (p *ServerlessPreFunction) ParseConf(in []byte) (interface{}, error) {
	conf := ServerlessPreFunctionConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *ServerlessPreFunction) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(ServerlessPreFunctionConf)
	if !ok {
		return ErrConfConvert
	}
	_ = config

	return nil
}
