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
	_ plugins.Plugin = new(ServerlessPostFunction)
)

func init() {
	err := plugins.RegisterPlugin(&ServerlessPostFunction{
		name:     "serverless-post-function",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin ServerlessPostFunction", "err", err)
	}
}

type ServerlessPostFunction struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type ServerlessPostFunctionConf struct {
	Disable   bool     `json:"disable"`
	Phase     string   `json:"phase"`
	Functions []string `json:"functions"`
}

func (p *ServerlessPostFunction) Name() string {
	return p.name
}

func (p *ServerlessPostFunction) Version() string {
	return p.version
}

func (p *ServerlessPostFunction) Priority() int64 {
	return p.priority
}

func (p *ServerlessPostFunction) ParseConf(in []byte) (interface{}, error) {
	conf := ServerlessPostFunctionConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *ServerlessPostFunction) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(ServerlessPostFunctionConf)
	if !ok {
		return ErrConfConvert
	}
	_ = config

	return nil
}
