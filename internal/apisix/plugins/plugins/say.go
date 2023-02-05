package plugins

import (
	"encoding/json"
	"fmt"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/store"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(Say)
)

func init() {
	name := "say"
	p := &Say{
		log:      logger.Log(name),
		name:     name,
		version:  "0.1",
		priority: 10000,
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

type Say struct {
	log       logger.Logger
	validator store.Validator

	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type SayConf struct {
	Disable bool   `json:"disable"`
	Body    string `json:"body"`
}

func (p *Say) Name() string {
	return p.name
}

func (p *Say) Version() string {
	return p.version
}

func (p *Say) Priority() int64 {
	return p.priority
}

func (p *Say) ParseConf(in []byte) (interface{}, error) {
	conf := SayConf{}
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

func (p *Say) ResponseFilter(conf interface{}, w *fasthttp.Response) error {
	config, ok := conf.(SayConf)
	if !ok {
		p.log.Warn(ErrConfConvert.Error())
		return ErrConfConvert
	}
	if config.Disable {
		return nil
	}

	// 业务处理
	w.Header.Add("X-Resp-A6-Runner", "Go")
	_, err := w.BodyWriter().Write([]byte(config.Body))
	if err != nil {
		logger.Error("failed to write", "err", err)
		return fmt.Errorf("failed to write")
	}

	return nil
}

func (p *Say) schema() string {
	return `
 {
  "$comment": "this is a mark for our injected plugin schema",
  "properties": {
    "disable": {
      "type": "boolean"
    },
    "body": {
      "maxLength": 1024,
      "minLength": 1,
      "type": "string"
    }
  },
  "type": "object"
}
`
}
