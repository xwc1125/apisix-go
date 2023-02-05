package proxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/chain5j/chain5j-pkg/util/convutil"
	"github.com/chain5j/logger"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"github.com/xwc1125/apisix-go/internal/apisix/discover"
	_ "github.com/xwc1125/apisix-go/internal/apisix/discover/polaris"
	"github.com/xwc1125/apisix-go/internal/apisix/lb"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
	_ "github.com/xwc1125/apisix-go/internal/apisix/plugins/plugins"
	_ "github.com/xwc1125/apisix-go/internal/apisix/plugins/plugins/cgw"
)

const (
	_fasthttpHostClientName = "reverse-proxy"
)

// Proxy 反向代理的handler
type Proxy struct {
	log        logger.Logger
	route      entity.Route           // 路由配置
	clients    []*fasthttp.HostClient // clients 客户端集合[做负载均衡时使用]
	clientsUrl []string               // client的URL集合

	lb lb.LoadBalance // lb 负载均衡

	// opt contains finally option to open reverseProxy
	opt              *buildOption
	compressionLevel int
}

// NewProxy create one Proxy with options
func NewProxy(route entity.Route, opts ...Option) (*Proxy, error) {
	log := logger.Log("proxy")
	dst := new(buildOption)
	for _, opt := range opts {
		opt.apply(dst)
	}
	log.Debug("options applied", "dst", dst, "sourceOpts", opts)
	proxy := &Proxy{
		log:        log,
		opt:        dst,
		route:      route,
		clientsUrl: make([]string, 0, 2),
	}
	if !route.EnableWebsocket {
		proxy.clients = make([]*fasthttp.HostClient, 0, 2)
	}

	// 	初始化
	err := proxy.init()
	if err != nil {
		return nil, err
	}
	return proxy, nil
}

// initialize 初始化proxy
func (p *Proxy) init() error {
	upstream := p.route.Upstream
	var (
		cert *tls.Certificate
	)
	if upstream.TLS != nil {
		cert1, err := tls.X509KeyPair([]byte(upstream.TLS.ClientCert), []byte(upstream.TLS.ClientKey))
		if err != nil {
			return err
		}
		cert = &cert1
	}

	if len(upstream.DiscoveryType) > 0 {
		// 启用服务发现
		disco := discover.FindDiscover(upstream.DiscoveryType)
		if disco != nil {
			client, err := disco.GetClient(upstream.DiscoveryArgs)
			if err != nil {
				return err
			}
			if cert != nil {
				client.TLSConfig = &tls.Config{
					Certificates: []tls.Certificate{*cert},
				}
			}
			p.clients = append(p.clients, client)
		}
	}
	nodes := upstream.GetNodes()
	if len(nodes) > 0 {
		// 负载均衡处理
		ws := make([]lb.W, len(nodes))
		p.clientsUrl = make([]string, len(nodes))
		if !p.route.EnableWebsocket {
			p.clients = make([]*fasthttp.HostClient, len(nodes))
		}
		for idx, node := range nodes {
			ws[idx] = lb.Weight(node.Weight)
			p.clientsUrl[idx] = fmt.Sprintf("%s:%d", node.Host, node.Port)
			if !p.route.EnableWebsocket {
				client := &fasthttp.HostClient{
					Addr:                   fmt.Sprintf("%s:%d", node.Host, node.Port),
					Name:                   _fasthttpHostClientName,
					IsTLS:                  upstream.TLS != nil,
					DisablePathNormalizing: p.opt.disablePathNormalizing,
				}
				if cert != nil {
					client.TLSConfig = &tls.Config{
						Certificates: []tls.Certificate{*cert},
					}
				}

				p.clients[idx] = client
			}
		}

		p.lb = lb.NewBalancer(upstream, ws)

		return nil
	}

	return fmt.Errorf("nodes is empty")
}

// GetClient 获取client
func (p *Proxy) GetClient(req *fasthttp.Request) (*fasthttp.HostClient, error) {
	if p.clients == nil || len(p.clients) == 0 {
		p.log.Error("proxy has been closed", "clientLen", len(p.clients))
		return nil, fmt.Errorf("client is empty")
	}

	if p.lb != nil {
		idx := p.lb.Distribute(req)
		return p.clients[idx], nil
	}

	return p.clients[0], nil
}

// GetWs 获取client
func (p *Proxy) GetWs(req *fasthttp.Request) (string, error) {
	if p.clientsUrl == nil || len(p.clientsUrl) == 0 {
		p.log.Error("Proxy has been closed", "clientsUrlLen", len(p.clientsUrl))
		return "", fmt.Errorf("client is empty")
	}

	if p.lb != nil {
		idx := p.lb.Distribute(req)
		return p.clientsUrl[idx], nil
	}

	return p.clientsUrl[0], nil
}

// ServeHTTP 代理服务
func (p *Proxy) ServeHTTP(ctx *fasthttp.RequestCtx) {
	p.log.Info("proxy new request [receive]", "id", getId(ctx), "uniqueKey", convutil.ToString(p.route.ID), "method", string(ctx.Method()), "uri", string(ctx.URI().FullURI()), "remoteIp", ctx.RemoteIP().String())
	if p.route.EnableWebsocket {
		p.serverWs(ctx)
	} else {
		p.serverHttp(ctx)
	}
}

func getId(ctx *fasthttp.RequestCtx) string {
	return fmt.Sprintf("#%016X", ctx.ID())
}

