// Package plugins
//
// @author: xwc1125
package plugins

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chain5j/logger"
	"github.com/dgrijalva/jwt-go"
	"github.com/savsgio/gotils/strconv"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ plugins.Plugin = new(JwtAuth)
)
var (
	bearerLength           = len(`Bearer `)
	badAuthorizationHeader = fmt.Errorf(`bad authorization header`)
)

func init() {
	err := plugins.RegisterPlugin(&JwtAuth{
		name:     "jwt-auth",
		version:  "0.1",
		priority: 10000,
	})
	if err != nil {
		logger.Fatal("failed to register plugin JwtAuth", "err", err)
	}
}

type JwtAuth struct {
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64
}

type JwtAuthConf struct {
	Disable bool `json:"disable"`

	Secret            string `json:"secret"`
	TokenFormat       string `json:"token_format"`
	TokenHeader       string `json:"token_header"`
	ParsedTokenHeader string `json:"parsed_token_header"`
}

func (p *JwtAuth) Name() string {
	return p.name
}

func (p *JwtAuth) Version() string {
	return p.version
}

func (p *JwtAuth) Priority() int64 {
	return p.priority
}

func (p *JwtAuth) ParseConf(in []byte) (interface{}, error) {
	conf := JwtAuthConf{}
	err := json.Unmarshal(in, &conf)
	if err != nil {
		return nil, err
	}
	if conf.TokenFormat != `Bearer` && conf.TokenFormat != `Custom` {
		return nil, fmt.Errorf(`must specify header format "Bearer" or "Custom"`)
	}

	return conf, err
}

func (p *JwtAuth) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(JwtAuthConf)
	if !ok {
		return fmt.Errorf("convert to JwtAuth conf err")
	}
	tokenString := getToken(r, config)
	if tokenString == `` {
		w.SetStatusCode(fasthttp.StatusForbidden)
		return badAuthorizationHeader
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return config.Secret, nil
	})
	if err != nil || !token.Valid {
		w.SetStatusCode(fasthttp.StatusForbidden)
		return err
	}
	segment, _ := jwt.DecodeSegment(strings.Split(tokenString, ".")[1]) // todo: remove aloc
	r.Header.SetBytesV(config.ParsedTokenHeader, segment)
	return nil
}

func getToken(req *fasthttp.Request, config JwtAuthConf) string {
	switch config.TokenFormat {
	case `Bearer`:
		header := req.Header.Peek(`Authorization`)
		if len(header) < bearerLength {
			// length of "Bearer "
			return ``
		}
		return strconv.B2S(header[7:])
	case "Custom":
		return strconv.B2S(req.Header.Peek(config.TokenHeader))
	default:
		return ``
	}
}
