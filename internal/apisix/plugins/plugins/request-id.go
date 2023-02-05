// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"

	"github.com/chain5j/logger"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(RequestId)
)

func init() {
	err := plugins.RegisterPlugin(&RequestId{
		name:     "request-id",
		version:  "0.1",
		priority: 12015,
	})
	if err != nil {
		logger.Fatal("failed to register plugin RequestId", "err", err)
	}
}

type RequestId struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type RequestIdConf struct {
	Disable           bool   `json:"disable"`
	HeaderName        string `json:"header_name,omitempty" default:"X-Request-Id"` // 唯一请求ID的标头名称
	IncludeInResponse bool   `json:"include_in_response,omitempty" default:"true"` // true时，在响应头中添加唯一的请求ID。
	Algorithm         string `json:"algorithm,omitempty" default:"uuid"`           // ["uuid", "snowflake", "nanoid"]
}

func (p *RequestId) Name() string {
	return p.name
}

func (p *RequestId) Version() string {
	return p.version
}

func (p *RequestId) Priority() int64 {
	return p.priority
}

func (p *RequestId) ParseConf(in []byte) (interface{}, error) {
	conf := RequestIdConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *RequestId) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(RequestIdConf)
	if !ok {
		return ErrConfConvert
	}
	if config.Disable {
		return nil
	}
	if len(config.HeaderName) == 0 {
		config.HeaderName = `X-Request-Id`
	}
	r.Header.Set(config.HeaderName, uuid.New().String())

	return nil
}
