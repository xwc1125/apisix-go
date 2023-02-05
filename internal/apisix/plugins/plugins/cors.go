// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/store"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
	"github.com/xwc1125/apisix-go/internal/apisix/utils/reg_uri"
)

var (
	_ plugins.Plugin = new(Cors)
)

const (
	MatchAllTag      = "*"
	MatchAllForceTag = "**"
)

func init() {
	p := &Cors{
		log:      logger.Log("cors"),
		name:     "cors",
		version:  "0.1",
		priority: 4000,
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

type Cors struct {
	log       logger.Logger
	validator store.Validator

	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type CorsConf struct {
	Disable bool `json:"disable"`
	// AllowOrigins 允许跨域访问的 Origin。格式为 scheme://host:port，示例如 https://somedomain.com:8081
	// 当 allow_credential 为 false 时，
	// 可以使用 * 来表示允许所有 Origin 通过。
	// 你也可以在启用了 allow_credential 后使用 ** 强制允许所有 Origin 均通过，
	// 但请注意这样存在安全隐患。
	AllowOrigins []string `json:"allow_origins,omitempty"`
	// AllowMethods 允许跨域访问的 Method，比如：GET，POST 等。
	// 当 allow_credential 为 false 时，
	// 可以使用 * 来表示允许所有 Method 通过。
	// 你也可以在启用了 allow_credential 后使用 ** 强制允许所有 Method 都通过，
	// 但请注意这样存在安全隐患。
	AllowMethods []string `json:"allow_methods,omitempty"` // 允许跨域访问的 Method
	// AllowHeaders 允许跨域访问时请求方携带哪些非 CORS 规范 以外的 Header。
	// 当 allow_credential 为 false 时，
	// 可以使用 * 来表示允许所有 Header 通过。
	// 你也可以在启用了 allow_credential 后使用 ** 强制允许所有 Header 都通过，
	// 但请注意这样存在安全隐患。
	AllowHeaders []string `json:"allow_headers,omitempty"`
	// ExposeHeaders 允许跨域访问时响应方携带哪些非 CORS 规范 以外的 Header。
	// 当 allow_credential 为 false 时，
	// 可以使用 * 来表示允许任意 Header 。
	// 你也可以在启用了 allow_credential 后使用 ** 强制允许任意 Header，
	// 但请注意这样存在安全隐患。
	ExposeHeaders []string `json:"expose_headers,omitempty"`
	// MaxAge 浏览器缓存 CORS 结果的最大时间，单位为秒。
	// 在这个时间范围内，浏览器会复用上一次的检查结果，-1 表示不缓存。
	// 请注意各个浏览器允许的最大时间不同
	MaxAge uint64 `json:"max_age,omitempty" default:"5"`
	// AllowCredential 是否允许跨域访问的请求方携带凭据（如 Cookie 等）。
	// 根据 CORS 规范，如果设置该选项为 true，
	// 那么将不能在其他属性中使用 *。
	AllowCredential bool `json:"allow_credential,omitempty"`
	// AllowOriginsByRegex 使用正则表达式数组来匹配允许跨域访问的 Origin，
	// 如 [".*\.test.com"] 可以匹配任何 test.com 的子域名 *。
	AllowOriginsByRegex []string `json:"allow_origins_by_regex,omitempty"`
	// AllowOriginsByMetadata 通过引用插件元数据的 allow_origins 配置允许跨域访问的 Origin。
	// 比如当插件元数据为 "allow_origins": {"EXAMPLE": "https://example.com"} 时，
	// 配置 ["EXAMPLE"] 将允许 Origin https://example.com 的访问。
	AllowOriginsByMetadata []string `json:"allow_origins_by_metadata,omitempty"`
}

func (p *Cors) Name() string {
	return p.name
}

func (p *Cors) Version() string {
	return p.version
}

func (p *Cors) Priority() int64 {
	return p.priority
}

func (p *Cors) ParseConf(in []byte) (interface{}, error) {
	conf := CorsConf{}
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
	return conf, err
}

func (p *Cors) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(CorsConf)
	if !ok {
		p.log.Warn(ErrConfConvert.Error())
		return ErrConfConvert
	}
	if config.Disable {
		return nil
	}
	// 业务逻辑
	if string(r.Header.Method()) == fasthttp.MethodOptions {
		p.handlePreflight(&config, r, w)
		w.SetStatusCode(200)
	} else {
		p.handleActual(&config, r, w)
	}
	return nil
}

func (p *Cors) schema() string {
	return `
 {
  "$comment": "this is a mark for our injected plugin schema",
  "properties": {
    "allow_credential": {
      "default": false,
      "description": "allow client append credential. according to CORS specification,if you set this option to 'true', you can not use '*' for other options.",
      "type": "boolean"
    },
    "allow_headers": {
      "default": "*",
      "description": "you can use '*' to allow all header when no credentials,'**' to allow forcefully(it will bring some security risks, be carefully),multiple header use ',' to split. default: *.",
      "type": "string"
    },
    "allow_methods": {
      "default": "*",
      "description": "you can use '*' to allow all methods when no credentials,'**' to allow forcefully(it will bring some security risks, be carefully),multiple method use ',' to split. default: *.",
      "type": "string"
    },
    "allow_origins": {
      "default": "*",
      "description": "you can use '*' to allow all origins when no credentials,'**' to allow forcefully(it will bring some security risks, be carefully),multiple origin use ',' to split. default: *.",
      "pattern": "^(\\*|\\*\\*|null|\\w+://[^,]+(,\\w+://[^,]+)*)$",
      "type": "string"
    },
    "allow_origins_by_metadata": {
      "description": "set allowed origins by referencing origins in plugin metadata",
      "items": {
        "maxLength": 4096,
        "minLength": 1,
        "type": "string"
      },
      "minItems": 1,
      "type": "array",
      "uniqueItems": true
    },
    "allow_origins_by_regex": {
      "description": "you can use regex to allow specific origins when no credentials,for example use [.*\\.test.com] to allow a.test.com and b.test.com",
      "items": {
        "maxLength": 4096,
        "minLength": 1,
        "type": "string"
      },
      "minItems": 1,
      "type": "array",
      "uniqueItems": true
    },
    "disable": {
      "type": "boolean"
    },
    "expose_headers": {
      "default": "*",
      "description": "you can use '*' to expose all header when no credentials,'**' to allow forcefully(it will bring some security risks, be carefully),multiple header use ',' to split. default: *.",
      "type": "string"
    },
    "max_age": {
      "default": 5,
      "description": "maximum number of seconds the results can be cached.-1 means no cached, the max value is depend on browser,more details plz check MDN. default: 5.",
      "type": "integer"
    }
  },
  "type": "object"
}
`
}
func (p *Cors) metadataSchema() string {
	return `
 {
  "properties": {
    "allow_origins": {
      "additionalProperties": {
        "pattern": "^(\\*|\\*\\*|null|\\w+://[^,]+(,\\w+://[^,]+)*)$",
        "type": "string"
      },
      "type": "object"
    }
  },
  "type": "object"
}
`
}

func (p *Cors) handlePreflight(config *CorsConf, r *fasthttp.Request, w *fasthttp.Response) {
	originHeader := string(r.Header.Peek("Origin"))
	if len(originHeader) == 0 || p.isAllowedOrigin(config, originHeader) == false {
		p.log.Warn("Origin ", originHeader, " is not in", config.AllowOrigins)
		return
	}
	method := string(r.Header.Peek("Access-Control-Request-Method"))
	if !p.isAllowedMethod(config, method) {
		p.log.Warn("Method ", method, " is not in", config.AllowMethods)
		return
	}
	headers := []string{}
	if len(r.Header.Peek("Access-Control-Request-Headers")) > 0 {
		headers = strings.Split(string(r.Header.Peek("Access-Control-Request-Headers")), ",")
	}
	if !p.isHeadersAllowed(config, headers) {
		p.log.Warn("Headers ", headers, " is not in", config.AllowMethods)
		return
	}

	w.Header.Set("Access-Control-Allow-Origin", originHeader)
	w.Header.Set("Access-Control-Allow-Methods", method)
	if len(headers) > 0 {
		w.Header.Set("Access-Control-Allow-Headers", strings.Join(headers, ", "))
	}
	if config.AllowCredential {
		w.Header.Set("Access-Control-Allow-Credentials", "true")
	}
	if config.MaxAge > 0 {
		w.Header.Set("Access-Control-Max-Age", strconv.Itoa(int(config.MaxAge)))
	}
}

func (p *Cors) handleActual(config *CorsConf, r *fasthttp.Request, w *fasthttp.Response) {
	originHeader := string(r.Header.Peek("Origin"))
	if len(originHeader) == 0 || p.isAllowedOrigin(config, originHeader) == false {
		p.log.Warn("Origin ", originHeader, " is not in", config.AllowOrigins)
		return
	}
	w.Header.Set("Access-Control-Allow-Origin", originHeader)
	if len(config.ExposeHeaders) > 0 {
		w.Header.Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
	}
	if config.AllowCredential {
		w.Header.Set("Access-Control-Allow-Credentials", "true")
	}
}

func (p *Cors) isAllowedOrigin(config *CorsConf, originHeader string) bool {
	// AllowOriginsByRegex
	if len(config.AllowOriginsByRegex) > 0 {
		for _, regex := range config.AllowOriginsByRegex {
			if reg_uri.DomainMatch(originHeader, regex) {
				return true
			}
		}
	}
	// allow origins
	for _, val := range config.AllowOrigins {
		if !config.AllowCredential && val == MatchAllTag {
			return true
		}
		if config.AllowCredential && val == MatchAllForceTag {
			return true
		}
		if val == originHeader {
			return true
		}
	}
	return false
}

func (p *Cors) isAllowedMethod(config *CorsConf, methodHeader string) bool {
	if methodHeader == "OPTIONS" {
		return true
	}
	for _, m := range config.AllowMethods {
		if !config.AllowCredential && m == MatchAllTag {
			return true
		}
		if config.AllowCredential && m == MatchAllForceTag {
			return true
		}
		if m == methodHeader {
			return true
		}
	}
	return false
}

func (p *Cors) isHeadersAllowed(config *CorsConf, headers []string) bool {
	if len(headers) == 0 {
		return true
	}
	for _, header := range config.AllowHeaders {
		if !config.AllowCredential && header == MatchAllTag {
			return true
		}
		if config.AllowCredential && header == MatchAllForceTag {
			return true
		}
	}

	for _, header := range headers {
		found := false
		for _, h := range config.AllowHeaders {
			if h == header {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
