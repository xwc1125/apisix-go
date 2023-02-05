// Package fasthttp
//
// @author: xwc1125
package fasthttp

import (
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/xgateway"
)

var (
	_ xgateway.Header = new(RequestHeader)
	_ xgateway.Header = new(ResponseHeader)
)

type RequestHeader struct {
	*fasthttp.RequestHeader
	statusCode int
}

func NewRequestHeader(requestHeader *fasthttp.RequestHeader) *RequestHeader {
	return &RequestHeader{
		RequestHeader: requestHeader,
	}
}

func (h *RequestHeader) Get(key string) string {
	return string(h.Peek(key))
}

func (h *RequestHeader) StatusCode() int {
	if h.statusCode == 0 {
		return fasthttp.StatusOK
	}
	return h.statusCode
}

func (h *RequestHeader) Headers() []byte {
	return h.RawHeaders()
}

func (h *RequestHeader) AllCookie(f func(key []byte, value []byte)) {
	h.RequestHeader.VisitAllCookie(f)
}

func (h *RequestHeader) SetCookie(cookie xgateway.Cookie) {

}

func (h *RequestHeader) AddCookie(key, value string) {
	h.RequestHeader.SetCookie(key, value)
}

func (h *RequestHeader) SetStatusCode(statusCode int) {
}

type ResponseHeader struct {
	*fasthttp.ResponseHeader
	cookie     *Cookie
	host       []byte
	userAgent  []byte
	method     []byte
	requestURI []byte
	protocol   []byte
}

func NewResponseHeader(respHeader *fasthttp.ResponseHeader) *ResponseHeader {
	return &ResponseHeader{
		ResponseHeader: respHeader,
	}
}

func (h *ResponseHeader) Get(key string) string {
	return string(h.Peek(key))
}

func (h *ResponseHeader) Host() []byte {
	return h.host
}

func (h *ResponseHeader) SetHost(host string) {
	h.host = append(h.host[:0], host...)
}

func (h *ResponseHeader) UserAgent() []byte {
	return h.userAgent
}

func (h *ResponseHeader) SetUserAgent(userAgent string) {
	h.userAgent = append(h.userAgent[:0], userAgent...)
}

func (h *ResponseHeader) Method() []byte {
	if len(h.method) == 0 {
		return []byte(fasthttp.MethodGet)
	}
	return h.method
}

func (h *ResponseHeader) SetMethod(method string) {
	h.method = append(h.method[:0], method...)
}

func (h *ResponseHeader) SetProtocol(method string) {
}

func (h *ResponseHeader) RequestURI() []byte {
	requestURI := h.requestURI
	if len(requestURI) == 0 {
		requestURI = []byte("/")
	}
	return requestURI
}

func (h *ResponseHeader) SetRequestURI(requestURI string) {
	h.requestURI = append(h.requestURI[:0], requestURI...)
}

func (h *ResponseHeader) SetCookie(cookie xgateway.Cookie) {
	h.cookie = cookie.(*Cookie)
	h.ResponseHeader.SetCookie(h.cookie.Cookie)
}

func (h *ResponseHeader) Cookie(key string) []byte {
	return h.ResponseHeader.PeekCookie(key)
}

func (h *ResponseHeader) AllCookie(f func(key []byte, value []byte)) {
	h.ResponseHeader.VisitAllCookie(f)
}

func (h *ResponseHeader) AddCookie(key, value string) {
	h.cookie.SetKey(key)
	h.cookie.SetValue(value)
}

func (h *ResponseHeader) Headers() []byte {
	return h.ResponseHeader.Header()
}
