// Package proxy_http
//
// @author: xwc1125
package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/chain5j/chain5j-pkg/util/convutil"
	"github.com/chain5j/logger"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	plugins2 "github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	// DefaultUpgrader specifies the parameters for upgrading an HTTP
	// connection to a WebSocket connection.
	DefaultUpgrader = &websocket.FastHTTPUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// DefaultDialer is a dialer with all fields set to the default zero values.
	DefaultDialer = websocket.DefaultDialer
)

func (p *Proxy) serverWs(ctx *fasthttp.RequestCtx) {
	if b := websocket.FastHTTPIsWebSocketUpgrade(ctx); b {
		p.log.Debug("Request is upgraded", "b", b)
	}
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		fmt.Println(string(key), "=", string(value))
	})

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	ctx.Request.CopyTo(req)

	// 设置x-forward-for
	xForwardFor(ctx, req)
	// handle request header
	forwardHeader := builtinForwardHeaderHandler(ctx)
	// 【1】预处理阶段
	// 1）通过IP和requestURL获取对应的插件配置
	// 2）读取预处理配置信息
	uniqueKey := convutil.ToString(p.route.ID)
	token, err := plugins2.PrepareConf(uniqueKey, p.route.Plugins)
	if err != nil {
		logger.Error("plugin prepare conf err", "err", err)
		p.respToClient(ctx, resp, err)
		return
	}
	// 【2】请求阶段
	// 1）读取配置信息
	// 2）执行请求阶段的插件
	err = plugins2.HTTPReqCall(token, req, resp)
	if err != nil {
		logger.Error("plugin http req call err", "err", err)
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

	{
		// 删除需要删除的header
		forwardHeader.Del("Sec-WebSocket-Protocol")
	}
	fmt.Println("=============forwardHeader===============")
	for key, val := range forwardHeader {
		fmt.Println(key, "=", val)
	}

	var (
		dialer   = DefaultDialer
		upgrader = DefaultUpgrader
	)
	// Connect to the backend URL, also pass the headers we get from the request
	// together with the Forwarded headers we prepared above.
	// TODO: support multiplexing on the same backend connection instead of
	// opening a new TCP connection time for each request. This should be
	// optional:
	// http://tools.ietf.org/html/draft-ietf-hybi-websocket-multiplexing-01
	var scheme = "ws"
	wsUrl, err := p.GetWs(req)
	if err != nil {
		p.log.Error("get client url err", "err", err)
		p.respToClient(ctx, resp, err)
		return
	}
	targetUrl := fmt.Sprintf("%s://%s", scheme, wsUrl)
	// dst的conn
	connBackend, respBackend, err := dialer.Dial(targetUrl+p.route.URI, forwardHeader)
	if err != nil {
		p.log.Error("websocket proxy: couldn't dial to remote backend", "err", err, "host", targetUrl)

		// logger.Debugf("resp_backent =%v", respBackend)
		if respBackend != nil {
			if err = wsCopyResponse(resp, respBackend); err != nil {
				p.log.Error("could not finish wsCopyResponse", "err", err)
			}
		} else {
			// ctx.SetStatusCode(http.StatusServiceUnavailable)
			// ctx.WriteString(http.StatusText(http.StatusServiceUnavailable))
			ctx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
		}
		return
	}
	// upgrader做了origin的校验，如果删除就不会校验了
	ctx.Request.Header.Del("origin")
	// 需要适配：Sec-WebSocket-Protocol
	if prot := ctx.Request.Header.Peek("Sec-WebSocket-Protocol"); len(prot) > 0 {
		prols := string(prot)
		split := strings.Split(prols, ",")
		upgrader.Subprotocols = append(upgrader.Subprotocols, strings.TrimSpace(split[0]))
	}
	// forwardHeader.Del("origin")
	// Now upgrade the existing incoming request to a WebSocket connection.
	// Also pass the header that we gathered from the Dial handshake.
	err = upgrader.Upgrade(ctx, func(connPub *websocket.Conn) {
		defer connPub.Close()
		var (
			errClient  = make(chan error, 1)
			errBackend = make(chan error, 1)
			message    string
		)

		p.log.Debug("upgrade handler working")
		go replicateWebsocketConn(p.log, connPub, connBackend, errClient)  // response
		go replicateWebsocketConn(p.log, connBackend, connPub, errBackend) // request

		for {
			select {
			case err = <-errClient:
				message = "websocketproxy: Error when copying response"
			case err = <-errBackend:
				message = "websocketproxy: Error when copying request"
			}

			// log error except '*websocket.CloseError'
			if _, ok := err.(*websocket.CloseError); !ok {
				p.log.Error(message, "err", err)
			}
		}
	})

	if err != nil {
		p.log.Error("websocket proxy: couldn't upgrade", "err", err)
		return
	}
	return
}

