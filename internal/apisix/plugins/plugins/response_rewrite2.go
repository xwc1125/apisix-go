package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(ResponseRewrite2)
)

func init() {
	err := plugins.RegisterPlugin(&ResponseRewrite2{
		name:     "response-rewrite2",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin ResponseRewrite2", "err", err)
	}
}

type RegexFilter struct {
	Regex   string `json:"regex"`
	Scope   string `json:"scope"`
	Replace string `json:"replace"`

	regexComplied *regexp.Regexp
}

type ResponseRewrite2 struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type ResponseRewrite2Conf struct {
	Disable bool              `json:"disable"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Filters []RegexFilter     `json:"filters"`
}

func (p *ResponseRewrite2) Name() string {
	return p.name
}

func (p *ResponseRewrite2) Version() string {
	return p.version
}

func (p *ResponseRewrite2) Priority() int64 {
	return p.priority
}

func (p *ResponseRewrite2) ParseConf(in []byte) (interface{}, error) {
	conf := ResponseRewrite2Conf{}
	err := json.Unmarshal(in, &conf)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(conf.Filters); i++ {
		if reg, err := regexp.Compile(conf.Filters[i].Regex); err != nil {
			return nil, fmt.Errorf("failed to compile regex `%s`: %v",
				conf.Filters[i].Regex, err)
		} else {
			conf.Filters[i].regexComplied = reg
		}
	}
	return conf, nil
}

func (p *ResponseRewrite2) ResponseFilter(conf interface{}, w *fasthttp.Response) (err error) {
	cfg := conf.(ResponseRewrite2Conf)
	if cfg.Status > 0 {
		w.SetStatusCode(200)
	}

	w.Header.Set("X-Resp-A6-Runner", "Go")
	if len(cfg.Headers) > 0 {
		for k, v := range cfg.Headers {
			w.Header.Set(k, v)
		}
	}

	body := []byte(cfg.Body)

	if len(cfg.Filters) > 0 {
		originBody := w.Body()
		if err != nil {
			logger.Error("failed to read response body", "err", err)
			return
		}
		matched := false
		for i := 0; i < len(cfg.Filters); i++ {
			f := cfg.Filters[i]
			found := f.regexComplied.Find(originBody)
			if found != nil {
				matched = true
				if f.Scope == "once" {
					originBody = bytes.Replace(originBody, found, []byte(f.Replace), 1)
				} else if f.Scope == "global" {
					originBody = bytes.ReplaceAll(originBody, found, []byte(f.Replace))
				}
			}
		}
		if matched {
			body = originBody
			goto write
		}
		// When configuring the Filters field, the Body field will be invalid.
		return
	}

	if len(body) == 0 {
		return
	}
write:
	_, err = w.BodyWriter().Write([]byte(body))
	if err != nil {
		logger.Error("failed to write", "err", err)
	}
	return nil
}
