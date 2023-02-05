// Package plugins
//
// @author: xwc1125
package plugins

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chain5j/chain5j-pkg/util/convutil"
	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/store"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
	"github.com/xwc1125/apisix-go/internal/xgateway"
)

var (
	_ plugins.Plugin = new(OAuth2)
)

const (
	AuthMethod_Client_Secret_Basic = "client_secret_basic"
)

func init() {
	p := &OAuth2{
		log:      logger.Log("oauth2"),
		name:     "oauth2",
		version:  "0.1",
		priority: 2515,
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

type OAuth2 struct {
	log logger.Logger
	plugins.DefaultPlugin
	name     string
	version  string
	priority int64

	validator store.Validator
}

type OAuth2Conf struct {
	Disable                         bool   `json:"disable"`
	ClientId                        string `json:"client_id,omitempty" comment:"客户端Id"`
	ClientSecret                    string `json:"client_secret,omitempty" comment:"客户端secret"`
	Discovery                       string `json:"discovery,omitempty" comment:"身份服务器的发现端点的 URL"`
	Real                            string `json:"real,omitempty" comment:"用于认证的领域； 默认为apisix"`
	BearerOnly                      bool   `json:"bearer_only,omitempty" comment:"设置为“true”将检查请求中带有承载令牌的授权标头； 默认为false"`
	LogoutPath                      string `json:"logout_path,omitempty" comment:"默认是/logout"`
	RedirectUri                     string `json:"redirect_uri,omitempty"`
	Timeout                         int    `json:"timeout,omitempty" comment:"默认是 3 秒"`
	SslVerify                       bool   `json:"ssl_verify,omitempty" comment:"默认是 false"`
	IntrospectionEndpoint           string `json:"introspection_endpoint,omitempty" comment:"身份服务器的令牌验证端点的 URL"` // token会直接拼接在连接的后面
	IntrospectionEndpointAuthMethod string `json:"introspection_endpoint_auth_method,omitempty" comment:"令牌自省的认证方法名称,如：client_secret_basic"`
	PublicKey                       string `json:"public_key,omitempty" comment:"验证令牌的公钥"`
	TokenSigningAlgValuesExpected   string `json:"token_signing_alg_values_expected,omitempty" comment:"用于对令牌进行签名的算法"`
}

func (p *OAuth2) Name() string {
	return p.name
}

func (p *OAuth2) Version() string {
	return p.version
}

func (p *OAuth2) Priority() int64 {
	return p.priority
}

func (p *OAuth2) ParseConf(in []byte) (interface{}, error) {
	conf := OAuth2Conf{}
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

func (p *OAuth2) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	config, ok := conf.(OAuth2Conf)
	if !ok {
		p.log.Warn(ErrConfConvert.Error())
		return ErrConfConvert
	}
	if config.Disable {
		return nil
	}

	// 业务处理
	authorization := r.Header.Peek("Authorization")
	if len(authorization) == 0 {
		p.log.Info("request header not find authorization")

		w.SetStatusCode(fasthttp.StatusUnauthorized)
		return fmt.Errorf("not find authorization")
	}
	authToken := string(authorization)
	if config.BearerOnly {
		if !strings.Contains(authToken, "Bearer") {
			p.log.Info("token without Bearer", "authorization", authorization)
			w.SetStatusCode(fasthttp.StatusUnauthorized)
			return fmt.Errorf("error authorization")
		}
	}

	token := strings.Replace(authToken, "Bearer ", "", 1)
	isToken := p.verifyToken(r, config, token)
	if !isToken {
		p.log.Info("Illegal token", "token", authToken)
		w.SetStatusCode(fasthttp.StatusUnauthorized)
		return fmt.Errorf("illegal token")
	}

	return nil
}

type oauthCheckResp struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// verifyToken 验证token的正确性
func (p *OAuth2) verifyToken(r *fasthttp.Request, config OAuth2Conf, token string) bool {
	timeoutDura := time.Duration(config.Timeout) * time.Second
	endpoint := config.IntrospectionEndpoint
	if len(endpoint) == 0 {
		p.log.Error("introspection endpoint is empty")
		return false
	}
	requestURI, err := url.ParseRequestURI(endpoint)
	if err != nil {
		p.log.Error("introspection endpoint err", "err", err)
		return false
	}
	client := fasthttp.HostClient{
		Addr:         requestURI.Host,
		Name:         "",
		IsTLS:        config.SslVerify,
		ReadTimeout:  timeoutDura,
		WriteTimeout: timeoutDura,
	}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)
	if ip := r.Header.Peek(xgateway.HeaderXForwardedFor); ip != nil {
		req.Header.Add(xgateway.HeaderXForwardedFor, string(ip))
	}
	reqUri := fmt.Sprintf("%s%s", endpoint, token)
	p.log.Debug("oauth introspection uri", "uri", reqUri)
	req.SetRequestURI(reqUri)

	switch config.IntrospectionEndpointAuthMethod {
	case AuthMethod_Client_Secret_Basic:
		req.Header.Add("Authorization", "Basic "+basicAuth(config.ClientId, config.ClientSecret))
	default:
		req.Header.Add("Authorization", "Basic "+basicAuth(config.ClientId, config.ClientSecret))
	}

	if config.Timeout <= 0 {
		err = client.Do(req, resp)
	} else {
		err = client.DoTimeout(req, resp, timeoutDura)
	}
	if err != nil {
		p.log.Error("validate token err", "err", err)
		return false
	}
	body := resp.Body()
	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		p.log.Error("unmarshal body to map err", "err", err)
		return false
	}
	if _, ok := bodyMap["error"]; ok {
		p.log.Error("body with err", "resp", string(body))
		return false
	}
	if code, ok := bodyMap["Code"]; !ok || convutil.ToInt(code) != 200 {
		p.log.Error("oauth resp err", "resp", string(body))
		return false
	}
	if data, ok := bodyMap["Data"]; ok {
		switch pData := data.(type) {
		case map[string]interface{}:
			oriBody := r.Body()
			var oriBodyMap map[string]interface{}
			err = json.Unmarshal(oriBody, &oriBodyMap)
			if err != nil {
				p.log.Error("unmarshal req body err", "resp.Data", data)
				return true
			}
			for key, val := range pData {
				oriBodyMap[key] = val
			}
			reqBytes, err := json.Marshal(oriBodyMap)
			if err != nil {
				p.log.Error("marshal data err", "resp.Data", data)
				return true
			}
			r.SetBody(reqBytes)
		default:
			dataBytes, err := json.Marshal(data)
			if err != nil {
				p.log.Error("marshal data err", "resp.Data", data)
				return true
			}
			r.AppendBody(dataBytes)
		}
	}

	return true
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func genCodeChallengeS256(s string) string {
	s256 := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(s256[:])
}

func (p *OAuth2) schema() string {
	return `
 {
  "$comment": "this is a mark for our injected plugin schema",
  "required": [
    "client_id",
    "client_secret",
    "discovery"
  ],
  "properties": {
    "disable": {
      "type": "boolean"
    },
    "client_id": {
      "maxLength": 1024,
      "minLength": 6,
      "type": "string"
    },
    "client_secret": {
      "maxLength": 1024,
      "minLength": 1,
      "type": "string"
    },
    "discovery": {
      "maxLength": 1024,
      "minLength": 1,
      "type": "string"
    },
    "real": {
      "maxLength": 100,
      "minLength": 0,
      "type": "string"
    },
    "bearer_only": {
      "type": "boolean"
    },
    "logout_path": {
      "maxLength": 1024,
      "type": "string"
    },
    "redirect_uri": {
      "maxLength": 1024,
      "type": "string"
    },
    "timeout": {
      "type": "integer"
    },
    "ssl_verify": {
      "type": "boolean"
    },
    "introspection_endpoint": {
      "maxLength": 1024,
      "type": "string"
    },
    "introspection_endpoint_auth_method": {
      "maxLength": 1024,
      "type": "string"
    },
    "public_key": {
      "maxLength": 10000,
      "type": "string"
    },
    "token_signing_alg_values_expected": {
      "maxLength": 10000,
      "type": "string"
    }
  },
  "type": "object"
}
`
}
