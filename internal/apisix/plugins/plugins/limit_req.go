// Package limit_req
//
// @author: xwc1125
package plugins

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
	"golang.org/x/time/rate"
)

var (
	_ plugins.Plugin = new(LimitReq)
)

func init() {
	err := plugins.RegisterPlugin(&LimitReq{
		name:     "limit-req",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin limit-req", "err", err)
	}
}

type LimitReq struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type LimitReqConf struct {
	Burst int     `json:"burst"`
	Rate  float64 `json:"rate"`

	limiter *rate.Limiter
}

func (p *LimitReq) Name() string {
	return p.name
}

func (p *LimitReq) Version() string {
	return p.version
}

func (p *LimitReq) Priority() int64 {
	return p.priority
}

// ParseConf is called when the configuration is changed. And its output is unique per route.
func (p *LimitReq) ParseConf(in []byte) (interface{}, error) {
	conf := LimitReqConf{}
	err := json.Unmarshal(in, &conf)
	if err != nil {
		return nil, err
	}

	limiter := rate.NewLimiter(rate.Limit(conf.Rate), conf.Burst)
	// the conf can be used to store route scope data
	conf.limiter = limiter
	return conf, nil
}

// RequestFilter is called when a request hits the route
func (p *LimitReq) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	li := conf.(LimitReqConf).limiter
	rs := li.Reserve()
	if !rs.OK() {
		// limit rate exceeded
		logger.Info("limit req rate exceeded")
		// stop filters with this response
		w.SetStatusCode(fasthttp.StatusServiceUnavailable)
		return fmt.Errorf("limit req rate exceeded")
	}
	time.Sleep(rs.Delay())
	return nil
}