// builtinForwardHeaderHandler built in handler for dealing forward request headers.
func builtinForwardHeaderHandler(ctx *fasthttp.RequestCtx) (forwardHeader http.Header) {
	forwardHeader = make(http.Header, 4)
	// ctx.Request.Header.VisitAll(func(key, value []byte) {
	// 	forwardHeader.Add(string(key), string(value))
	// })
	// Pass headers from the incoming request to the dialer to forward them to
	// the final destinations.
	if origin := ctx.Request.Header.Peek("Origin"); string(origin) != "" {
		forwardHeader.Add("Origin", string(origin))
	}

	if prot := ctx.Request.Header.Peek("Sec-WebSocket-Protocol"); string(prot) != "" {
		forwardHeader.Add("Sec-WebSocket-Protocol", string(prot))
	}

	if cookie := ctx.Request.Header.Peek("Cookie"); string(cookie) != "" {
		forwardHeader.Add("Cookie", string(cookie))
	}

	// var wsDefaultKey = []string{
	// 	"Upgrade",
	// 	"Connection",
	// 	"Sec-WebSocket-Key",
	// 	"Sec-WebSocket-Version",
	// 	"Sec-Websocket-Extensions",
	// 	"Sec-Websocket-Protocol",
	// }
	// for _, key := range wsDefaultKey {
	// 	if cookie := ctx.Request.Header.Peek(key); string(cookie) != "" {
	// 		forwardHeader.Del(key)
	// 	}
	// }

	if string(ctx.Request.Host()) != "" {
		forwardHeader.Set("Host", string(ctx.Request.Host()))
	}

	// Pass X-Forwarded-For headers too, code below is a part of
	// httputil.ReverseProxy. See http://en.wikipedia.org/wiki/X-Forwarded-For
	// for more information
	// TODO: use RFC7239 http://tools.ietf.org/html/rfc7239
	if clientIP, _, err := net.SplitHostPort(ctx.RemoteAddr().String()); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior := ctx.Request.Header.Peek("X-Forwarded-For"); string(prior) != "" {
			clientIP = string(prior) + ", " + clientIP
		}
		forwardHeader.Set("X-Forwarded-For", clientIP)
	}

	// Set the originating protocol of the incoming HTTP request. The SSL might
	// be terminated on our site and because we doing proxy adding this would
	// be helpful for applications on the backend.
	forwardHeader.Set("X-Forwarded-Proto", "http")
	if ctx.IsTLS() {
		forwardHeader.Set("X-Forwarded-Proto", "https")
	}

	return
}

// replicateWebsocketConn to
// copy message from src to dst
func replicateWebsocketConn(log logger.Logger, dst, src *websocket.Conn, errChan chan error) {
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			// true: handle websocket close error
			log.Debug("src.ReadMessage failed", "srcIp", src.RemoteAddr().String(), "dstIp", dst.RemoteAddr().String(), "msgType", msgType, "msg", msg, "err", err)
			if ce, ok := err.(*websocket.CloseError); ok {
				msg = websocket.FormatCloseMessage(ce.Code, ce.Text)
			} else {
				log.Error("src.ReadMessage failed", "err", err)
				msg = websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, err.Error())
			}

			errChan <- err
			if err = dst.WriteMessage(websocket.CloseMessage, msg); err != nil {
				log.Error("write close message failed", "err", err)
			}
			break
		}

		err = dst.WriteMessage(msgType, msg)
		if err != nil {
			log.Error("dst.WriteMessage failed", "err", err)
			errChan <- err
			break
		}
	}
}

// wsCopyResponse .
// to help copy origin websocket response to client
func wsCopyResponse(dst *fasthttp.Response, src *http.Response) error {
	for k, vv := range src.Header {
		for _, v := range vv {
			dst.Header.Add(k, v)
		}
	}

	dst.SetStatusCode(src.StatusCode)
	defer src.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, src.Body); err != nil {
		return err
	}
	dst.SetBody(buf.Bytes())
	return nil
}
