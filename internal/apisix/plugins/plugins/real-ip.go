// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/store"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(RealIP)
)

func init() {
	p := &RealIP{
		log:      logger.Log("real-ip"),
		name:     "real-ip",
		version:  "0.1",
		priority: 23000,
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

// RealIP 插件用于动态改变传递到 客户端的 IP 地址和端口
type RealIP struct {
	log       logger.Logger
	validator store.Validator

	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

// RealIPConf 和nginx的配置一致
//
//	http{
//	   #真实服务器上一级代理的IP地址或者IP段,可以写多行
//	   set_real_ip_from 127.0.0.1;
//	    #从哪个header头检索出要的IP地址
//	   real_ip_header X-Forwarded-For;
//	   # 递归排除IP地址,ip串从右到左开始排除set_real_ip_from里面出现的IP,如果出现了未出现这些ip段的IP，那么这个IP将被认为是用户的IP
//	   real_ip_recursive on;
//	}
type RealIPConf struct {
	Disable bool `json:"disable"`
	// Source 数据来源，如 arg_realip 或 http_x_forwarded_for,
	// 动态设置客户端的 IP 地址和端口。
	// 如果该值不包含端口，则不会更改客户端的端口。
	Source string `json:"source" validate:"required"`
	// TrustedAddresses IP 或 CIDR 范围列表
	// 动态设置 set_real_ip_from 字段
	TrustedAddresses []string `json:"trusted_addresses,omitempty"`
	// Recursive 如果禁用递归搜索，则与受信任地址之一匹配的原始客户端地址将替换为配置的source中发送的最后一个地址。
	// 如果启用递归搜索，则与受信任地址之一匹配的原始客户端地址将替换为配置的source中发送的最后一个非受信任地址。
	Recursive bool `json:"recursive,omitempty"`
}

func (p *RealIP) Name() string {
	return p.name
}

func (p *RealIP) Version() string {
	return p.version
}

func (p *RealIP) Priority() int64 {
	return p.priority
}

func (p *RealIP) ParseConf(in []byte) (interface{}, error) {
	conf := RealIPConf{}
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

func (p *RealIP) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(RealIPConf)
	if !ok {
		p.log.Warn(ErrConfConvert.Error())
		return ErrConfConvert
	}
	if config.Disable {
		return nil
	}

	return nil
}

func (p *RealIP) schema() string {
	return `
{
  "$comment": "this is a mark for our injected plugin schema",
  "properties": {
    "disable": {
      "type": "boolean"
    },
    "recursive": {
      "default": false,
      "type": "boolean"
    },
    "source": {
      "minLength": 1,
      "type": "string"
    },
    "trusted_addresses": {
      "items": {
        "anyOf": [
          {
            "format": "ipv4",
            "title": "IPv4",
            "type": "string"
          },
          {
            "pattern": "^([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])/([12]?[0-9]|3[0-2])$",
            "title": "IPv4/CIDR",
            "type": "string"
          },
          {
            "format": "ipv6",
            "title": "IPv6",
            "type": "string"
          },
          {
            "pattern": "^([a-fA-F0-9]{0,4}:){1,8}(:[a-fA-F0-9]{0,4}){0,8}([a-fA-F0-9]{0,4})?/[0-9]{1,3}$",
            "title": "IPv6/CIDR",
            "type": "string"
          }
        ]
      },
      "minItems": 1,
      "type": "array"
    }
  },
  "required": [
    "source"
  ],
  "type": "object"
}
`
}
