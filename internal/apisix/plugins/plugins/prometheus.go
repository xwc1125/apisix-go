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
	_ plugins.Plugin = new(Prometheus)
)

func init() {
	err := plugins.RegisterPlugin(&Prometheus{
		name:     "prometheus",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin Prometheus", "err", err)
	}
}

type Prometheus struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type PrometheusConf struct {
	Disable    bool `json:"disable"`
	PreferName bool `json:"prefer_name"` // true:打印Route/Service名称，而不是Prometheus中的ID

	Path             string    `json:"path"`
	Port             string    `json:"Port"`
	HistogramBuckets []float64 `json:"histogram_buckets"`
}

func (p *Prometheus) Name() string {
	return p.name
}

func (p *Prometheus) Version() string {
	return p.version
}

func (p *Prometheus) Priority() int64 {
	return p.priority
}

func (p *Prometheus) ParseConf(in []byte) (interface{}, error) {
	conf := PrometheusConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *Prometheus) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(PrometheusConf)
	if !ok {
		return ErrConfConvert
	}
	_ = config

	return nil
}
