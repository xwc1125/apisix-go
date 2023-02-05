package cgw

import (
	"encoding/json"
	"fmt"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/store"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(CgwInterfacePart)
)

var (
	ErrConfConvert = fmt.Errorf("convert to conf err")
)

func init() {
	p := &CgwInterfacePart{
		log:      logger.Log("cgw-interface-part"),
		name:     "cgw-interface-part",
		version:  "0.1",
		priority: 2801,
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

type CgwInterfacePart struct {
	log       logger.Logger
	validator store.Validator

	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type CgwInterfacePartConf struct {
	Disable    bool            `json:"disable"`
	Interfaces []InterfaceInfo `json:"interfaces"` // InterfaceName-->api info
}

type InterfaceInfo struct {
	InterfaceName string `json:"interface_name" comment:"InterfaceName"`
	DestUri       string `json:"dest_uri" comment:"目标URI"`
	Method        string `json:"method,omitempty" comment:"代理的方法"`
}

type InterfaceReq struct {
	InterfaceName string      `json:"InterfaceName" comment:"InterfaceName"`
	Data          interface{} `json:"Data,omitempty" comment:"业务内容"`
}

func (p *CgwInterfacePart) Name() string {
	return p.name
}

func (p *CgwInterfacePart) Version() string {
	return p.version
}

func (p *CgwInterfacePart) Priority() int64 {
	return p.priority
}

func (p *CgwInterfacePart) ParseConf(in []byte) (interface{}, error) {
	conf := CgwInterfacePartConf{}
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

func (p *CgwInterfacePart) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) (err error) {
	config, ok := conf.(CgwInterfacePartConf)
	if !ok {
		p.log.Warn(ErrConfConvert.Error())
		return ErrConfConvert
	}
	if config.Disable {
		return nil
	}

	// 业务处理
	body := r.Body()
	var reqInterface InterfaceReq
	err = json.Unmarshal(body, &reqInterface)
	if err != nil {
		p.log.Error("json unmarshal body err", "err", err)
		w.SetStatusCode(fasthttp.StatusBadRequest)
		return fmt.Errorf("invalid body")
	}
	for _, info := range config.Interfaces {
		if info.InterfaceName == reqInterface.InterfaceName {
			// 找到代理
			r.SetRequestURI(info.DestUri)
			if len(info.Method) > 0 {
				r.Header.SetMethod(info.Method)
			}
			if reqInterface.Data != nil {
				bytes, err := json.Marshal(reqInterface.Data)
				if err != nil {
					p.log.Error("json marshal req data err", "err", err)
					return err
				}
				r.SetBody(bytes)
			}
			return nil
		}
	}
	w.SetStatusCode(fasthttp.StatusNotFound)
	return fmt.Errorf("interface name no found")
}

func (p *CgwInterfacePart) schema() string {
	return `
{
  "$comment": "this is a mark for our injected plugin schema",
  "properties": {
    "disable": {
      "type": "boolean"
    },
    "interfaces": {
      "items": {
        "interface_name": {
          "type": "string"
        },
        "dest_uri": {
          "type": "string"
        },
        "method": {
          "type": "string"
        }
      },
      "required": [
        "dest_uri"
      ],
      "minItems": 1,
      "type": "array"
    }
  },
  "required": [
    "interfaces"
  ],
  "type": "object"
}
`
}
