// Package plugins
//
// @author: xwc1125
package plugins

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(Gzip)
)

func init() {
	err := plugins.RegisterPlugin(&Gzip{
		name:     "gzip",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin Gzip", "err", err)
	}
}

type Gzip struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type GzipConf struct {
	Disable           bool   `json:"disable"`
	HeaderName        string `json:"header_name,omitempty" default:"X-Request-Id"` // 唯一请求ID的标头名称
	IncludeInResponse bool   `json:"include_in_response,omitempty" default:"true"` // true时，在响应头中添加唯一的请求ID。
	Algorithm         string `json:"algorithm,omitempty" default:"uuid"`           // ["uuid", "snowflake", "nanoid"]
}

func (p *Gzip) Name() string {
	return p.name
}

func (p *Gzip) Version() string {
	return p.version
}

func (p *Gzip) Priority() int64 {
	return p.priority
}

func (p *Gzip) ParseConf(in []byte) (interface{}, error) {
	conf := GzipConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *Gzip) ResponseFilter(conf interface{}, w *fasthttp.Response) (err error) {
	config, ok := conf.(GzipConf)
	if !ok {
		return fmt.Errorf("convert to Gzip conf err")
	}
	_ = config

	// head := ctx.Request().Header().GetHeader("Accept-Encoding")
	// if err == nil && strings.Contains(head, "gzip") {
	//
	// res, err := g.compress(resp.GetBody())
	//		if err != nil {
	//			return err
	//		}
	//		resp.SetBody(res)
	//		resp.SetHeader("Content-Encoding", "gzip")
	//		if g.conf.Vary {
	//			resp.SetHeader("Vary", "Accept-Encoding")
	//		}
	// }

	return nil
}

func (p *Gzip) compress(content []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(content)
	if err != nil {
		return nil, err
	}
	err = zw.Flush()
	if err != nil {
		return nil, err
	}
	err = zw.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