func (p *Proxy) serverHttp(ctx *fasthttp.RequestCtx) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	ctx.Request.CopyTo(req)

	// 设置x-forward-for
	xForwardFor(ctx, req)

	// 【1】预处理阶段
	// 1）通过IP和requestURL获取对应的插件配置
	// 2）读取预处理配置信息
	uniqueKey := convutil.ToString(p.route.ID)
	p.log.Info("proxy prepare conf [start]", "id", getId(ctx), "uniqueKey", uniqueKey, "method", string(req.Header.Method()), "uri", string(req.URI().FullURI()))
	key, err := plugins.PrepareConf(uniqueKey, p.route.Plugins)
	if err != nil {
		p.log.Error("plugin prepare conf err", "err", err)
		p.respToClient(ctx, resp, err)
		return
	}

	// 根据route的lb需求进行初始化和调用
	c, err := p.GetClient(req)
	if err != nil {
		p.log.Error("get client err", "err", err)
		p.respToClient(ctx, resp, err)
		return
	}
	// 先设置目标addr，如果中间插件改写，那么此数据将会变化
	req.SetHost(c.Addr)

	p.log.Info("proxy req call [start]", "id", getId(ctx), "uniqueKey", uniqueKey, "method", string(req.Header.Method()), "uri", string(req.URI().FullURI()))
	// 【2】请求阶段
	// 1）读取配置信息
	// 2）执行请求阶段的插件
	err = plugins.HTTPReqCall(key, req, resp)
	if err != nil {
		p.log.Error("plugin req call err", "err", err)
		p.respToClient(ctx, resp, err)
		return
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		p.respToClient(ctx, resp, nil)
		return
	}

	// 删除部分header
	for _, h := range hopHeaders {
		req.Header.Del(h)
	}

	p.log.Info("proxy req call [end]", "id", getId(ctx), "uniqueKey", uniqueKey, "method", string(req.Header.Method()), "uri", string(req.URI().FullURI()), "tlsConfig", c.TLSConfig, "clientTlsEmpty", c.TLSConfig == nil, "clientIsTLS", c.IsTLS)

	// execute the request and rev response with timeout
	if err := p.doWithTimeout(c, req, resp); err != nil {
		p.log.Error("p.doWithTimeout failed", "err", err, "status", resp.StatusCode())
		resp.SetStatusCode(http.StatusInternalServerError)

		if errors.Is(err, fasthttp.ErrTimeout) {
			resp.SetStatusCode(http.StatusRequestTimeout)
		}
		p.respToClient(ctx, resp, err)
		return
	}
	p.log.Info("proxy resp call [start]", "id", getId(ctx), "uniqueKey", uniqueKey, "method", string(req.Header.Method()), "uri", string(req.URI().FullURI()))

	// 【3】响应阶段
	// 1）读取配置信息
	// 2）执行响应阶段的插件
	err = plugins.HTTPRespCall(key, &ctx.Response)
	if err != nil {
		p.log.Error("proxy resp call err", "err", err)
		p.respToClient(ctx, resp, err)
		return
	}
	// deal with response headers
	p.log.Info("proxy resp call [end]", "id", getId(ctx), "headers", resp.Header.String())
	for _, h := range hopHeaders {
		resp.Header.Del(h)
	}
	p.respToClient(ctx, resp, nil)
	return
}

func (p *Proxy) respToClient(ctx *fasthttp.RequestCtx, resp *fasthttp.Response, err error) {
	if err != nil {
		resp.SetBody([]byte(err.Error()))
		if resp.StatusCode() == 0 || resp.StatusCode() == fasthttp.StatusOK {
			resp.SetStatusCode(fasthttp.StatusInternalServerError)
		}
	}

	compress(p.compressionLevel)(func(ctx *fasthttp.RequestCtx) {})(ctx)
	resp.CopyTo(&ctx.Response)
}

// doWithTimeout calls fasthttp.HostClient Do or DoTimeout, this is depends on p.opt.timeout
func (p *Proxy) doWithTimeout(pc *fasthttp.HostClient, req *fasthttp.Request, res *fasthttp.Response) error {
	if p.opt.timeout <= 0 {
		return pc.Do(req, res)
	}

	return pc.DoTimeout(req, res, p.opt.timeout)
}

// SetClient ...
func (p *Proxy) SetClient(addr string) *Proxy {
	for idx := range p.clients {
		p.clients[idx].Addr = addr
	}
	return p
}

// Reset ...
func (p *Proxy) Reset() {
	for idx := range p.clients {
		p.clients[idx].Addr = ""
	}
}

// Close ... clear and release
func (p *Proxy) Close() {
	p.clients = nil
	p.opt = nil
	// p.bla = nil
	p = nil
}

//
// func copyResponse(src *fasthttp.Response, dst *fasthttp.Response) {
//	src.CopyTo(dst)
//	p.log.Debugf("response header=%v", src.Header)
// }
//
// func copyRequest(src *fasthttp.Request, dst *fasthttp.Request) {
//	src.CopyTo(dst)
// }
//
// func cloneResponse(src *fasthttp.Response) *fasthttp.Response {
//	dst := new(fasthttp.Response)
//	copyResponse(src, dst)
//	return dst
// }
//
// func cloneRequest(src *fasthttp.Request) *fasthttp.Request {
//	dst := new(fasthttp.Request)
//	copyRequest(src, dst)
//	return dst
// }

// Hop-by-hop headers. These are removed when sent to the backend.
// As of RFC 7230, hop-by-hop headers are required to appear in the
// Connection header field. These are the headers defined by the
// obsoleted RFC 2616 (section 13.5.1) and are used for backward
// compatibility.
var hopHeaders = []string{
	"Connection",          // Connection
	"Proxy-Connection",    // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",          // Keep-Alive
	"Proxy-Authenticate",  // Proxy-Authenticate
	"Proxy-Authorization", // Proxy-Authorization
	"Te",                  // canonicalized version of "TE"
	"Trailer",             // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",   // Transfer-Encoding
	"Upgrade",             // Upgrade
}
