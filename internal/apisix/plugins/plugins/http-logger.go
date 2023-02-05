// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"
	"fmt"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(HttpLogger)
)

func init() {
	err := plugins.RegisterPlugin(&HttpLogger{
		name:     "http-logger",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin HttpLogger", "err", err)
	}
}

type HttpLogger struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type HttpLoggerConf struct {
	Disable         bool   `json:"disable"`
	URI             string `json:"uri"`
	AuthHeader      string `json:"auth_header"`
	Timeout         uint64 `json:"timeout"`
	LoggerName      string `json:"name"`
	BatchMaxSize    uint64 `json:"batch_max_size"`
	InactiveTimeout uint64 `json:"inactive_timeout"`
	BufferDuration  uint64 `json:"buffer_duration"`
	MaxRetryCount   uint64 `json:"max_retry_count"`
	RetryDelay      uint64 `json:"retry_delay"`
	IncludeReqBody  bool   `json:"include_req_body"`
	ConcatMethod    string `json:"concat_method"`
}

func (p *HttpLogger) Name() string {
	return p.name
}

func (p *HttpLogger) Version() string {
	return p.version
}

func (p *HttpLogger) Priority() int64 {
	return p.priority
}

func (p *HttpLogger) ParseConf(in []byte) (interface{}, error) {
	conf := HttpLoggerConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *HttpLogger) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(HttpLoggerConf)
	if !ok {
		return fmt.Errorf("convert to HttpLogger conf err")
	}
	_ = config

	return nil
}
