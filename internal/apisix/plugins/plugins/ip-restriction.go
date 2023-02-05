// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"
	"fmt"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/store"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
	"github.com/xwc1125/apisix-go/internal/apisix/utils/iputils"
)

var (
	_ plugins.Plugin = new(IpRestriction)
)
var (
	defaultMsgIpErr = fmt.Errorf("Your IP address is not allowed")
	ErrConfConvert  = fmt.Errorf("convert to conf err")
)

func init() {
	p := &IpRestriction{
		log:      logger.Log("ip-restriction"),
		name:     "ip-restriction",
		version:  "0.1",
		priority: 3000,
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

// IpRestriction ...
type IpRestriction struct {
	log       logger.Logger
	validator store.Validator

	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type IpRestrictionConf struct {
	Disable bool `json:"disable"`
	// Whitelist 加入白名单的 IP 地址或 CIDR 范围。
	Whitelist []string `json:"whitelist,omitempty"`
	// 加入黑名单的 IP 地址或 CIDR 范围。
	Blacklist []string `json:"blacklist,omitempty"`
	// Message 不允许访问IP地址时返回的消息[1, 1024]
	Message string `json:"message,omitempty"`

	msgErr error
}

func (p *IpRestriction) Name() string {
	return p.name
}

func (p *IpRestriction) Version() string {
	return p.version
}

func (p *IpRestriction) Priority() int64 {
	return p.priority
}

func (p *IpRestriction) ParseConf(in []byte) (interface{}, error) {
	conf := IpRestrictionConf{}
	err := json.Unmarshal(in, &conf)
	if err != nil {
		p.log.Error("json unmarshal conf err", "err", err)
		return nil, err
	}
	if len(conf.Message) > 0 {
		conf.msgErr = fmt.Errorf(conf.Message)
	} else {
		conf.msgErr = defaultMsgIpErr
	}
	// Validate
	err = p.validator.Validate(conf)
	if err != nil {
		p.log.Error("validate conf err", "err", err)
		return nil, err
	}
	return conf, nil
}

func (p *IpRestriction) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(IpRestrictionConf)
	if !ok {
		p.log.Warn(ErrConfConvert.Error())
		return ErrConfConvert
	}
	if config.Disable {
		return nil
	}
	remoteIp := iputils.ClientIP(r)

	if len(config.Blacklist) > 0 {
		for _, ip := range config.Blacklist {
			if iputils.Match(remoteIp, ip) {
				// 属于黑名单
				p.log.Debug("ip is black", "ip", ip)
				w.SetStatusCode(fasthttp.StatusForbidden)
				return config.msgErr
			}
		}
	}

	if len(config.Whitelist) > 0 {
		for _, ip := range config.Whitelist {
			if iputils.Match(remoteIp, ip) {
				// 属于白名单
				p.log.Debug("ip is white", "ip", ip)
				return nil
			}
		}
		// 默认不符合要求
		w.SetStatusCode(fasthttp.StatusForbidden)
		return config.msgErr
	}

	return nil
}

func (p *IpRestriction) schema() string {
	return `
 {
  "$comment": "this is a mark for our injected plugin schema",
  "oneOf": [
    {
      "required": [
        "whitelist"
      ]
    },
    {
      "required": [
        "blacklist"
      ]
    }
  ],
  "properties": {
    "blacklist": {
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
    },
    "disable": {
      "type": "boolean"
    },
    "message": {
      "default": "Your IP address is not allowed",
      "maxLength": 1024,
      "minLength": 1,
      "type": "string"
    },
    "whitelist": {
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
  "type": "object"
}
`
}
