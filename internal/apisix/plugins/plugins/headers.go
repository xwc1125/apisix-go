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
	_ plugins.Plugin = new(Headers)
)

func init() {
	err := plugins.RegisterPlugin(&Headers{
		name:     "headers",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin Headers", "err", err)
	}
}

type Headers struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type HeadersConf struct {
	Disable  bool                  `json:"disable"`
	Request  types.Map             `json:"request"`
	Response types.Map             `json:"response"`
	STS      *PluginHeadersSTSType `json:"sts"`
}

type PluginHeadersSTSType struct {
	MaxAge            uint64 `json:"max_age"`
	IncludeSubDomains bool   `json:"include_sub_domains"`
	Preload           bool   `json:"preload"`
}

func (p *Headers) Name() string {
	return p.name
}

func (p *Headers) Version() string {
	return p.version
}

func (p *Headers) Priority() int64 {
	return p.priority
}

func (p *Headers) ParseConf(in []byte) (interface{}, error) {
	conf := HeadersConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *Headers) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(HeadersConf)
	if !ok {
		return fmt.Errorf("convert to Headers conf err")
	}
	_ = config

	return nil
}
