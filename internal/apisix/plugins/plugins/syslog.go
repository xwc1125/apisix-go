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
	_ plugins.Plugin = new(Syslog)
)

func init() {
	err := plugins.RegisterPlugin(&Syslog{
		name:     "syslog",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin Syslog", "err", err)
	}
}

type Syslog struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type SyslogConf struct {
	Disable      bool   `json:"disable"`
	Host         string `tfsdk:"host"`
	Port         int    `tfsdk:"port"`
	BatchMaxSize uint64 `tfsdk:"batch_max_size"`
}

func (p *Syslog) Name() string {
	return p.name
}

func (p *Syslog) Version() string {
	return p.version
}

func (p *Syslog) Priority() int64 {
	return p.priority
}

func (p *Syslog) ParseConf(in []byte) (interface{}, error) {
	conf := SyslogConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *Syslog) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(SyslogConf)
	if !ok {
		return ErrConfConvert
	}
	_ = config

	return nil
}
