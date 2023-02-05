// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"
	"path"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/store"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
	"github.com/xwc1125/apisix-go/internal/apisix/utils/reg_uri"
	"github.com/xwc1125/apisix-go/internal/xgateway"
)

var (
	_ plugins.Plugin = new(ProxyRewrite)
)

func init() {
	name := "proxy-rewrite"
	p := &ProxyRewrite{
		log:      logger.Log(name),
		name:     name,
		version:  "0.1",
		priority: 1008,
	}
	var err error
	if p.validator, err = store.NewSchemaValidator(p.schema()); err != nil {
		p.log.Error(p.schema()+" new schema validator err", "err", err)
		return
	}
	if err = plugins.RegisterPlugin(p); err != nil {
		p.log.Error("failed to register plugin"+p.Name(), "err", err)
	}
}

type ProxyRewrite struct {
	log       logger.Logger
	validator store.Validator

	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type ProxyRewriteConf struct {
	Disable  bool              `json:"disable"`
	Scheme   string            `json:"scheme,omitempty"`    // 协议改写，http/https
	Method   string            `json:"method,omitempty"`    // 被改写的method
	Uri      string            `json:"uri,omitempty"`       // url改写
	Host     string            `json:"host,omitempty"`      // host进行改写
	Headers  map[string]string `json:"headers,omitempty"`   // 修改的header头，如果value为空时，代表删除该key
	RegexUri []string          `json:"regex_uri,omitempty"` // uri的正则
}

func (p *ProxyRewrite) Name() string {
	return p.name
}

func (p *ProxyRewrite) Version() string {
	return p.version
}

func (p *ProxyRewrite) Priority() int64 {
	return p.priority
}

func (p *ProxyRewrite) ParseConf(in []byte) (interface{}, error) {
	conf := ProxyRewriteConf{}
	err := json.Unmarshal(in, &conf)
	if err != nil {
		p.log.Error("json unmarshal conf err", "err", err)
		return nil, err
	}

	// Validate
	err = p.validator.Validate(conf)
	if err != nil {
		p.log.Error("validate conf err", "err", err)
		return nil, err
	}
	return conf, nil
}

func (p *ProxyRewrite) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(ProxyRewriteConf)
	if !ok {
		p.log.Warn(ErrConfConvert.Error())
		return ErrConfConvert
	}
	if config.Disable {
		return nil
	}

	uri := r.URI()
	// 业务处理
	if len(config.Scheme) > 0 {
		// 修改schema
		uri.SetScheme(config.Scheme)
	}
	if len(config.Method) > 0 {
		r.Header.SetMethod(config.Method)
	}
	if len(config.Uri) > 0 {
		uri.SetPath(config.Uri)
	} else if len(config.RegexUri) == 2 {
		requestURI := string(r.RequestURI())
		reg := config.RegexUri[0]  // 代表reg的所有请求
		temp := config.RegexUri[1] // 都转发到temp这个API的$1==*
		// 示例
		// path: /check/getIp
		// reg: ^/check/(.*)
		// temp: /check/rewrite/$1
		rule, err := reg_uri.NewRule(reg, temp)
		if err != nil {
			p.log.Error("reg compile err", "reg", reg, "err", err)
		} else {
			if rule.MatchString(requestURI) {
				newPath := rule.Replace(requestURI)
				newPath = path.Clean(newPath)
				p.log.Debug("rewrite request uri", "old", requestURI, "new", newPath)
				uri.SetPath(newPath)
				r.Header.Set(xgateway.HeaderXRewriteOriginURI, requestURI)
			}
		}
	}
	if len(config.Host) > 0 {
		uri.SetHost(config.Host)
	}

	if len(config.Headers) > 0 {
		for key, val := range config.Headers {
			if len(val) == 0 {
				r.Header.Del(val)
			} else {
				r.Header.Set(key, val)
			}
		}
	}
	return nil
}

func (p *ProxyRewrite) schema() string {
	return `
{
        "$comment": "this is a mark for our injected plugin schema",
        "minProperties": 1,
        "properties": {
          "disable": {
            "type": "boolean"
          },
          "headers": {
            "description": "new headers for request",
            "minProperties": 1,
            "type": "object"
          },
          "host": {
            "description": "new host for upstream",
            "pattern": "^[0-9a-zA-Z-.]+(:\\d{1,5})?$",
            "type": "string"
          },
          "method": {
            "description": "proxy route method",
            "enum": [
              "COPY",
              "DELETE",
              "GET",
              "HEAD",
              "LOCK",
              "MKCOL",
              "MOVE",
              "OPTIONS",
              "PATCH",
              "POST",
              "PROPFIND",
              "PUT",
              "TRACE",
              "UNLOCK"
            ],
            "type": "string"
          },
          "regex_uri": {
            "description": "new uri that substitute from client uri for upstream, lower priority than uri property",
            "items": {
              "description": "regex uri",
              "type": "string"
            },
            "maxItems": 2,
            "minItems": 2,
            "type": "array"
          },
          "scheme": {
            "description": "new scheme for upstream",
            "enum": [
              "http",
              "https"
            ],
            "type": "string"
          },
          "uri": {
            "description": "new uri for upstream",
            "maxLength": 4096,
            "minLength": 1,
            "pattern": "^\\/.*",
            "type": "string"
          }
        },
        "type": "object"
      }
`
}
