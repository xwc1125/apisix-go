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
	_ plugins.Plugin = new(GelfUdpLogger)
)

func init() {
	err := plugins.RegisterPlugin(&GelfUdpLogger{
		name:     "gelf-udp-logger",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin GelfUdpLogger", "err", err)
	}
}

type GelfUdpLogger struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type GelfUdpLoggerConf struct {
	Disable         bool   `json:"disable"`
	Host            string `json:"host"`
	Port            uint64 `json:"port"`
	Timeout         uint64 `json:"timeout"`
	LoggerName      string `json:"name"`
	InactiveTimeout uint64 `json:"inactive_timeout"`
	BufferDuration  uint64 `json:"buffer_duration"`
	MaxRetryCount   uint64 `json:"max_retry_count"`
	RetryDelay      uint64 `json:"retry_delay"`
	IncludeReqBody  bool   `json:"include_req_body"`
	BatchMaxSize    uint64 `json:"batch_max_size"`
}

func (p *GelfUdpLogger) Name() string {
	return p.name
}

func (p *GelfUdpLogger) Version() string {
	return p.version
}

func (p *GelfUdpLogger) Priority() int64 {
	return p.priority
}
func (p *GelfUdpLogger) ParseConf(in []byte) (interface{}, error) {
	conf := GelfUdpLoggerConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *GelfUdpLogger) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(GelfUdpLoggerConf)
	if !ok {
		return fmt.Errorf("convert to GelfUdpLogger conf err")
	}
	_ = config

	return nil
}
